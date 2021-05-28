package data

import (
	"strings"

	mapset "github.com/deckarep/golang-set"
	"github.com/tidwall/gjson"
)

func ParsePackage(packageData []byte) (mapset.Set, error) {
	dependencyCollector := mapset.NewSet()
<<<<<<< HEAD

	results := gjson.GetManyBytes(packageData, "devDependencies", "dependencies")

	for _, result := range results {
		result.ForEach(func(key, value gjson.Result) bool {
			formatted := strings.ReplaceAll(strings.ReplaceAll(key.String(), "@",
				"%40"), "/", "%2f")
			dependencyCollector.Add(formatted)
			return true
		})
	}

=======
	
	results := gjson.GetManyBytes(packageData, "devDependencies", "dependencies")
	
	for _, result := range results {
		result.ForEach(func(key, value gjson.Result) bool {
			formatted := strings.ReplaceAll(strings.ReplaceAll(key.String(), "@", "%40"), "/", "%2f")
			dependencyCollector.Add(formatted)
			return true
		}) 
	} 	
	
>>>>>>> eb81f59490540055815871b64626e1d74e1131dc
	return dependencyCollector, nil
}
