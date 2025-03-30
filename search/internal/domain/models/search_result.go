package models

type SearchResult struct {
	Answer     string      `json:"answer"`
	References []Reference `json:"references,omitempty"`
}
