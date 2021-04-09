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

var dependencyCache = cache.New(cache.NoExpiration, 0)

func repoDependencies (dependencies []string) []string {

	
	var repoUrls = []string{}
	for _ ,dependency := range dependencies {
		
		url := fmt.Sprintf("https://api.npms.io/v2/package/%s",dependency)

		repoUrl, found := dependencyCache.Get(url)

		if found {
			repoUrls = append(repoUrls, repoUrl.(string))	
			continue
		} 
		
		response, err := http.Get(url)

		if err != nil {
			log.Println(err)
			continue	
		}
		defer response.Body.Close()

		bodyBytes,err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Println(err)
			continue
		}

		results := gjson.GetManyBytes(bodyBytes, "collected.metadata.links.repository","collected.metadata.repository.url")

		var result string
		if len(results[0].String()) > 0 {
			result = strings.Replace(results[0].String(), "github", "api.github", 1)

		} else if len(results[1].String()) > 0 && strings.Contains(results[1].String(),"https") {
			result = strings.TrimSuffix(strings.Replace(results[1].String(), "git+https://github","https://api.github", 1), ".git")

		} else {
			continue
		}

		repoUrls = append(repoUrls, result)	
		dependencyCache.Add(dependency, repoUrl, cache.NoExpiration)
	}
	return repoUrls
}

