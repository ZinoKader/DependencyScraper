package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ZinoKader/KEX/model"
	"github.com/ZinoKader/KEX/pkg/data"
	"github.com/ZinoKader/KEX/pkg/scraping"
	"github.com/patrickmn/go-cache"
)

var SLICES = runtime.NumCPU()

func mapPackageFiles(repos []model.RepositoryFileRow, edgeAccumulator chan  <- model.PackageEdges, dependencyCache *cache.Cache) {
	var edgeWg sync.WaitGroup
	partSize := len(repos) / SLICES
	for i := 0; i < len(repos); i += partSize {
		var treeWg sync.WaitGroup
		var reposPart []model.RepositoryFileRow
		if i+partSize > len(repos) {
			reposPart = repos[i:]
			i = len(repos) // break out of next loop
		} else {
			reposPart = repos[i : i+partSize]
		}

		treeAccumulator := make(chan model.DependencyTree)

		treeWg.Add(1)
		go func() {
			for _, row := range reposPart {
				URLParts := strings.Split(row.URL, "/")
				ownerName := URLParts[len(URLParts)-2]
				repoName := row.Name
				dependencyTree, err := scraping.RepoDependencyTree(ownerName, repoName)
				dependencyTree.ID = row.ID
				if err != nil {
					fmt.Printf("Something went wrong when scraping the dependency tree for repo %s\n%v\n", row.URL, err)
					switch err.(type) {
					case *model.ConnectionError:
						// handle marking this repo as "retry"
						// connectionError := err.(*model.ConnectionError)
					case *model.RepoNoPackage:
						// handle marking this repo as "do-not-retry"
						// noPackageError := err.(*model.RepoNoPackage)
					case *model.RepoNotExist:
						// handle marking this repo as "do-not-retry"
						// notExistError := err.(*model.RepoNotExist)
					default:
					}
					continue
				}
				// push parsed dependency tree to accumulator
				treeAccumulator <- dependencyTree
			}
			treeWg.Done()
		}()
	
		edgeWg.Add(1)
		go func() {

			for tree := range treeAccumulator {
				fmt.Println(dependencyCache.Items())
				dependencyURLs := scraping.RepoDependencies(tree.Dependencies, dependencyCache) 				

				edges := model.PackageEdges{
					ID: 							tree.ID, 
					DependencyURLs: 	dependencyURLs,
				}	

				edgeAccumulator <- edges
			}
			edgeWg.Done()
		}()

		treeWg.Wait()
		close(treeAccumulator)
	}
	edgeWg.Wait()
	close(edgeAccumulator)
}

func setup() {
	// init psuedorandom seed with nano time
	rand.Seed(time.Now().UnixNano())
	// ensure all CPUs are used
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	// read input file and output file
	if len(os.Args) != 3 {
		fmt.Println("Usage: ", os.Args[0], "{input file path} [csv]", "{output file path} [csv]")
		return
	}
	inputPath := os.Args[1]
	//outputPath := os.Args[2]
	repoRows := data.RepositoryFileRows(inputPath)

	dependencyCache := cache.New(cache.NoExpiration, 0)
	edgeAccumulator := make(chan model.PackageEdges)

	setup()

	// map repo urls to dependency contents (dependencies and devDependencies)
	go mapPackageFiles(repoRows, edgeAccumulator, dependencyCache)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			v, ok := <- edgeAccumulator
			if !ok {
				break
			}
			fmt.Printf("%v\n\n", v)
		}
		wg.Done()
	}()

	// map dependencies and devDependencies to trees of

	// reduce/accumulate results into file (reduce to csv where id is parent repo ID and one row for every dependency where the dependency is the GitHub URL)

	wg.Wait()
}
