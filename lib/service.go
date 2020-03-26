package lib

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v30/github"
	"golang.org/x/oauth2"
)

type searchResult struct {
	path string
	sha  string
}

type fileMatch struct {
	path string
	data []byte
}

type Github struct {
	client *github.Client
	owner  string
	repo   string
	ctx    context.Context
}

func (g *Github) GetOrCreateBranch(name string) (*github.Reference, bool, error) {
	var err error

	if ref, _, err := g.client.Git.GetRef(g.ctx, g.owner, g.repo, "refs/heads/"+name); err == nil {
		return ref, false, nil
	}

	var baseRef *github.Reference
	if baseRef, _, err = g.client.Git.GetRef(g.ctx, g.owner, g.repo, "refs/heads/master"); err != nil {
		return nil, false, err
	}

	newRef := &github.Reference{Ref: github.String("refs/heads/" + name), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err := g.client.Git.CreateRef(g.ctx, g.owner, g.repo, newRef)
	return ref, true, err
}

func (g *Github) MakeChanges(branch *github.Reference, changes ...*Change) (*github.Tree, error) {
	entries := make([]*github.TreeEntry, len(changes))

	for i, change := range changes {
		blobs, err := g.getFileContent(change.replacement.GetSearchText())
		if err != nil {
			return nil, err
		}

		for _, blob := range blobs {
			entries[i] = &github.TreeEntry{
				Path: github.String(blob.path),
				Type: github.String("blob"),
				Content: github.String(
					change.Apply(blob.data),
				),
				Mode: github.String("100644"),
			}
		}
	}

	tree, _, err := g.client.Git.CreateTree(g.ctx, g.owner, g.repo, *branch.Object.SHA, entries)
	return tree, err
}

func (g *Github) Push(branch *github.Reference, tree *github.Tree) error {
	// Get the parent commit to attach the commit to.
	parent, _, err := g.client.Repositories.GetCommit(g.ctx, g.owner, g.repo, *branch.Object.SHA)
	if err != nil {
		return err
	}
	// This is not always populated, but is needed.
	parent.Commit.SHA = parent.SHA
	user, err := g.currentUser()
	if err != nil {
		return err
	}

	fmt.Printf("\n%v\n", github.Stringify(user))

	// Create the commit using the tree.
	date := time.Now()
	author := &github.CommitAuthor{
		Date:  &date,
		Name:  user.Login,
		Email: emailOrDefault(user.Email),
	}

	commit := &github.Commit{
		Author:  author,
		Message: github.String("A Message"),
		Tree:    tree,
		Parents: []*github.Commit{parent.Commit},
	}

	newCommit, _, err := g.client.Git.CreateCommit(g.ctx, g.owner, g.repo, commit)
	if err != nil {
		return err
	}

	// Attach the commit to the  branch.
	branch.Object.SHA = newCommit.SHA
	_, _, err = g.client.Git.UpdateRef(g.ctx, g.owner, g.repo, branch, false)
	return err
}

func (g *Github) currentUser() (*github.User, error) {
	user, _, err := g.client.Users.Get(g.ctx, "")

	return user, err
}

// getFileContent loads the local content of a file and return the target name
// of the file in the target repository and its contents.
func (g *Github) getFileContent(searchText string) ([]*fileMatch, error) {
	results, err := g.search(searchText)
	if err != nil {
		return nil, err
	}

	matches := make([]*fileMatch, len(results))

	for i, res := range results {
		data, _, err := g.client.Git.GetBlobRaw(g.ctx, g.owner, g.repo, res.sha)
		if err != nil {
			return nil, err
		}

		matches[i] = &fileMatch{path: res.path, data: data}
	}

	return matches, nil
}

func (g *Github) search(text string) ([]searchResult, error) {
	query := newQuery(text, "go", g.owner+"/"+g.repo)

	results, _, err := g.client.Search.Code(g.ctx, query.InFile(), &github.SearchOptions{TextMatch: true})
	if err != nil {
		return nil, err
	}

	searchResults := make([]searchResult, 0)

	for _, res := range results.CodeResults {
		if *res.Repository.Name == g.repo {
			fmt.Printf("%s\n", res.GetPath())
			searchResults = append(searchResults, searchResult{path: res.GetPath(), sha: res.GetSHA()})
		}
	}

	return searchResults, nil
}

func emailOrDefault(email *string) *string {
	if email != nil {
		return email
	}

	return github.String("test@test.com")
}

func New(repoName string) (*Github, error) {
	ctx := context.Background()
	parts := strings.Split(repoName, "/")

	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("Unauthorized: No token present")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return &Github{
		client: github.NewClient(tc),
		owner:  parts[0],
		repo:   parts[1],
		ctx:    ctx,
	}, nil
}
