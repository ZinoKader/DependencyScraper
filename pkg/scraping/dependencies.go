package scraping

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/patrickmn/go-cache"
	"github.com/tidwall/gjson"
)

func RepoDependencies(dependencies []string, dependencyCache *cache.Cache) []string {

	var repoURLs = []string{}
	for _, dependency := range dependencies {


		cachedURL, found := dependencyCache.Get(dependency)

		if found {
			repoURLs = append(repoURLs, cachedURL.(string))
			continue
		}

		URL := fmt.Sprintf("https://api.npms.io/v2/package/%s", dependency)
		req, err := CreateRequest(URL)
		if err != nil {
			log.Println(err)
			continue
		}
		response, err := httpClient.Do(req)

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
		
		repoURL = strings.ReplaceAll(repoURL, "www.", "")
		repoURLs = append(repoURLs, repoURL)
		dependencyCache.Add(dependency, repoURL, cache.NoExpiration)
	}
	return repoURLs
}
