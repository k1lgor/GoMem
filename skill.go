package gomem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AgentConfig describes where and how to install the skill for a specific agent.
type AgentConfig struct {
	Name        string
	Dir         string
	InstallFile string
	Format      string
	Description string
}

// Agents lists all supported AI coding agents.
var Agents = map[string]AgentConfig{
	"claude":      {Name: "Claude Code", Dir: ".claude/skills/gomem", Format: "skill", Description: "Install for Claude Code (.claude/skills/gomem/)"},
	"claude-code": {Name: "Claude Code", Dir: ".claude/skills/gomem", Format: "skill", Description: "Install for Claude Code (.claude/skills/gomem/)"},
	"pi":          {Name: "Pi", Dir: ".pi/skills/gomem", Format: "skill", Description: "Install for Pi (.pi/skills/gomem/)"},
	"cline":       {Name: "Cline", Dir: ".cline/skills/gomem", Format: "skill", Description: "Install for Cline (.cline/skills/gomem/)"},
	"codex":       {Name: "OpenAI Codex CLI", Dir: ".codex/skills/gomem", Format: "skill", Description: "Install for OpenAI Codex CLI (.codex/skills/gomem/)"},
	"copilot":     {Name: "GitHub Copilot", Dir: ".copilot/skills/gomem", Format: "skill", Description: "Install for GitHub Copilot (.copilot/skills/gomem/)"},
	"cursor":      {Name: "Cursor", Dir: ".cursor/rules/gomem", Format: "skill", Description: "Install for Cursor (.cursor/rules/gomem/)"},
	"windsurf":    {Name: "Windsurf", Dir: ".windsurf/rules/gomem", Format: "skill", Description: "Install for Windsurf (.windsurf/rules/gomem/)"},
	"zed":         {Name: "Zed", Dir: ".zed/skills/gomem", Format: "skill", Description: "Install for Zed (.zed/skills/gomem/)"},
	"kilo":        {Name: "Kilo Code", Dir: ".kilo/skills/gomem", Format: "skill", Description: "Install for Kilo Code (.kilo/skills/gomem/)"},
	"kilo-code":   {Name: "Kilo Code", Dir: ".kilo/skills/gomem", Format: "skill", Description: "Install for Kilo Code (.kilo/skills/gomem/)"},
	"continue":    {Name: "Continue", Dir: ".continue/skills/gomem", Format: "skill", Description: "Install for Continue (.continue/skills/gomem/)"},
	"gemini":      {Name: "Gemini CLI", Dir: ".gemini/skills/gomem", Format: "skill", Description: "Install for Gemini CLI (.gemini/skills/gomem/)"},
	"gemini-cli":  {Name: "Gemini CLI", Dir: ".gemini/skills/gomem", Format: "skill", Description: "Install for Gemini CLI (.gemini/skills/gomem/)"},
}

// WriteAgentsMd creates or updates AGENTS.md at the project root
// so AI coding agents automatically use GoMem before filesystem tools.
func WriteAgentsMd(projectRoot string) {
	path := filepath.Join(projectRoot, "AGENTS.md")

	var existing string
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	}

	if strings.Contains(existing, "## GoMem") {
		return
	}

	content := "## GoMem \u2014 Persistent Memory\n\n" +
		"This project uses GoMem for persistent, searchable memory.\n\n" +
		"**Always use GoMem before filesystem tools.** Follow this order:\n\n" +
		"1. **Search first** \u2014 Before `read`, `ls`, `grep`, `find`, `cat`, or `memo_search`:\n" +
		"   ```\n" +
		"   gomem list                    # See what's in memory\n" +
		"   gomem search \"<topic>\"         # Find relevant context\n" +
		"   ```\n\n" +
		"2. **Filesystem only if needed** \u2014 Only if GoMem returns nothing useful:\n" +
		"   ```\n" +
		"   ls -la\n" +
		"   read <file>\n" +
		"   ```\n\n" +
		"3. **Store what you learn** \u2014 After finding something important:\n" +
		"   ```\n" +
		"   gomem remember <id> \"<concise summary>\"\n" +
		"   ```\n\n" +
		"4. **Index new projects** \u2014 To snapshot the whole project:\n" +
		"   ```\n" +
		"   gomem save-all .\n" +
		"   ```\n\n" +
		"GoMem stores concise structural summaries, not raw file contents.\n" +
		"It saves 10x-100x tokens vs re-reading source files.\n" +
		"Memory persists across context resets and survives compaction.\n\n" +
		"Commands: `gomem remember`, `gomem search`, `gomem list`, `gomem delete`, `gomem save-all`\n"

	if existing != "" {
		content = existing + "\n" + content
	}

	os.WriteFile(path, []byte(content), 0644)
}

// FindSkillSource finds the skills/gomem directory relative to the binary or cwd.
func FindSkillSource() (string, error) {
	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		candidate := filepath.Join(exeDir, "skills", "gomem")
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			return candidate, nil
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine current directory")
	}

	candidate := filepath.Join(cwd, "skills", "gomem")
	if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
		return candidate, nil
	}

	dir := cwd
	for i := 0; i < 5; i++ {
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		candidate := filepath.Join(parent, "skills", "gomem")
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			return candidate, nil
		}
		dir = parent
	}

	return "", fmt.Errorf("cannot find skills/gomem directory. Run from the GoMem project root or place gomem binary in the project directory")
}
