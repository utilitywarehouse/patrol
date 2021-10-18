package main

import (
	"log"
	"os"

	"github.com/utilitywarehouse/patrol/patrol"
	"golang.org/x/mod/modfile"
)

type Repo struct {
	Module *modfile.File
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("please provide the path to the repository")
	}

	repoPath := os.Args[1]

	_, err := patrol.NewRepo(repoPath)
	if err != nil {
		panic(err)
	}
}
