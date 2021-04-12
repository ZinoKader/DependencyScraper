package data

import (
	"os"
	"testing"
)

func TestFileIncrement(t *testing.T) {
	os.Create("test.txt")

	WriteToFile("test.txt", []byte("just some text"))

	if _, err := os.Stat("test0.txt"); err != nil {
		t.Errorf("Test Failed: Did not create correct file")
	}
	os.Remove("test.txt")
	os.Remove("test0.txt")
}
func TestFileIncrementNoExtension(t *testing.T) {
	os.Create("test")

	WriteToFile("test", []byte("just some text"))

	if _, err := os.Stat("test0"); err != nil {
		t.Errorf("Test Failed: Did not create correct file")
	}
	os.Remove("test")
	os.Remove("test0")
}

func TestMultipleFileIncrement(t *testing.T) {
	os.Create("test.txt")

	WriteToFile("test.txt", []byte("just some text"))
	WriteToFile("test.txt", []byte("just some text"))

	if _, err := os.Stat("test0.txt"); err != nil {
		t.Errorf("Test Failed: Did not create first incremented file")
	}
	if _, err := os.Stat("test1.txt"); err != nil {
		t.Errorf("Test Failed: Did not create second incremented file")
	}
}
