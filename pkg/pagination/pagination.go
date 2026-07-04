package pagination

import "net/http"

type Page struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"-"`
}

func FromRequest(r *http.Request) Page {
	page := atoiDefault(r.URL.Query().Get("page"), 1)
	size := atoiDefault(r.URL.Query().Get("page_size"), 20)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return Page{Page: page, PageSize: size, Offset: (page - 1) * size}
}

func atoiDefault(value string, fallback int) int {
	var n int
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return fallback
		}
		n = n*10 + int(ch-'0')
	}
	if value == "" {
		return fallback
	}
	return n
}
