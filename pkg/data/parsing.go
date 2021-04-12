package data

import (
	"encoding/json"
	"strings"

	mapset "github.com/deckarep/golang-set"
)

func ParsePackage(packageData []byte) (mapset.Set, error) {
	dependencyCollector := mapset.NewSet()

	var v interface{}
	err := json.Unmarshal(packageData, &v)
	if err != nil {
		return dependencyCollector, err
	}
	data := v.(map[string]interface{})

	devDependencies, found := data["devDependencies"]
	if found {
		for dependency := range devDependencies.(map[string]interface{}) {
			// replace @  and / with hexadecimal counterparts
			formatted := strings.ReplaceAll(strings.ReplaceAll(dependency, "@", "%40"), "/", "%2f")
			dependencyCollector.Add(formatted)
		}
	}

	dependencies, found := data["dependencies"]
	if found {
		for dependency := range dependencies.(map[string]interface{}) {
			formatted := strings.ReplaceAll(strings.ReplaceAll(dependency, "@", "%40"), "/", "%2f")
			dependencyCollector.Add(formatted)
		}
	}

	return dependencyCollector, nil
}
