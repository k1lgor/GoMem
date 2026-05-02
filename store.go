package gomem

import (
	"fmt"
	"os"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search/query"
)

// Store wraps a Bleve index for persistent memory storage.
type Store struct {
	index bleve.Index
}

// NewStore opens an existing Bleve index at path or creates a new one.
func NewStore(path string) (*Store, error) {
	// Ensure data directory is only accessible by the owner
	os.MkdirAll(path, 0700)

	index, err := bleve.Open(path)
	if err != nil {
		// If the index doesn't exist (path missing) or the directory is empty
		// (meta missing), create a new index.
		if err == bleve.ErrorIndexPathDoesNotExist || err == bleve.ErrorIndexMetaMissing {
			m := buildIndexMapping()
			index, err = bleve.New(path, m)
		}
		if err != nil {
			return nil, fmt.Errorf("open index: %w", err)
		}
	}
	return &Store{index: index}, nil
}

// Remember stores a text entry in the index under the given ID.
func (s *Store) Remember(id, text string) error {
	doc := MemoryDoc{
		ID:   id,
		Text: text,
	}
	if err := s.index.Index(id, doc); err != nil {
		return fmt.Errorf("index document: %w", err)
	}
	return nil
}

// Search performs a full-text query against the index and returns matching hits.
func (s *Store) Search(q string, limit int) ([]SearchHit, uint64, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var qry query.Query
	if q == "*" {
		// MatchAllQuery for listing all documents
		qry = query.NewMatchAllQuery()
	} else {
		// MatchQuery (simpler, faster than QueryStringQuery).
		// The query text is analyzed with the field's analyzer for token matching.
		mq := query.NewMatchQuery(q)
		mq.SetField("text")
		qry = mq
	}
	searchRequest := bleve.NewSearchRequestOptions(qry, limit, 0, false)
	searchRequest.Fields = []string{"text"}

	result, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, 0, fmt.Errorf("search: %w", err)
	}

	hits := make([]SearchHit, 0, len(result.Hits))
	for _, hit := range result.Hits {
		hits = append(hits, SearchHit{
			ID:    hit.ID,
			Score: hit.Score,
			Text:  fieldString(hit.Fields, "text"),
		})
	}
	return hits, result.Total, nil
}

// Delete removes a document from the index by ID.
// Returns an error if the document doesn't exist.
func (s *Store) Delete(id string) error {
	// Search for the document first to verify it exists.
	// Bleve's Delete doesn't error on missing docs.
	q := query.NewDocIDQuery([]string{id})
	search := bleve.NewSearchRequestOptions(q, 1, 0, false)
	result, err := s.index.Search(search)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	if result.Total == 0 {
		return fmt.Errorf("document %q not found", id)
	}

	if err := s.index.Delete(id); err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	return nil
}

// Close closes the underlying Bleve index, flushing all pending writes.
func (s *Store) Close() error {
	return s.index.Close()
}

// DocCount returns the total number of indexed documents.
func (s *Store) DocCount() (uint64, error) {
	return s.index.DocCount()
}

// buildIndexMapping creates a Bleve index mapping for MemoryDoc.
func buildIndexMapping() mapping.IndexMapping {
	m := bleve.NewIndexMapping()

	docMapping := bleve.NewDocumentMapping()

	// ID field — keyword (not analyzed)
	idFieldMapping := bleve.NewTextFieldMapping()
	idFieldMapping.Analyzer = "keyword"
	docMapping.AddFieldMappingsAt("id", idFieldMapping)

	// Text field — analyzed for full-text search (no stemming = faster)
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = "standard"
	docMapping.AddFieldMappingsAt("text", textFieldMapping)

	m.AddDocumentMapping("memory", docMapping)
	m.DefaultAnalyzer = "standard"

	return m
}

// fieldString safely extracts a string from a document field map.
func fieldString(fields map[string]interface{}, key string) string {
	v, ok := fields[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}
