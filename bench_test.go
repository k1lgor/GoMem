package gomem

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"testing"
	"time"
)

// words used to generate random documents for benchmarking.
var benchmarkWords = []string{
	"authentication", "authorization", "caching", "database", "encryption",
	"framework", "goroutine", "handler", "index", "json",
	"key", "language", "memory", "network", "optimization",
	"protocol", "query", "router", "server", "template",
	"unicode", "validator", "worker", "xml", "yaml",
	"buffer", "channel", "deploy", "event", "function",
}

// generateDoc creates a random document of approximately n bytes.
func generateDoc(rng *rand.Rand, n int) string {
	b := make([]byte, 0, n)
	for len(b) < n {
		word := benchmarkWords[rng.Intn(len(benchmarkWords))]
		if len(b) > 0 {
			b = append(b, ' ')
		}
		b = append(b, word...)
	}
	return string(b)
}

// BenchmarkSearch measures query latency against an index with N documents.
// Run with: go test -bench=BenchmarkSearch -benchtime=100x
func BenchmarkSearch(b *testing.B) {
	dir, err := os.MkdirTemp("", "gomem-bench-*")
	if err != nil {
		b.Fatal(err)
	}

	store, err := NewStore(dir)
	if err != nil {
		b.Fatal(err)
	}

	const numDocs = 10_000
	const docSize = 200

	rng := rand.New(rand.NewSource(42))
	docs := make([]string, numDocs)
	for i := 0; i < numDocs; i++ {
		id := fmt.Sprintf("doc-%d", i)
		text := generateDoc(rng, docSize)
		docs[i] = text
		if err := store.Remember(id, text); err != nil {
			b.Fatal(err)
		}
	}

	// Pre-generate queries that are known to exist in the corpus
	queries := make([]string, 100)
	for i := range queries {
		doc := docs[rng.Intn(numDocs)]
		words := splitWords(doc)
		if len(words) < 3 {
			continue
		}
		start := rng.Intn(len(words) - 2)
		queries[i] = fmt.Sprintf("%s %s %s", words[start], words[start+1], words[start+2])
	}

	// Warm-up: 10 queries
	for i := 0; i < 10; i++ {
		store.Search(queries[i%len(queries)], 10)
	}

	// Reset timer after warm-up and setup
	b.ResetTimer()

	latencies := make([]float64, b.N)
	for i := 0; i < b.N; i++ {
		start := time.Now()
		_, _, err := store.Search(queries[i%len(queries)], 10)
		elapsed := time.Since(start)
		if err != nil {
			b.Fatal(err)
		}
		latencies[i] = float64(elapsed.Microseconds()) / 1000.0 // ms
	}

	// Report percentiles
	sort.Float64s(latencies)
	p50 := latencies[len(latencies)/2]
	p95 := latencies[int(float64(len(latencies))*0.95)]
	p99 := latencies[int(float64(len(latencies))*0.99)]

	b.ReportMetric(p50, "P50_ms")
	b.ReportMetric(p95, "P95_ms")
	b.ReportMetric(p99, "P99_ms")

	// Cleanup: close store first, then remove temp dir
	store.Close()
	os.RemoveAll(dir)

	// Log hardware info for documentation
	b.Logf("OS: %s, Arch: %s, Go: %s", runtime.GOOS, runtime.GOARCH, runtime.Version())
	b.Logf("NumDocs: %d, DocSize: %d bytes", numDocs, docSize)
}

// splitWords splits a string into words (simple space split).
func splitWords(s string) []string {
	var words []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ' ' {
			if i > start {
				words = append(words, s[start:i])
			}
			start = i + 1
		}
	}
	return words
}
