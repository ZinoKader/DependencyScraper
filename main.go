package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ZinoKader/KEX/model"
	"github.com/ZinoKader/KEX/pkg/data"
	"github.com/ZinoKader/KEX/pkg/scraping"
	"github.com/patrickmn/go-cache"
)

var SLICES = runtime.NumCPU()

func mapPackageFiles(repos []model.RepositoryFileRow, treeAccumulator chan<- model.DependencyTree) {
	var wg sync.WaitGroup
	partSize := len(repos) / SLICES
	for i := 0; i < len(repos); i += partSize {
		var reposPart []model.RepositoryFileRow
		if i+partSize > len(repos) {
			reposPart = repos[i:]
			i = len(repos) // break out of next loop
		} else {
			reposPart = repos[i : i+partSize]
		}

		wg.Add(1)
		go func(wg *sync.WaitGroup, treeAccumulator chan<- model.DependencyTree, reposPart []model.RepositoryFileRow) {
			defer wg.Done()
			for _, row := range reposPart {
				URLParts := strings.Split(row.URL, "/")
				ownerName := URLParts[len(URLParts)-2]
				repoName := row.Name
				dependencyTree, err := scraping.RepoDependencyTree(ownerName, repoName)
				if err != nil {
					log.Printf("Something went wrong when scraping the dependency tree for repo %s\n%v\n", row.URL, err)
					switch err.(type) {
					case *model.ConnectionError:
						// handle marking this repo as "retry"
						connectionError := err.(*model.ConnectionError)
						e := data.AppendToFile("retry.txt", connectionError.RepositoryURL)
						if e != nil {
							log.Printf("Failed adding repository %s to retry list", connectionError.RepositoryURL)
						}
						continue
					case *model.RepoNoPackage:
						// handle marking this repo as "do-not-retry"
						/*
							noPackageError := err.(*model.RepoNoPackage)
							e := data.AppendToFile("no-retry.txt", noPackageError.RepositoryURL)
							if e != nil {
								log.Printf("Failed adding repository %s to do-not-retry list", noPackageError.RepositoryURL)
							}
						*/
						continue
					case *model.RepoNotExist:
						// handle marking this repo as "do-not-retry"
						/*
							notExistError := err.(*model.RepoNotExist)
							e := data.AppendToFile("no-retry.txt", notExistError.RepositoryURL)
							if e != nil {
								log.Printf("Failed adding repository %s to do-not-retry list", notExistError.RepositoryURL)
							}
						*/
						continue
					default:
						continue
					}
				} else {
					dependencyTree.ID = row.ID
					// push parsed dependency tree to accumulator
					treeAccumulator <- dependencyTree
				}
			}
		}(&wg, treeAccumulator, reposPart)
	}
	wg.Wait()
	close(treeAccumulator)
}

func getCache() *cache.Cache {
	//TODO: Read map from file
	file, err := os.Open("cache.json")
	if err != nil {
		return cache.New(cache.NoExpiration, 0)
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return cache.New(cache.NoExpiration, 0)
	}

	cacheMap := make(map[string]cache.Item)

	err = json.Unmarshal(data, &cacheMap)
	if err != nil {
		return cache.New(cache.NoExpiration, 0)
	}
	return cache.NewFrom(cache.NoExpiration, 0, cacheMap)
}

func saveCache(dependencyCache *cache.Cache) error {
	cachedData := dependencyCache.Items()

	jsonData, err := json.Marshal(cachedData)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile("cache.json", jsonData, os.ModePerm)

	return err
}

func mapDependencies(treeAccumulator <-chan model.DependencyTree, edgeAccumulator chan<- model.PackageEdges, dependencyCache *cache.Cache) {
	var wg sync.WaitGroup
	threads := SLICES / 2
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, treeActreeAccumulator <-chan model.DependencyTree, ededgeAccumulator chan<- model.PackageEdges) {
			defer wg.Done()
			for tree := range treeAccumulator {
				dependencyURLs := scraping.RepoDependencies(tree.Dependencies, dependencyCache)
				edges := model.PackageEdges{
					ID:             tree.ID,
					DependencyURLs: dependencyURLs,
				}
				edgeAccumulator <- edges
			}
		}(&wg, treeAccumulator, edgeAccumulator)
	}
	wg.Wait()
	close(edgeAccumulator)
}

func reduceToFile(edgeAccumulator <-chan model.PackageEdges, outputPath string) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	defer writer.Flush()

	for node := range edgeAccumulator {
		for _, edge := range node.DependencyURLs {
			writer.Write([]string{strconv.Itoa(node.ID), edge})
		}
	}
	data.WriteToFile(outputPath, buf.Bytes())

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
		log.Println("Usage: ", os.Args[0], "{input file path} [csv]", "{output file path} [csv]")
		return
	}
	inputPath := os.Args[1]
	outputPath := os.Args[2]
	repoRows := data.RepositoryFileRows(inputPath)

	dependencyCache := getCache()

	treeAccumulator := make(chan model.DependencyTree, SLICES/2)
	edgeAccumulator := make(chan model.PackageEdges, SLICES/2)

	setup()

	go mapPackageFiles(repoRows, treeAccumulator)
	go mapDependencies(treeAccumulator, edgeAccumulator, dependencyCache)
	reduceToFile(edgeAccumulator, outputPath)

	saveCache(dependencyCache)

	// map dependencies and devDependencies to trees of

	// reduce/accumulate results into file (reduce to csv where id is parent repo ID and one row for every dependency where the dependency is the GitHub URL)
}
