package gomem

// MemoryDoc is the document stored in the Bleve index.
type MemoryDoc struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// SearchHit is a single result from a search.
type SearchHit struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
	Text  string  `json:"text"`
}
