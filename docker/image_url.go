package docker

import (
	"fmt"
	"strings"
)

// ImageUrl provides a structured mechanism for dealing with docker Image URLs
// This is commonly used to alter a single section of the Image URL when deploying
type ImageUrl struct {
	RepoUrl string
	Name    string
	Tag     string
	Digest  string
}

func (u ImageUrl) String() string {
	cur := u.Name
	if u.RepoUrl != "" {
		cur = fmt.Sprintf("%s/%s", u.RepoUrl, cur)
	}
	if u.Tag != "" {
		cur = fmt.Sprintf("%s:%s", cur, u.Tag)
	}
	if u.Digest != "" {
		cur = fmt.Sprintf("%s@%s", cur, u.Digest)
	}
	return cur
}

func ParseImageUrl(raw string) ImageUrl {
	it := ImageUrl{Name: raw}

	if tokens := strings.SplitN(raw, "/", 2); len(tokens) == 2 {
		it.RepoUrl = tokens[0]
		it.Name = tokens[1]
	}

	if tokens := strings.SplitN(it.Name, "@", 2); len(tokens) == 2 {
		it.Name = tokens[0]
		it.Digest = tokens[1]
	} else if tokens = strings.SplitN(it.Name, ":", 2); len(tokens) == 2 {
		it.Name = tokens[0]
		it.Tag = tokens[1]
	}

	return it
}
