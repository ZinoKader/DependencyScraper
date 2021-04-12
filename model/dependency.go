package model

type PackageEdges struct {
	ID             int
	DependencyURLs []string
}

type DependencyTree struct {
	ID           int
	Dependencies []string
}
