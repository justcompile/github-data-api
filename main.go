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
	client, err := lib.New("justcompile/github-data-api")

	checkErr(err)

	branch, created, err := client.GetOrCreateBranch("testing")

	checkErr(err)

	fmt.Printf("Branch: testing. Created: %t\n", created)

	tree, err := client.MakeChanges(branch, lib.NewChange("data/test.txt", lib.ReplaceAll("line", "LiNe")))

	checkErr(err)

	checkErr(client.Push(branch, tree))

	fmt.Println("ok!")
}
