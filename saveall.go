package gomem

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SaveAll indexes a project into GoMem with CONCISE summaries of each file.
// This saves 10x-100x tokens compared to re-reading raw files.
// It opens or creates a store at rootDir/.gomem, indexes all files,
// and returns the number of files indexed.
func SaveAll(rootDir string, store *Store) (fileCount int, err error) {
	projectName := filepath.Base(rootDir)

	// Store project overview
	overview := summarizeProject(rootDir, projectName)
	if err := store.Remember("project-overview", overview); err != nil {
		return 0, fmt.Errorf("store overview: %w", err)
	}

	// Store directory tree
	tree := buildCompactTree(rootDir)
	if err := store.Remember("project-structure", tree); err != nil {
		return 0, fmt.Errorf("store tree: %w", err)
	}

	// Walk and summarize each file
	err = filepath.Walk(rootDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(rootDir, path)

		if fi.IsDir() {
			name := fi.Name()
			if name == ".git" || name == "node_modules" || name == ".gomem" ||
				name == ".oh-my-pi" || name == ".pi" || name == "vendor" ||
				name == "target" || name == "dist" || name == "build" ||
				name == "__pycache__" || name == ".venv" ||
				strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if fi.Size() > 500*1024 {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if isBinaryExt(ext) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil || len(data) == 0 {
			return nil
		}

		content := string(data)
		id := "file:" + filepath.ToSlash(rel)

		summary := summarizeFile(rel, content, ext)
		if summary == "" {
			return nil
		}

		if err := store.Remember(id, summary); err != nil {
			return fmt.Errorf("index %s: %w", rel, err)
		}

		fileCount++
		return nil
	})

	if err != nil {
		return fileCount, fmt.Errorf("walk: %w", err)
	}

	return fileCount, nil
}

// summarizeProject creates a one-line project overview.
func summarizeProject(root, name string) string {
	// Count files by type
	goFiles, mdFiles, otherFiles := 0, 0, 0
	filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".go":
			goFiles++
		case ".md":
			mdFiles++
		default:
			otherFiles++
		}
		return nil
	})

	total := goFiles + mdFiles + otherFiles
	return fmt.Sprintf("Project %s: %d files (%d Go, %d Markdown, %d other). Root: %s",
		name, total, goFiles, mdFiles, otherFiles, root)
}

// buildCompactTree creates a 2-level deep directory tree.
func buildCompactTree(root string) string {
	base := filepath.Base(root)
	var b strings.Builder
	b.WriteString(base + "/\n")
	buildTreeLevel(root, "", 0, 2, &b)
	return b.String()
}

// buildTreeLevel recursively builds a directory tree.
func buildTreeLevel(dir string, prefix string, depth int, maxDepth int, b *strings.Builder) {
	if depth >= maxDepth {
		b.WriteString(prefix + "...\n")
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	skipDirs := map[string]bool{
		".git": true, "node_modules": true, ".gomem": true,
		".oh-my-pi": true, ".pi": true,
	}

	for i, entry := range entries {
		if skipDirs[entry.Name()] {
			continue
		}

		isLast := i == len(entries)-1
		marker := "├── "
		if isLast {
			marker = "└── "
		}

		if entry.IsDir() {
			b.WriteString(prefix + marker + entry.Name() + "/\n")
			nextPrefix := prefix + "│   "
			if isLast {
				nextPrefix = prefix + "    "
			}
			buildTreeLevel(filepath.Join(dir, entry.Name()), nextPrefix, depth+1, maxDepth, b)
		} else {
			b.WriteString(prefix + marker + entry.Name() + "\n")
		}
	}
}

// summarizeFile generates a concise summary for a file based on its extension.
func summarizeFile(path, content, ext string) string {
	switch ext {
	// Go — full structural extraction
	case ".go":
		return summarizeGoFile(content)

	// Markdown / docs
	case ".md":
		return summarizeMarkdown(content)
	case ".rst":
		return summarizeText(content)

	// Config / data formats
	case ".json":
		return summarizeJSON(content)
	case ".yaml", ".yml":
		return summarizeYAML(content)
	case ".toml":
		return summarizeTOML(content)
	case ".xml":
		return summarizeGenericCode(content, "<!--", "-->|<!--")
	case ".proto":
		return summarizeGenericCode(content, "//", "")
	case ".graphql", ".gql":
		return summarizeGenericCode(content, "#", "")
	case ".env":
		return summarizeEnv(content)

	// Shell / scripts
	case ".sh", ".bash", ".zsh", ".fish":
		return summarizeScript(content)
	case ".bat", ".cmd":
		return summarizeScript(content)
	case ".ps1", ".psm1":
		return summarizeScript(content)

	// Python
	case ".py":
		return summarizeGenericCode(content, "#", `"""|'''`)

	// JavaScript / TypeScript
	case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".mts", ".cts":
		return summarizeGenericCode(content, "//", "/*|*/")

	// Java & JVM languages
	case ".java":
		return summarizeGenericCode(content, "//", "/*|*/")
	case ".kt", ".kts":
		return summarizeGenericCode(content, "//", "/*|*/")
	case ".scala":
		return summarizeGenericCode(content, "//", "/*|*/")
	case ".groovy":
		return summarizeGenericCode(content, "//", "/*|*/")
	case ".clj", ".cljs", ".cljc", ".edn":
		return summarizeGenericCode(content, ";", "")

	// .NET
	case ".cs", ".csx":
		return summarizeGenericCode(content, "//", "/*|*/")
	case ".fs", ".fsx":
		return summarizeGenericCode(content, "//", "(*|*)")
	case ".vb":
		return summarizeGenericCode(content, "'", "")

	// C family
	case ".c", ".h":
		return summarizeGenericCode(content, "//", "/*|*/")
	case ".cpp", ".cxx", ".cc", ".hpp", ".hxx":
		return summarizeGenericCode(content, "//", "/*|*/")
	case ".m", ".mm":
		return summarizeGenericCode(content, "//", "/*|*/")

	// Rust
	case ".rs":
		return summarizeGenericCode(content, "//", "/*|*/")

	// Swift
	case ".swift":
		return summarizeGenericCode(content, "//", "/*|*/")

	// Ruby
	case ".rb", ".rbw", ".gemspec", ".rake":
		return summarizeGenericCode(content, "#", "=begin|=end")

	// PHP
	case ".php", ".phtml", ".php3", ".php4", ".php5", ".php7", ".phps":
		return summarizeGenericCode(content, "//", "/*|*/")

	// Dart
	case ".dart":
		return summarizeGenericCode(content, "//", "/*|*/")

	// Lua
	case ".lua":
		return summarizeGenericCode(content, "--", "--[[|]]")

	// Perl
	case ".pl", ".pm", ".t":
		return summarizeGenericCode(content, "#", "=cut")

	// Haskell
	case ".hs", ".lhs":
		return summarizeGenericCode(content, "--", "{-|-}")

	// Elixir
	case ".ex", ".exs":
		return summarizeGenericCode(content, "#", "")

	// Erlang
	case ".erl", ".hrl":
		return summarizeGenericCode(content, "%", "")

	// SQL
	case ".sql":
		return summarizeGenericCode(content, "--", "/*|*/")

	// R
	case ".r", ".R", ".Rmd":
		return summarizeGenericCode(content, "#", "")

	// Zig
	case ".zig":
		return summarizeGenericCode(content, "//", "")

	// Nim
	case ".nim":
		return summarizeGenericCode(content, "#", "")

	// Terraform / HCL
	case ".tf", ".tfvars", ".hcl":
		return summarizeGenericCode(content, "#", "/*|*/")

	// Docker / Make / Vagrant (detected by filename, ext="")
	case "":
		base := strings.ToLower(filepath.Base(path))
		if base == "dockerfile" || strings.HasPrefix(base, "dockerfile.") {
			return summarizeDockerfile(content)
		}
		if base == "makefile" || base == "gnumakefile" {
			return summarizeMakefile(content)
		}
		if base == "vagrantfile" {
			return summarizeGenericCode(content, "#", "")
		}
		return summarizeText(content)

	// Web styles
	case ".css", ".scss", ".sass", ".less":
		return summarizeGenericCode(content, "/*", "*/")
	case ".html", ".htm", ".xhtml":
		return summarizeGenericCode(content, "<!--", "-->")

	// Plain text / config
	case ".txt", ".cfg", ".conf", ".ini", ".properties":
		return summarizeText(content)
	case ".gitignore", ".dockerignore", ".gitattributes", ".editorconfig":
		return summarizeText(content)

	default:
		return summarizeText(content)
	}
}

// isBinaryExt returns true for binary file extensions.
func isBinaryExt(ext string) bool {
	binaryExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".bin": true, ".o": true, ".a": true, ".lib": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
		".ico": true, ".svg": true, ".woff": true, ".woff2": true,
		".ttf": true, ".eot": true, ".zip": true, ".tar": true,
		".gz": true, ".bz2": true, ".7z": true, ".rar": true,
		".pyc": true, ".pyo": true, ".class": true, ".jar": true,
		".wasm": true, ".deb": true, ".rpm": true, ".dmg": true,
		".iso": true, ".db": true, ".sqlite": true, ".sum": true,
		".mod": true,
	}
	return binaryExts[ext]
}

// --- Language-specific summarizers ---

func summarizeGoFile(content string) string {
	extract := structSummary{}

	// Package
	if m := regexp.MustCompile(`package (\w+)`).FindStringSubmatch(content); len(m) > 1 {
		extract.pkg = m[1]
	}

	// Imports (compact one line)
	imports := extractImports(content)
	if imports != "" {
		extract.imports = imports
	}

	// Structs
	structRe := regexp.MustCompile(`(?s)type (\w+) struct \{([^}]*)\}`)
	for _, m := range structRe.FindAllStringSubmatch(content, -1) {
		fields := extractFields(m[2])
		extract.structs = append(extract.structs, fmt.Sprintf("%s{%s}", m[1], fields))
	}

	// Types
	typeRe := regexp.MustCompile(`type (\w+) (string|int|float|bool|\[\].*?|map\[.*?\]|func\(.*?\)|interface)`)
	for _, m := range typeRe.FindAllStringSubmatch(content, -1) {
		extract.types = append(extract.types, fmt.Sprintf("%s (%s)", m[1], m[2]))
	}

	// Functions and methods with doc comments
	funcRe := regexp.MustCompile(`(?m)^(// .*?\n)?func\s+(\([^)]*\)\s+)?(\w+)\(([^)]*)\)\s*([^(]*)`)
	for _, m := range funcRe.FindAllStringSubmatch(content, -1) {
		doc := strings.TrimSpace(m[1])
		receiver := strings.TrimSpace(m[2])
		name := m[3]
		params := m[4]
		returns := strings.TrimSpace(m[5])

		sig := name + "(" + params + ")"
		if returns != "" && returns != "{" {
			sig += " " + returns
		}
		if receiver != "" {
			sig = receiver + " " + sig
		}
		if doc != "" {
			sig = doc + " " + sig
		}
		extract.funcs = append(extract.funcs, sig)
	}

	return buildSummary(extract)
}

func summarizeMarkdown(content string) string {
	// Extract title and headings
	lines := strings.Split(content, "\n")
	var headings []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			headings = append(headings, strings.TrimPrefix(line, "# "))
		} else if strings.HasPrefix(line, "## ") {
			headings = append(headings, "  "+strings.TrimPrefix(line, "## "))
		}
	}

	// First paragraph
	var firstPara string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "```") {
			if len(line) > 200 {
				firstPara = line[:200] + "..."
			} else {
				firstPara = line
			}
			break
		}
	}

	var b strings.Builder
	if len(headings) > 0 {
		b.WriteString("Title: " + headings[0] + ". ")
		if len(headings) > 1 {
			b.WriteString("Sections: " + strings.Join(headings[1:], ", ") + ". ")
		}
	}
	if firstPara != "" {
		b.WriteString(firstPara)
	}
	return b.String()
}

func summarizeText(content string) string {
	lines := strings.Split(content, "\n")
	var nonEmpty []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, ";") {
			nonEmpty = append(nonEmpty, line)
		}
	}

	if len(nonEmpty) == 0 {
		return ""
	}

	// First 3 meaningful lines, truncated
	var b strings.Builder
	maxLines := 3
	for i := 0; i < len(nonEmpty) && i < maxLines; i++ {
		line := nonEmpty[i]
		if len(line) > 150 {
			line = line[:147] + "..."
		}
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(line)
	}
	if len(nonEmpty) > maxLines {
		b.WriteString(fmt.Sprintf(" (...%d more lines)", len(nonEmpty)-maxLines))
	}
	return b.String()
}

// Helper types and functions

type structSummary struct {
	pkg     string
	imports string
	structs []string
	types   []string
	funcs   []string
}

func buildSummary(s structSummary) string {
	var parts []string

	if s.pkg != "" {
		parts = append(parts, "Package "+s.pkg)
	}
	if s.imports != "" {
		parts = append(parts, "Imports: "+s.imports)
	}
	if len(s.types) > 0 {
		parts = append(parts, "Types: "+strings.Join(s.types, ", "))
	}
	if len(s.structs) > 0 {
		parts = append(parts, "Structs: "+strings.Join(s.structs, ", "))
	}
	if len(s.funcs) > 0 {
		parts = append(parts, "Functions: "+strings.Join(s.funcs, ", "))
	}

	return strings.Join(parts, ". ")
}

func extractImports(content string) string {
	re := regexp.MustCompile(`(?s)import \((.*?)\)`)
	m := re.FindStringSubmatch(content)
	if m == nil {
		// Single import line
		re2 := regexp.MustCompile(`import "([^"]+)"`)
		m2 := re2.FindStringSubmatch(content)
		if m2 != nil {
			return shortenPath(m2[1])
		}
		return ""
	}

	imports := strings.Split(m[1], "\n")
	var cleaned []string
	for _, imp := range imports {
		imp = strings.TrimSpace(imp)
		if imp == "" {
			continue
		}
		// Remove quotes and alias
		parts := strings.Fields(imp)
		path := parts[len(parts)-1]
		path = strings.Trim(path, `"`)
		cleaned = append(cleaned, shortenPath(path))
	}
	return strings.Join(cleaned, ", ")
}

func shortenPath(path string) string {
	// Take last 2 segments
	parts := strings.Split(path, "/")
	if len(parts) <= 2 {
		return path
	}
	return strings.Join(parts[len(parts)-2:], "/")
}

func extractFields(structBody string) string {
	lines := strings.Split(structBody, "\n")
	var fields []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		// Extract just the field name
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			name := parts[0]
			if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
				fields = append(fields, name)
			}
		}
	}
	return strings.Join(fields, ", ")
}

// Stub summarizers for other languages (can be expanded)
// --- Generic code summarizer ---

// summarizeGenericCode extracts structural info from any code file.
// lineComment and blockComment are comment markers for the language.
// blockComment can be "open|close" for languages with paired delimiters.
func summarizeGenericCode(content, lineComment, blockComment string) string {
	// Extract first doc comment as summary
	summary := extractDocComment(content, lineComment, blockComment)

	// Extract class/interface/enum declarations
	classRe := regexp.MustCompile(`(?m)^(.*?(?:class|interface|enum|trait|struct|record|module|namespace|type|protocol|extension)\s+(\w+))`)
	matches := classRe.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		decl := strings.TrimSpace(m[1])
		// Truncate long declarations
		if len(decl) > 100 {
			decl = decl[:97] + "..."
		}
		if summary != "" {
			summary += ". "
		}
		summary += decl
	}

	// Extract function/method declarations
	funcRe := regexp.MustCompile(`(?m)^(\s*(?:public|private|protected|static|def|fun|fn|sub|func|async|sync|export|internal)?\s*(?:function|def|fun|fn|sub|func|macro|proc|get|set)?\s*(?:\w+)?\s*\(?[^)]*\)?\s*[:{]?\s*(?:=>|=|->|where|when)?)`)
	_ = funcRe // reserved for future use

	if summary == "" {
		// Fall back to first meaningful lines
		return summarizeText(content)
	}
	return summary
}

// extractDocComment gets the first block of doc comments from code.
func extractDocComment(content, lineComment, blockComment string) string {
	// Try block comment first
	if blockComment != "" {
		parts := strings.SplitN(blockComment, "|", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			open := regexp.QuoteMeta(parts[0])
			close := regexp.QuoteMeta(parts[1])
			re := regexp.MustCompile(`(?s)` + open + `\s*(.*?)` + close)
			if m := re.FindStringSubmatch(content); len(m) > 1 {
				text := strings.TrimSpace(m[1])
				// Take first paragraph
				if idx := strings.Index(text, "\n\n"); idx > 0 {
					text = text[:idx]
				}
				if len(text) > 300 {
					text = text[:297] + "..."
				}
				return text
			}
		}
	}

	// Try line comments (first block)
	if lineComment != "" {
		quoted := regexp.QuoteMeta(lineComment)
		re := regexp.MustCompile(`(?m)^` + quoted + `\s?(.*)$`)
		matches := re.FindAllStringSubmatch(content, -1)
		if len(matches) > 0 {
			// Take first few comment lines
			var lines []string
			for i, m := range matches {
				if i >= 5 {
					break
				}
				line := strings.TrimSpace(m[1])
				if line != "" {
					lines = append(lines, line)
				}
			}
			if len(lines) > 0 {
				text := strings.Join(lines, ". ")
				if len(text) > 300 {
					text = text[:297] + "..."
				}
				return text
			}
		}
	}

	return ""
}

// --- Language-specific summarizers ---

func summarizeJSON(content string) string {
	return summarizeText(content)
}

func summarizeYAML(content string) string {
	return summarizeText(content)
}

func summarizeTOML(content string) string {
	return summarizeText(content)
}

func summarizeScript(content string) string {
	// Extract shebang and first comments
	lines := strings.SplitN(content, "\n", 10)
	var parts []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#!") {
			parts = append(parts, line[2:])
		} else if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "#") {
			parts = append(parts, strings.TrimPrefix(strings.TrimPrefix(line, "# "), "#"))
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, ". ")
	}
	return summarizeText(content)
}

func summarizeEnv(content string) string {
	lines := strings.Split(content, "\n")
	var keys []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			keys = append(keys, parts[0])
		}
	}
	if len(keys) > 0 {
		result := "Environment variables: " + strings.Join(keys, ", ")
		if len(result) > 300 {
			result = result[:297] + "..."
		}
		return result
	}
	return ""
}

func summarizeDockerfile(content string) string {
	lines := strings.Split(content, "\n")
	var instructions []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Extract instruction (FROM, RUN, COPY, etc.)
		parts := strings.Fields(line)
		if len(parts) > 0 {
			inst := strings.ToUpper(parts[0])
			if inst == "FROM" || inst == "RUN" || inst == "COPY" ||
				inst == "ADD" || inst == "CMD" || inst == "ENTRYPOINT" ||
				inst == "ENV" || inst == "WORKDIR" || inst == "EXPOSE" ||
				inst == "ARG" || inst == "LABEL" || inst == "MAINTAINER" {
				if len(line) > 80 {
					line = line[:77] + "..."
				}
				instructions = append(instructions, line)
			}
		}
	}
	if len(instructions) > 0 {
		return strings.Join(instructions, ". ")
	}
	return summarizeText(content)
}

func summarizeMakefile(content string) string {
	lines := strings.Split(content, "\n")
	var targets []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "\t") ||
			strings.HasPrefix(line, " ") || strings.Contains(line, "=") {
			continue
		}
		// Match target: pattern
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			name := strings.TrimSpace(parts[0])
			if name != "" && !strings.HasPrefix(name, ".") {
				targets = append(targets, name)
			}
		}
	}
	if len(targets) > 0 {
		result := "Targets: " + strings.Join(targets, ", ")
		if len(result) > 300 {
			result = result[:297] + "..."
		}
		return result
	}
	return summarizeText(content)
}
