package main

import (
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/ZinoKader/KEX/model"
	"github.com/ZinoKader/KEX/pkg/data"
	"github.com/ZinoKader/KEX/pkg/scraping"
)

var SLICES = runtime.NumCPU()

func mapPackageFiles(repos []model.RepositoryFileRow, accumulator <-chan []model.DependencyTree) {
	partSize := len(repos) / SLICES
	var wg sync.WaitGroup
	for i := 0; i < len(repos); i += partSize {
		var reposPart []model.RepositoryFileRow
		if i+partSize > len(repos) {
			reposPart = repos[i:]
			i = len(repos) // break out of next loop
		} else {
			reposPart = repos[i : i+partSize]
		}

		wg.Add(1)

		go func(rows []model.RepositoryFileRow, task *sync.WaitGroup) {
			for _, row := range rows {
				URLParts := strings.Split(row.URL, "/")
				ownerName := URLParts[len(URLParts)-2]
				repoName := URLParts[len(URLParts)-1]
				scraping.ExtractRepoFileTree(ownerName, repoName)
			}
			task.Done()
		}(reposPart, &wg)
	}
}

func main() {
	// ensure all CPUs used
	runtime.GOMAXPROCS(runtime.NumCPU())

	repoRows := data.GetRepositoryFileRows()
	packageFileAccumulator := make(chan []model.DependencyTree)

	go mapPackageFiles(repoRows, packageFileAccumulator)

	fmt.Println(repoRows)
}
