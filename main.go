package main

import (
	"fmt"
	"log"

	"github.com/justcompile/github-data-api/lib"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	client, err := lib.New("FUTRLI/flux-s3-watcher")

	checkErr(err)

	branch, created, err := client.GetOrCreateBranch("master")

	checkErr(err)

	fmt.Printf("Branch: master. Created: %t\n", created)

	tree, err := client.MakeChanges(branch, lib.NewChange("data/test.txt", lib.ReplaceAll("aws.k8s.futrli.com/", "aws.k8s.futrli.com")))

	checkErr(err)

	checkErr(client.Push(branch, tree))

	fmt.Println("ok!")
}
