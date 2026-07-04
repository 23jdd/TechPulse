package parser

import (
	"html"
	"regexp"
	"strings"
)

var tagRe = regexp.MustCompile(`<[^>]+>`)
var spaceRe = regexp.MustCompile(`\s+`)

func CleanHTML(input string) string {
	text := tagRe.ReplaceAllString(input, " ")
	text = html.UnescapeString(text)
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = spaceRe.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func DetectLanguage(text string) string {
	for _, r := range text {
		if r > 127 {
			return "mixed"
		}
	}
	return "en"
}
