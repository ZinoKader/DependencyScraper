package data

import (
	"encoding/json"
	"strings"
)


func ParsePackage (packageJson string) (map[string]bool, error) {
	dependencyCollector := make(map[string]bool)
	packageData :=  []byte(packageJson)
	
	var v interface{}
	err := json.Unmarshal(packageData, &v)	
	if err != nil {
		return dependencyCollector ,err		
	}	
	data := v.(map[string]interface{})

	
	devDependencies, found := data["devDependencies"]
	if found {
		for dependecy := range devDependencies.(map[string]interface{}) {
			// replace @  and / with hexadecimal counterparts
			formated := strings.ReplaceAll(strings.ReplaceAll(dependecy, "@", "%40"), "/", "%2f")
			dependencyCollector[formated] = true
		}	
	}
	dependencies, found := data["dependencies"]
	if found {
		for dependecy := range dependencies.(map[string]interface{}) {
			formated := strings.ReplaceAll(strings.ReplaceAll(dependecy, "@", "%40"), "/", "%2f")
			dependencyCollector[formated] = true
		}
	}
	return dependencyCollector, nil
}
