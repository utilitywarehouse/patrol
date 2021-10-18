package main

import (
	"log"
	"os"

	"github.com/uw-labs/patrol/patrol"
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

	//fmt.Println("own packages:")
	//for _, pkg := range repo.Packages {
	//if !pkg.PartOfModule {
	//continue
	//}
	//fmt.Println(pkg.Name)
	//}

	//fmt.Println()
	//fmt.Println("external packages:")
	//for _, pkg := range repo.Packages {
	//if pkg.PartOfModule {
	//continue
	//}
	//if len(pkg.Dependants) == 0 {
	//continue
	//}
	//fmt.Println(pkg.Name)
	//fmt.Println("dependants:")
	//for _, d := range pkg.Dependants {
	//fmt.Println("\t", d.Name)
	//}
	//}
}
