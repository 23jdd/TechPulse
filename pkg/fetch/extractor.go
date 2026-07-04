package fetch

import (
	url2 "net/url"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
)

type ExtractedContent struct {
	Title       string
	Markdown    string
	Text        string
	Links       []string
	Images      []string
	Author      string
	PublishedAt time.Time
	SourceURL   string
	Chunks      []string
}

type Extractor struct {
	converter *htmltomarkdown.Converter
}

func NewExtractor() *Extractor {
	return &Extractor{
		converter: htmltomarkdown.NewConverter("", true, nil),
	}
}

func (e *Extractor) Extract(url string, rawHTML string) (*ExtractedContent, error) {

	html := wrapHTML(rawHTML)

	parse, err := url2.Parse(url)
	if err != nil {
		return nil, err
	}
	article, err := readability.FromReader(strings.NewReader(html), parse)
	if err != nil {
		return e.fallback(url, html)
	}

	md, err := e.converter.ConvertString(article.Content)
	if err != nil {
		return nil, err
	}

	md = postProcessMarkdown(md)

	links, images := extractAssets(article.Content)

	text := stripMarkdown(md)

	return &ExtractedContent{
		Title:       article.Title,
		Markdown:    md,
		Text:        text,
		Links:       links,
		Images:      images,
		Author:      article.Byline,
		PublishedAt: time.Now(),
		SourceURL:   url,
		Chunks:      chunkText(text, 800),
	}, nil
}

func (e *Extractor) fallback(url string, html string) (*ExtractedContent, error) {

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	doc.Find("script, style, nav, footer, aside, ads, header").Remove()

	text := doc.Text()

	md, _ := e.converter.ConvertString(text)

	md = postProcessMarkdown(md)

	return &ExtractedContent{
		Title:     "",
		Markdown:  md,
		Text:      stripMarkdown(md),
		SourceURL: url,
		Chunks:    chunkText(text, 800),
	}, nil
}

func wrapHTML(html string) string {
	return "<html><body>" + html + "</body></html>"
}

func extractAssets(html string) ([]string, []string) {

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, nil
	}

	var links []string
	var images []string

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href != "" {
			links = append(links, href)
		}
	})

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		if src != "" {
			images = append(images, src)
		}
	})

	return links, images
}

func postProcessMarkdown(md string) string {

	md = normalize(md)
	md = fixCode(md)
	md = fixHeaders(md)
	md = removeNoise(md)

	return md
}

func normalize(md string) string {
	lines := strings.Split(md, "\n")

	var out []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}

	return strings.Join(out, "\n")
}

func fixCode(md string) string {
	md = strings.ReplaceAll(md, "``` ", "```")
	md = strings.ReplaceAll(md, " ```", "```")
	return md
}

func fixHeaders(md string) string {
	return strings.ReplaceAll(md, "####", "###")
}

func removeNoise(md string) string {

	bad := []string{
		"subscribe",
		"advertisement",
		"read more",
		"follow us",
		"share this",
	}

	var out []string

	for _, l := range strings.Split(md, "\n") {
		skip := false

		for _, b := range bad {
			if strings.Contains(strings.ToLower(l), b) {
				skip = true
				break
			}
		}

		if !skip {
			out = append(out, l)
		}
	}

	return strings.Join(out, "\n")
}

// ===============================
// TEXT STRIP (for RAG)
// ===============================
func stripMarkdown(md string) string {
	md = strings.ReplaceAll(md, "#", "")
	md = strings.ReplaceAll(md, "*", "")
	md = strings.ReplaceAll(md, "`", "")
	return md
}

func chunkText(text string, size int) []string {

	words := strings.Fields(text)

	var chunks []string

	for i := 0; i < len(words); i += size {
		end := i + size
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[i:end], " "))
	}

	return chunks
}
