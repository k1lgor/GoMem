package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const usage = `GoMem — persistent memory for AI coding agents.

Stores memory in a project-local .gomem directory.
Each project gets its own isolated memory.

Usage:
  gomem <command> [arguments]

Commands:
  remember <id> <text>   Store a memory
  search <query>         Search memories (use * for all)
  list                   List all memories
  delete <id>            Delete a memory by ID
  help                   Show this help

Examples:
  gomem remember my-key "The API uses JWT for authentication"
  gomem search "JWT authentication"
  gomem search "*"
  gomem list
  gomem delete my-key
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "remember":
		cmdRemember(os.Args[2:])
	case "search":
		cmdSearch(os.Args[2:])
	case "list":
		cmdList()
	case "delete":
		cmdDelete(os.Args[2:])
	case "help", "--help", "-h":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n\n", cmd)
		fmt.Print(usage)
		os.Exit(1)
	}
}

// dataDir returns the project-local .gomem directory.
// It uses the current working directory as the project root.
func dataDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot get current directory: %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(cwd, ".gomem")
}

// openStore opens the project-local store, creating it if needed.
func openStore() *Store {
	dir := dataDir()
	s, err := NewStore(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot open memory store: %v\n", err)
		os.Exit(1)
	}
	return s
}

func requireArgs(args []string, n int, usageLine string) {
	if len(args) < n {
		fmt.Fprintf(os.Stderr, "Usage: gomem %s\n", usageLine)
		os.Exit(1)
	}
}

func cmdRemember(args []string) {
	requireArgs(args, 2, "remember <id> <text>")
	id := args[0]
	text := strings.Join(args[1:], " ")

	s := openStore()
	defer s.Close()

	if err := s.Remember(id, text); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Remembered %q\n", id)
}

func cmdSearch(args []string) {
	requireArgs(args, 1, "search <query>")
	query := strings.Join(args, " ")

	s := openStore()
	defer s.Close()

	hits, total, err := s.Search(query, 100)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if total == 0 {
		fmt.Println("No results found.")
		return
	}

	fmt.Printf("Found %d result(s):\n\n", total)
	for _, hit := range hits {
		fmt.Printf("  [%s] (score: %.3f)\n", hit.ID, hit.Score)
		fmt.Printf("  %s\n\n", hit.Text)
	}
}

func cmdList() {
	s := openStore()
	defer s.Close()

	// Use wildcard search to get all documents
	hits, total, err := s.Search("*", 100)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if total == 0 {
		fmt.Println("No memories stored yet.")
		return
	}

	fmt.Printf("Total memories: %d\n\n", total)
	for _, hit := range hits {
		text := hit.Text
		if len(text) > 80 {
			text = text[:77] + "..."
		}
		fmt.Printf("  %-30s  %s\n", hit.ID, text)
	}
}

func cmdDelete(args []string) {
	requireArgs(args, 1, "delete <id>")
	id := args[0]

	s := openStore()
	defer s.Close()

	if err := s.Delete(id); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Deleted %q\n", id)
}
