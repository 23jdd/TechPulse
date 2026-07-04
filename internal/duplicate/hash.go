package duplicate

import "techpulse/pkg/hash"

func URLHash(url string) string {
	return hash.SHA256(url)
}

func ContentHash(content string) string {
	return hash.SHA256(content)
}
