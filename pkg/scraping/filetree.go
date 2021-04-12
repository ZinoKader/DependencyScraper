package scraping

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ZinoKader/KEX/model"
	"github.com/ZinoKader/KEX/pkg/data"
	mapset "github.com/deckarep/golang-set"
)

type GithubFileTree struct {
	Paths []string `json:"paths"`
}

func RepoDependencyTree(ownerName string, repoName string) (model.DependencyTree, error) {

	// try setting proxy
	proxyUrl, _ := url.Parse(fmt.Sprintf("http://%s", RandomProxy()))
	httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	// visit main repo page and extract main branch name and file finder URL
	ghURL := strings.Join([]string{"https://github.com", ownerName, repoName}, "/")

	req, err := http.NewRequest("GET", ghURL, nil)
	if err != nil {
		return model.DependencyTree{}, err
	}
	req.Close = true

	res1, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("Could not fetch github page of %v\n%v", ghURL, err)
		return model.DependencyTree{}, err
	}
	defer res1.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res1.Body)
	if err != nil {
		fmt.Printf("Could not parse HTML of %v\n %v", ghURL, err)
		return model.DependencyTree{}, err
	}

	repoFileFinderURL, _ := doc.Find("a").FilterFunction(func(i int, s *goquery.Selection) bool {
		href := s.AttrOr("href", "")
		return strings.Contains(href, "/find/")
	}).First().Attr("href")

	// We did not end up on a functioning page for the file finder, this repo is not available
	if len(repoFileFinderURL) == 0 {
		return model.DependencyTree{}, &model.RepoNotExist{
			RepositoryURL: ghURL,
		}
	}

	repoFileFinderURLParts := strings.Split(repoFileFinderURL, "/")
	repoMainBranchName := repoFileFinderURLParts[len(repoFileFinderURLParts)-1]

	// visit file finder page for repo and find the URL for the filetree
	repoFileFinderURL = fmt.Sprintf("https://github.com%s", repoFileFinderURL)
	req, err = http.NewRequest("GET", repoFileFinderURL, nil)
		if err != nil {
			return model.DependencyTree{}, err
		}
	req.Close = true
	res2, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("Could not fetch file finder page of %v\n%v", repoFileFinderURL, err)
		return model.DependencyTree{}, err
	}
	defer res2.Body.Close()

	doc, err = goquery.NewDocumentFromReader(res2.Body)
	if err != nil {
		fmt.Println("Could not init goquery for file finder page", err)
		return model.DependencyTree{}, err
	}

	repoFileTreeURL, _ := doc.Find("fuzzy-list").First().Attr("data-url")
	repoFileTreeURL = fmt.Sprintf("https://github.com%s", repoFileTreeURL)

	// visit the file tree for the repo and extract the paths to the package.json files
	req, err = http.NewRequest("GET", repoFileTreeURL, nil)
	req.Close = true
	if err != nil {
		fmt.Println("Could not create new request for file tree page", err)
		return model.DependencyTree{}, err
	}
	// this header is needed to trick GitHub into thinking the request was made from the client
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	res3, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("Could not fetch file tree page for %v\n%v", repoFileTreeURL, err)
		return model.DependencyTree{}, err
	}
	defer res3.Body.Close()

	fileTreeBody, err := ioutil.ReadAll(res3.Body)
	if err != nil {
		fmt.Printf("Could not read body of tree page for %v\n%v", repoFileTreeURL, err)
		return model.DependencyTree{}, err
	}

	var dependencyTree = new(model.DependencyTree)
	var fileTree = new(GithubFileTree)
	json.Unmarshal(fileTreeBody, &fileTree)
	dependencyCollector := mapset.NewSet()

	// fetch and read raw package.json files
	for _, path := range fileTree.Paths {
		if strings.Contains(path, "package.json") && !strings.Contains(path, "node_modules") {
			packageFileURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", ownerName, repoName, repoMainBranchName, path)
			req, err := http.NewRequest("GET", packageFileURL, nil)
			if err != nil {
						return model.DependencyTree{}, err
			}
			req.Close = true
			res, err := httpClient.Do(req)
			if err != nil {
				fmt.Printf("Could not fetch raw dependency for %v\n%v", packageFileURL, err)
				continue // do our best effort, if this particular file did not work, hope the others do
			}
			defer res.Body.Close()

			packageFileBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				fmt.Printf("Could not read body of package.json for %v\n%v", packageFileURL, err)
				continue
			}
			dependencies, err := data.ParsePackage(packageFileBody)
			if err != nil {
				fmt.Printf("Could not parse package.json for %v\n%v", packageFileURL, err)
				continue
			}
			dependencyCollector = dependencyCollector.Union(dependencies)
		}
	}

	for _, dependency := range dependencyCollector.ToSlice() {
		dependencyTree.Dependencies = append(dependencyTree.Dependencies, dependency.(string))
	}

	// no package.json file, not a repo we can include in the dependency graph
	if len(dependencyTree.Dependencies) == 0 {
		return model.DependencyTree{}, &model.RepoNoPackage{
			RepositoryURL: ghURL,
		}
	}

	return *dependencyTree, nil
}
