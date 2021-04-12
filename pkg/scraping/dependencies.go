package scraping

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/patrickmn/go-cache"
	"github.com/tidwall/gjson"
)

func RepoDependencies(dependencies []string, dependencyCache *cache.Cache) []string {

	var repoURLs = []string{}
	for _, dependency := range dependencies {

		url := fmt.Sprintf("https://api.npms.io/v2/package/%s", dependency)

		cachedURL, found := dependencyCache.Get(url)

		if found {
			repoURLs = append(repoURLs, cachedURL.(string))
			fmt.Println(cachedURL.(string))
			continue
		}

		response, err := http.Get(url)

		if err != nil {
			log.Println(err)
			continue
		}
		defer response.Body.Close()

		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println(err)
			continue
		}

		results := gjson.GetManyBytes(bodyBytes, "collected.metadata.links.repository", "collected.metadata.repository.url")

		var repoURL string
		if len(results[0].String()) > 0 {
			repoURL = strings.Replace(results[0].String(), "github", "api.github", 1)
		} else if len(results[1].String()) > 0 && strings.Contains(results[1].String(), "https") {
			repoURL = strings.TrimSuffix(strings.Replace(results[1].String(), "git+https://github", "https://api.github", 1), ".git")
		} else {
			continue
		}

		repoURLs = append(repoURLs, repoURL)
		dependencyCache.Add(dependency, repoURL, cache.NoExpiration)
	}
	return repoURLs
}
