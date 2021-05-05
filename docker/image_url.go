package docker

import (
	"encoding/json"
	"fmt"
	"strings"
)

var _ json.Unmarshaler = &ImageUrl{}

// ImageUrl provides a structured mechanism for dealing with docker Image URLs
// This is commonly used to alter a single section of the Image URL when deploying
type ImageUrl struct {
	Registry string
	User     string
	Repo     string
	Tag      string
	Digest   string
}

func (u *ImageUrl) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*u = ParseImageUrl(str)
	return nil
}

func (u ImageUrl) String() string {
	cur := u.Repo
	if u.User != "" {
		cur = fmt.Sprintf("%s/%s", u.User, cur)
	}
	if u.Registry != "" {
		cur = fmt.Sprintf("%s/%s", u.Registry, cur)
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
	it := ImageUrl{Repo: raw}

	tokens := strings.SplitN(raw, "/", 3)
	if len(tokens) == 3 {
		// registry/user/repo
		it.Registry = tokens[0]
		it.User = tokens[1]
		it.Repo = tokens[2]
	} else if len(tokens) == 2 {
		if strings.ContainsAny(tokens[0], ".:") || tokens[0] == "localhost" {
			// This signifies a domain if the first identifier contains '.', ':' or equals 'localhost'
			// registry/repo
			it.Registry = tokens[0]
			it.Repo = tokens[1]
		} else {
			// user/repo
			it.User = tokens[0]
			it.Repo = tokens[1]
		}
	}

	if tokens := strings.SplitN(it.Repo, "@", 2); len(tokens) == 2 {
		it.Repo = tokens[0]
		it.Digest = tokens[1]
	} else if tokens = strings.SplitN(it.Repo, ":", 2); len(tokens) == 2 {
		it.Repo = tokens[0]
		it.Tag = tokens[1]
	}

	return it
}
