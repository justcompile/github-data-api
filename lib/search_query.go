package lib

import "fmt"

type searchQuery struct {
	text     string
	language string
	repo     string
}

func (s *searchQuery) InFile() string {
	x := fmt.Sprintf(
		"%s+in:file+language:%s+repo:%s",
		s.text,
		s.language,
		s.repo,
	)

	fmt.Println(x)

	return x
}

func newQuery(text, language, repoName string) *searchQuery {
	return &searchQuery{
		text:     text,
		language: language,
		repo:     repoName,
	}
}
