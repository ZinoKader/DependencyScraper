package scraping

import (
	"fmt"
	"net/http"
	"strings"
)

func ExtractRepoFileTree(ownerName string, repoName string) {
	ghURL := strings.Join([]string{"https://github.com", ownerName, repoName}, "/")
	res, err := http.Get(ghURL)
	if err != nil {
		fmt.Printf("Could not fetch github page of %v\n %v", ghURL, err)
	}
}
