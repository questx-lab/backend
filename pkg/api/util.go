package api

import (
	"net/url"
	"strings"
)

func PercentEncode(s string) string {
	s = url.QueryEscape(s)
	return strings.ReplaceAll(s, "+", "%20")
}
