package helper

import (
	"fmt"
	"regexp"
	"strings"
)

var reNonAlpha = regexp.MustCompile(`[^a-z0-9]+`)

func Slug(name string, id int64) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = reNonAlpha.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "item"
	}
	return fmt.Sprintf("%s-%d", s, id)
}
