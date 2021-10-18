package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/utilitywarehouse/patrol/patrol"
)

func main() {
	revision := flag.String("from", "", "revision that should be used to detected "+
		"changes in HEAD.\nE.g.: -from=a0e002f951f56d53d552f9427b3331b11ea66e92")

	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "please provide the path to the repository\n")
		os.Exit(1)
	}

	if *revision == "" {
		fmt.Fprintf(os.Stderr, "please set `from` flag:\n\tpatrol -from=a0e002f951f56d53d552f9427b3331b11ea66e92 .\n")
		os.Exit(1)
	}

	repoPath := args[0]

	repo, err := patrol.NewRepo(repoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}

	changes, err := repo.ChangesFrom(*revision)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}

	for _, c := range changes {
		fmt.Println(c)
	}
}
