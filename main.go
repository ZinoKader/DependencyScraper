package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/ZinoKader/KEX/model"
	"github.com/ZinoKader/KEX/pkg/data"
	"github.com/ZinoKader/KEX/pkg/scraping"
)

var SLICES = runtime.NumCPU()

func mapPackageFiles(repos []model.RepositoryFileRow, accumulator chan<- model.DependencyTree) {
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
		go func() {
			for _, row := range reposPart {
				URLParts := strings.Split(row.URL, "/")
				ownerName := URLParts[len(URLParts)-2]
				repoName := row.Name
				dependencyTree, err := scraping.RepoDependencyTree(ownerName, repoName)
				dependencyTree.ID = row.ID
				if err != nil {
					fmt.Printf("Something went wrong when scraping the dependency tree for repo %s \n%v", row.URL, err)
					continue
				}
				// push parsed dependency tree to accumulator
				accumulator <- dependencyTree
			}
			wg.Done()
		}()
	}

	wg.Wait()
	close(accumulator)
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: ", os.Args[0], "{input file path} [csv]", "{output file path} [csv]")
		return
	}

	inputPath := os.Args[1]
	//outputPath := os.Args[2]

	// command line arguments to function
	// ensure all CPUs used
	runtime.GOMAXPROCS(runtime.NumCPU())

	repoRows := data.RepositoryFileRows(inputPath)
	packageFileAccumulator := make(chan model.DependencyTree)

	var wg sync.WaitGroup

	// map repo urls to dependency contents (dependencies and devDependencies)
	go mapPackageFiles(repoRows, packageFileAccumulator)

	wg.Add(1)
	go func() {
		for {
			v, ok := <-packageFileAccumulator
			if !ok {
				break
			}
			fmt.Printf("bla: %v\n", v)
		}
		wg.Done()
	}()
	// map dependencies and devDependencies to trees of

	// reduce/accumulate results into file (reduce to csv where id is parent repo ID and one row for every dependency where the dependency is the GitHub URL)

	wg.Wait()
}
