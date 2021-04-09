package scraping

import (
	"testing"

	"github.com/patrickmn/go-cache"
)

var DEPENDENCY_CACHE = cache.New(cache.NoExpiration, 0)
var TEST_DEPENDENCIES = []string{"react", "core-js", "%40abios%2fsmorgasbord-themes"}

func TestResultNotEmpty(t *testing.T) {
	res := RepoDependencies(TEST_DEPENDENCIES, DEPENDENCY_CACHE)

	if len(res) == 0 {
		t.Error("Test Failed: Result is empty, expected 2 repos")
	}
}

func TestCorrectResult(t *testing.T) {
	res := RepoDependencies(TEST_DEPENDENCIES, DEPENDENCY_CACHE)

	if res[0] != "https://api.github.com/facebook/react" || res[1] != "https://api.github.com/zloirock/core-js" {
		t.Errorf("Test Failed: Result is not correct. Expected  [https://api.github.com/facebook/react, https://api.github.com/zloirock/core-js] got %v", res)
	}
}
