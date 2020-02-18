package lib

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
)

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
	entries := make([]github.TreeEntry, len(changes))

	for i, change := range changes {
		file, content, err := getFileContent(change.filePath)
		if err != nil {
			return nil, err
		}

		entries[i] = github.TreeEntry{
			Path: github.String(file),
			Type: github.String("blob"),
			Content: github.String(
				change.Apply(content),
			),
			Mode: github.String("100644"),
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

	//fmt.Printf("\n%v\n", github.Stringify(user))

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
		Parents: []github.Commit{*parent.Commit},
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

func emailOrDefault(email *string) *string {
	if email != nil {
		return email
	}

	return github.String("test@test.com")
}

// getFileContent loads the local content of a file and return the target name
// of the file in the target repository and its contents.
func getFileContent(fileArg string) (targetName string, b []byte, err error) {
	var localFile string
	files := strings.Split(fileArg, ":")
	switch {
	case len(files) < 1:
		return "", nil, errors.New("empty `-files` parameter")
	case len(files) == 1:
		localFile = files[0]
		targetName = files[0]
	default:
		localFile = files[0]
		targetName = files[1]
	}

	b, err = ioutil.ReadFile(localFile)
	return targetName, b, err
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
