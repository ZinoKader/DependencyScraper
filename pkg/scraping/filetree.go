package scraping

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type GithubFileTree struct {
	Paths []string `json:"paths"`
}

func ExtractRepoFileTree(ownerName string, repoName string) []string {
	// visit main repo page and extract main branch name and file finder URL
	ghURL := strings.Join([]string{"https://github.com", ownerName, repoName}, "/")
	res, err := http.Get(ghURL)
	if err != nil {
		fmt.Printf("Could not fetch github page of %v\n %v", ghURL, err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	repoFileFinderURL, _ := doc.Find("a").FilterFunction(func(i int, s *goquery.Selection) bool {
		href := s.AttrOr("href", "")
		return strings.Contains(href, "/find/")
	}).First().Attr("href")

	repoFileFinderURLParts := strings.Split(repoFileFinderURL, "/")
	repoMainBranchName := repoFileFinderURLParts[len(repoFileFinderURLParts)-1]
	
	// visit file finder page for repo and find the URL for the filetree
	repoFileFinderURL = fmt.Sprintf("https://github.com%s",repoFileFinderURL)
	res, err = http.Get(repoFileFinderURL)
	if err != nil {
		fmt.Printf("Could not fetch file finder page of %v\n %v", repoFileFinderURL, err)
	}
	defer res.Body.Close()

	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal("Could not init goquery for file finder page", err)
	}

	repoFileTreeURL, _ := doc.Find("fuzzy-list").First().Attr("data-url")
	repoFileTreeURL = fmt.Sprintf("https://github.com%s",repoFileTreeURL)

	// visit the file tree for the repo and extract the paths to the package.json files
	var packageFileURLs []string
	req, err := http.NewRequest("GET", repoFileTreeURL, nil)
	if err != nil {
		log.Fatal("Could not create new request for file tree page", err)
	}
	// this header is needed to trick GitHub into thinking the request was made from the client
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Could not fetch file tree page for %v\n %v", repoFileTreeURL, err)
	}
	defer res.Body.Close()

	fileTreeBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}
	var fileTree = new(GithubFileTree)
	json.Unmarshal(fileTreeBody, &fileTree)
	for _, path := range fileTree.Paths {
		if strings.Contains(path, "package.json") && !strings.Contains(path, "node_modules") {
			packageFileURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",ownerName,repoName,repoMainBranchName,path)
			packageFileURLs = append(packageFileURLs, packageFileURL)
		}
	}
	return packageFileURLs
}
