package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1lgor/gomem"
)

// Version is set at build time via -ldflags.
// Example: go build -ldflags="-X main.Version=v1.0.0" ./cmd/gomem/
var Version = "dev"

const usage = `GoMem — persistent memory for AI coding agents.

Stores memory in a project-local .gomem directory.
Each project gets its own isolated memory.

Core principle: Search GoMem FIRST, read files SECOND.
GoMem stores CONCISE SUMMARIES, not raw file contents,
saving 10x-100x tokens vs re-reading source files.

Usage:
  gomem <command> [arguments]

Commands:
  remember <id> <text>      Store a memory (summary, key insight, decision)
  search <query>            Search memories (use * for all)
  list                      List all memories
  delete <id>               Delete a memory by ID
  save-all                  Index current project as concise summaries
  skill <agent>             Install skill for an AI agent
  version                   Show version
  help                      Show this help

Examples:
  gomem remember auth-design "Auth uses JWT with RSA256, expires in 24h"
  gomem search "JWT authentication"
  gomem search "*"
  gomem list
  gomem save-all
  gomem skill pi
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
	case "save-all":
		cmdSaveAll()
	case "skill":
		cmdSkill(os.Args[2:])
	case "version", "--version", "-v":
		cmdVersion()
	case "help", "--help", "-h":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n\n", cmd)
		fmt.Print(usage)
		os.Exit(1)
	}
}

func dataDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot get current directory: %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(cwd, ".gomem")
}

func openStore() *gomem.Store {
	dir := dataDir()
	s, err := gomem.NewStore(dir)
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

func cmdSaveAll() {
	absDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	s := openStore()
	defer s.Close()

	fmt.Printf("Indexing project: %s\n", absDir)
	fmt.Println()

	count, err := gomem.SaveAll(absDir, s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during indexing: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Indexed %d files.\n", count)

	// Write AGENTS.md
	gomem.WriteAgentsMd(absDir)
	fmt.Printf("  ✓ AGENTS.md updated at project root\n")
	fmt.Printf("\nSearch with: gomem search \"<query>\"\n")
}

func cmdVersion() {
	fmt.Printf("GoMem %s\n", Version)
}

func cmdSkill(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: gomem skill <agent-name> [--global|-g]\n")
		fmt.Fprintf(os.Stderr, "\nInstall skill for an AI coding agent:\n")
		// Display agents
		primary := []struct{ key, desc string }{
			{"claude", gomem.Agents["claude"].Description},
			{"pi", gomem.Agents["pi"].Description},
			{"cline", gomem.Agents["cline"].Description},
			{"codex", gomem.Agents["codex"].Description},
			{"copilot", gomem.Agents["copilot"].Description},
			{"cursor", gomem.Agents["cursor"].Description},
			{"windsurf", gomem.Agents["windsurf"].Description},
			{"zed", gomem.Agents["zed"].Description},
			{"kilo", gomem.Agents["kilo"].Description},
			{"continue", gomem.Agents["continue"].Description},
			{"gemini", gomem.Agents["gemini"].Description},
		}
		for _, p := range primary {
			fmt.Fprintf(os.Stderr, "  %-16s %s\n", p.key, p.desc)
		}
		fmt.Fprintf(os.Stderr, "\nUse --global or -g to install in your home directory.\n")
		os.Exit(1)
	}

	agentName := strings.ToLower(args[0])
	global := false
	for _, a := range args[1:] {
		if a == "--global" || a == "-g" {
			global = true
		}
	}

	cfg, ok := gomem.Agents[agentName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown agent: %q\n", agentName)
		fmt.Fprintf(os.Stderr, "Run 'gomem skill' to see all supported agents.\n")
		os.Exit(1)
	}

	skillSrc, err := gomem.FindSkillSource()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var baseDir string
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot find home directory: %v\n", err)
			os.Exit(1)
		}
		baseDir = home
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		baseDir = cwd
	}

	targetDir := filepath.Join(baseDir, cfg.Dir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot create %s: %v\n", targetDir, err)
		os.Exit(1)
	}

	// Copy SKILL.md and scripts
	fmt.Printf("  ✓ SKILL.md\n")
	copyFile := func(src, dst string) {
		data, err := os.ReadFile(src)
		if err != nil {
			return
		}
		os.WriteFile(dst, data, 0644)
	}

	srcFile := filepath.Join(skillSrc, "SKILL.md")
	dstFile := filepath.Join(targetDir, "SKILL.md")
	copyFile(srcFile, dstFile)

	// Copy scripts
	srcScripts := filepath.Join(skillSrc, "scripts")
	if fi, err := os.Stat(srcScripts); err == nil && fi.IsDir() {
		dstScripts := filepath.Join(targetDir, "scripts")
		os.MkdirAll(dstScripts, 0755)
		entries, _ := os.ReadDir(srcScripts)
		for _, entry := range entries {
			if !entry.IsDir() {
				src := filepath.Join(srcScripts, entry.Name())
				dst := filepath.Join(dstScripts, entry.Name())
				copyFile(src, dst)
				fmt.Printf("  ✓ scripts/%s\n", entry.Name())
			}
		}
	}

	fmt.Printf("✓ Installed GoMem skill for %s\n", cfg.Name)
	fmt.Printf("  Location: %s\n", filepath.Join(targetDir, cfg.InstallFile))

	if !global {
		gomem.WriteAgentsMd(baseDir)
		fmt.Printf("  ✓ AGENTS.md updated at project root\n")
		fmt.Printf("\nRestart your agent. AGENTS.md will auto-instruct it to use GoMem.\n")
	}
}
