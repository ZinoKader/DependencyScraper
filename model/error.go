package model

import "fmt"

type RepoNotExist struct {
	RepositoryURL string
}

func (e *RepoNotExist) Error() string {
	return fmt.Sprintf("The repository with URL %s did not exist\n", e.RepositoryURL)
}

type RepoNoPackage struct {
	RepositoryURL string
}

func (e *RepoNoPackage) Error() string {
	return fmt.Sprintf("The repository with URL %s did not have a package.json file\n", e.RepositoryURL)
}

type ConnectionError struct {
	RepositoryURL string
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("The repository with URL %s could not be scraped\n", e.RepositoryURL)
}
