package lib

import (
	"strings"
)

type Replacement interface {
	GetSearchText() string
	Replace(string) (string, int)
}

type Change struct {
	filePath    string
	replacement Replacement
}

func (c *Change) Apply(data []byte) string {
	changedData, _ := c.replacement.Replace(string(data))

	return changedData
}

func NewChange(path string, replacement Replacement) *Change {
	return &Change{
		filePath:    path,
		replacement: replacement,
	}
}

type simpleTextReplace struct {
	find    string
	replace string
}

func (s *simpleTextReplace) Replace(input string) (string, int) {
	count := strings.Count(input, s.find)
	return strings.ReplaceAll(input, s.find, s.replace), count
}

func (s *simpleTextReplace) GetSearchText() string {
	return s.find
}

func ReplaceAll(find, replace string) Replacement {
	return &simpleTextReplace{
		find:    find,
		replace: replace,
	}
}
