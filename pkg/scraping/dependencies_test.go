package scraping

import (
	"testing"
)

var TEST_DEPENDENCIES = []string{"react", "core-js", "%40abios%2fsmorgasbord-themes"}

func TestResultNotEmpty(t *testing.T) {
	res := repoDependencies(TEST_DEPENDENCIES)

	if len(res) == 0 {
		t.Error("Test Failed: Result is empty, expected 2 repos")
	}
}

func TestCorrectResult(t *testing.T) {
	res := repoDependencies(TEST_DEPENDENCIES)

	if res[0] != "https://api.github.com/facebook/react" || res[1] != "https://api.github.com/zloirock/core-js" {
		t.Errorf("Test Failed: Result is not correct. Expected  [https://api.github.com/facebook/react, https://api.github.com/zloirock/core-js] got %v", res)
	}
}
