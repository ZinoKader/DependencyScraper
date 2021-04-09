package data

import (
	"strings"
	"testing"

	"github.com/ZinoKader/KEX/pkg/scraping"
)

func TestSmall(t *testing.T) {
	filerows := GetRepositoryFileRows()

	for _, row := range filerows {
		split := strings.Split(row.URL, "/")
		owner := split[len(split) - 2]
		scraping.ExtractRepoFileTree(owner, row.Name)
	}
}
