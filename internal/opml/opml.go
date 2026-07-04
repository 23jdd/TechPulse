package opml

import (
	"encoding/xml"
	"io"

	"techpulse/internal/model"
)

type Document struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    Head     `xml:"head"`
	Body    Body     `xml:"body"`
}

type Head struct {
	Title string `xml:"title"`
}

type Body struct {
	Outlines []Outline `xml:"outline"`
}

type Outline struct {
	Text     string    `xml:"text,attr,omitempty"`
	Title    string    `xml:"title,attr,omitempty"`
	Type     string    `xml:"type,attr,omitempty"`
	XMLURL   string    `xml:"xmlUrl,attr,omitempty"`
	Category string    `xml:"category,attr,omitempty"`
	Children []Outline `xml:"outline,omitempty"`
}

func Encode(w io.Writer, feeds []model.RSSFeed) error {
	doc := Document{Version: "2.0", Head: Head{Title: "TechPulse Feeds"}}
	for _, feed := range feeds {
		doc.Body.Outlines = append(doc.Body.Outlines, Outline{
			Text: feed.Title, Title: feed.Title, Type: "rss", XMLURL: feed.URL, Category: feed.Category,
		})
	}
	_, _ = w.Write([]byte(xml.Header))
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	return enc.Encode(doc)
}

func Decode(r io.Reader) ([]model.RSSFeed, error) {
	var doc Document
	if err := xml.NewDecoder(r).Decode(&doc); err != nil {
		return nil, err
	}
	var feeds []model.RSSFeed
	collect(doc.Body.Outlines, &feeds)
	return feeds, nil
}

func collect(outlines []Outline, feeds *[]model.RSSFeed) {
	for _, outline := range outlines {
		if outline.XMLURL != "" {
			title := outline.Title
			if title == "" {
				title = outline.Text
			}
			*feeds = append(*feeds, model.RSSFeed{URL: outline.XMLURL, Title: title, Category: outline.Category, Status: "active"})
		}
		collect(outline.Children, feeds)
	}
}
