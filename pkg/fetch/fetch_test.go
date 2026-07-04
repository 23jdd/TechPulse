package fetch

import (
	"fmt"
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestFeed(t *testing.T) {
	parser := gofeed.NewParser()

	feed, err := parser.ParseURL("https://go.dev/blog/feed.atom")
	if err != nil {
		panic(err)
	}
	extractor := NewExtractor()
	c, err := extractor.Extract(feed.Items[0].Link, feed.Items[0].Content)
	if err != nil {
		panic(err)
	}
	fmt.Println(c.Markdown)
}
