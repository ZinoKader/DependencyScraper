package data

import (
	"io/ioutil"
	"testing"
)


var testPackage, _ = ioutil.ReadFile("./testpackage.json")

func TestErrors(t *testing.T) {
	_, err := ParsePackage(string(testPackage))

	if err != nil {
		t.Errorf("Test failed: raised error: %s", err)
	}
}

func TestResultNotEmpty(t *testing.T) {
	deps, _ := ParsePackage(string(testPackage))
	
	if len(deps) == 0 {
		t.Error("Test Failed: Result is empty")
	}
}

func TestNoDuplicates(t *testing.T) {
	
	// modified json file by anding "dependencies"
	deps, _ := ParsePackage(string(testPackage))

	if len(deps) != 41 {
		t.Errorf("Test Failed: Duplicate dependencies, expected 41 dependenices got %d",len(deps))
	}
}
