package data

import (
	"strings"

	mapset "github.com/deckarep/golang-set"
	"github.com/tidwall/gjson"
)

func ParsePackage(packageData []byte) (mapset.Set, error) {
	dependencyCollector := mapset.NewSet()

	results := gjson.GetManyBytes(packageData, "devDependencies", "dependencies")

	for _, result := range results {
		result.ForEach(func(key, value gjson.Result) bool {
			formatted := strings.ReplaceAll(strings.ReplaceAll(key.String(), "@",
				"%40"), "/", "%2f")
			dependencyCollector.Add(formatted)
			return true
		})
	}

	return dependencyCollector, nil
}
