package model

type DependencyTree struct {
	ID           int
	Dependencies []string
}

type PackageEdges struct {
	ID 					 			int
	DependencyURLs 		[]string
}
