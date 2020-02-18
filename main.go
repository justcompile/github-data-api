package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v29/github"
)

func main() {
	client := github.NewClient(nil)

	repos, _, err := client.Repositories.List(context.Background(), "justcompile", nil)

	if err != nil {
		log.Fatal(err)
	}

	for _, org := range repos {
		fmt.Println(*org.Name)
	}
}
