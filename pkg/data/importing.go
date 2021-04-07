package data

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	model "github.com/ZinoKader/KEX/model"
)

func GetRepositoryFileRows() []model.RepositoryFileRow {
	repoFile, err := os.Open("test.csv")
	if err != nil {
		fmt.Println("Could not open repository file", err)
	}
	defer repoFile.Close()

	csvLines, err := csv.NewReader(repoFile).ReadAll()
	if err != nil {
		fmt.Println("Could not read repository file lines", err)
	}

	var repoFileRows []model.RepositoryFileRow
	for _, line := range csvLines {
		var name = line[1]
		var url = line[2]
		id, err := strconv.Atoi(line[0])
		if err != nil {
			fmt.Printf("Could not load row of repo %s\n%v", url, err)
			continue
		}
		row := model.RepositoryFileRow{
			ID:   id,
			Name: name,
			URL:  url,
		}
		repoFileRows = append(repoFileRows, row)
	}

	return repoFileRows
}
