package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	fileReadToolName = "Read"
	globToolName     = "Glob"
	grepToolName     = "Grep"
)

type ToolDefinition struct {
	Name        string
	Description string
	ReadOnly    bool
	Parameters  map[string]any
	Handler     func(context.Context, json.RawMessage) (string, error)
}

type ToolRegistry struct {
	byName map[string]ToolDefinition
	order  []ToolDefinition
}

func NewDefaultToolRegistry() *ToolRegistry {
	tools := []ToolDefinition{
		newFileReadTool(),
		newGlobTool(),
		newGrepTool(),
	}
	byName := make(map[string]ToolDefinition, len(tools))
	for _, tool := range tools {
		byName[tool.Name] = tool
	}
	return &ToolRegistry{byName: byName, order: tools}
}

func (r *ToolRegistry) OpenAITools() []map[string]any {
	out := make([]map[string]any, 0, len(r.order))
	for _, t := range r.order {
		out = append(out, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  t.Parameters,
			},
		})
	}
	return out
}

func (r *ToolRegistry) Execute(ctx context.Context, call APIToolCall) (string, error) {
	def, ok := r.byName[call.Function.Name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", call.Function.Name)
	}
	return def.Handler(ctx, json.RawMessage(call.Function.Arguments))
}

type fileReadInput struct {
	FilePath string `json:"file_path"`
	Offset   *int   `json:"offset,omitempty"`
	Limit    *int   `json:"limit,omitempty"`
}

func newFileReadTool() ToolDefinition {
	return ToolDefinition{
		Name:        fileReadToolName,
		ReadOnly:    true,
		Description: "Read a file from the local filesystem.",
		Parameters: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"file_path"},
			"properties": map[string]any{
				"file_path": map[string]any{"type": "string", "description": "The absolute path to the file to read"},
				"offset":    map[string]any{"type": "integer", "minimum": 0, "description": "The line number to start reading from. Only provide if the file is too large to read at once"},
				"limit":     map[string]any{"type": "integer", "minimum": 1, "description": "The number of lines to read. Only provide if the file is too large to read at once."},
			},
		},
		Handler: runFileRead,
	}
}

func runFileRead(_ context.Context, raw json.RawMessage) (string, error) {
	var in fileReadInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if strings.TrimSpace(in.FilePath) == "" {
		return "", errors.New("file_path is required")
	}
	if !filepath.IsAbs(in.FilePath) {
		return "", fmt.Errorf("file_path must be absolute, got: %s", in.FilePath)
	}
	if in.Offset != nil && *in.Offset < 0 {
		return "", errors.New("offset must be >= 0")
	}
	if in.Limit != nil && *in.Limit <= 0 {
		return "", errors.New("limit must be > 0")
	}

	b, err := os.ReadFile(in.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
	offset := 0
	if in.Offset != nil {
		offset = *in.Offset
	}
	if offset > len(lines) {
		offset = len(lines)
	}
	end := len(lines)
	if in.Limit != nil {
		end = offset + *in.Limit
		if end > len(lines) {
			end = len(lines)
		}
	}

	selected := lines[offset:end]
	if len(selected) == 0 {
		return "", nil
	}

	var sb strings.Builder
	for i, line := range selected {
		lineNum := offset + i + 1
		sb.WriteString(strconv.Itoa(lineNum))
		sb.WriteString("\t")
		sb.WriteString(line)
		if i < len(selected)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String(), nil
}

type globInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
}

func newGlobTool() ToolDefinition {
	return ToolDefinition{
		Name:     globToolName,
		ReadOnly: true,
		Description: `- Fast file pattern matching tool that works with any codebase size
- Supports glob patterns like "**/*.js" or "src/**/*.ts"
- Returns matching file paths sorted by modification time
- Use this tool when you need to find files by name patterns
- When you are doing an open ended search that may require multiple rounds of globbing and grepping, use the Agent tool instead`,
		Parameters: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"pattern"},
			"properties": map[string]any{
				"pattern": map[string]any{"type": "string", "description": "The glob pattern to match files against"},
				"path":    map[string]any{"type": "string", "description": "The directory to search in. If not specified, the current working directory will be used."},
			},
		},
		Handler: runGlob,
	}
}

func runGlob(_ context.Context, raw json.RawMessage) (string, error) {
	var in globInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if strings.TrimSpace(in.Pattern) == "" {
		return "", errors.New("pattern is required")
	}
	base := in.Path
	if strings.TrimSpace(base) == "" {
		cwd, _ := os.Getwd()
		base = cwd
	}
	if !filepath.IsAbs(base) {
		return "", fmt.Errorf("path must be absolute, got: %s", base)
	}
	st, err := os.Stat(base)
	if err != nil {
		return "", fmt.Errorf("path does not exist: %w", err)
	}
	if !st.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", base)
	}

	matches, err := globSearch(base, in.Pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "No files found", nil
	}
	if len(matches) > 100 {
		matches = matches[:100]
	}
	return strings.Join(matches, "\n"), nil
}

func globSearch(base string, pattern string) ([]string, error) {
	type result struct {
		path  string
		mtime int64
	}
	results := make([]result, 0, 64)
	err := filepath.WalkDir(base, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		rel, err := filepath.Rel(base, p)
		if err != nil {
			return nil
		}
		if rel == "." {
			return nil
		}
		if matched, _ := filepath.Match(pattern, rel); matched {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			results = append(results, result{path: filepath.ToSlash(rel), mtime: info.ModTime().UnixNano()})
		}
		if strings.Contains(pattern, "**") {
			sub := strings.ReplaceAll(filepath.ToSlash(rel), "/", string(filepath.Separator)+"*")
			if matched, _ := filepath.Match(strings.ReplaceAll(pattern, "**", "*"), sub); matched {
				info, err := d.Info()
				if err != nil {
					return nil
				}
				results = append(results, result{path: filepath.ToSlash(rel), mtime: info.ModTime().UnixNano()})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].mtime == results[j].mtime {
			return results[i].path < results[j].path
		}
		return results[i].mtime > results[j].mtime
	})
	uniq := make([]string, 0, len(results))
	seen := map[string]struct{}{}
	for _, r := range results {
		if _, ok := seen[r.path]; ok {
			continue
		}
		seen[r.path] = struct{}{}
		uniq = append(uniq, r.path)
	}
	return uniq, nil
}

type grepInput struct {
	Pattern    string `json:"pattern"`
	Path       string `json:"path,omitempty"`
	Glob       string `json:"glob,omitempty"`
	OutputMode string `json:"output_mode,omitempty"`
	IgnoreCase bool   `json:"-i,omitempty"`
	LineNum    *bool  `json:"-n,omitempty"`
	HeadLimit  *int   `json:"head_limit,omitempty"`
	Offset     *int   `json:"offset,omitempty"`
	Multiline  bool   `json:"multiline,omitempty"`
}

func newGrepTool() ToolDefinition {
	return ToolDefinition{
		Name:     grepToolName,
		ReadOnly: true,
		Description: `A powerful search tool built on ripgrep

Usage:
- ALWAYS use Grep for search tasks. NEVER invoke grep or rg as a Bash command. The Grep tool has been optimized for correct permissions and access.
- Supports full regex syntax (e.g., "log.*Error", "function\\s+\\w+")
- Filter files with glob parameter (e.g. "*.js", "**/*.tsx")
- Output modes: "content" shows matching lines, "files_with_matches" shows only file paths (default), "count" shows match counts`,
		Parameters: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"pattern"},
			"properties": map[string]any{
				"pattern":     map[string]any{"type": "string", "description": "The regular expression pattern to search for in file contents"},
				"path":        map[string]any{"type": "string", "description": "File or directory to search in. Defaults to current working directory."},
				"glob":        map[string]any{"type": "string", "description": "Glob pattern to filter files (e.g. *.js, *.{ts,tsx})"},
				"output_mode": map[string]any{"type": "string", "enum": []string{"content", "files_with_matches", "count"}, "description": "Output mode. Defaults to files_with_matches."},
				"-i":          map[string]any{"type": "boolean", "description": "Case insensitive search"},
				"-n":          map[string]any{"type": "boolean", "description": "Show line numbers for content mode"},
				"head_limit":  map[string]any{"type": "integer", "description": "Limit output to first N entries"},
				"offset":      map[string]any{"type": "integer", "description": "Skip first N entries before head_limit"},
				"multiline":   map[string]any{"type": "boolean", "description": "Enable multiline mode"},
			},
		},
		Handler: runGrep,
	}
}

func runGrep(_ context.Context, raw json.RawMessage) (string, error) {
	var in grepInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if strings.TrimSpace(in.Pattern) == "" {
		return "", errors.New("pattern is required")
	}
	base := in.Path
	if strings.TrimSpace(base) == "" {
		cwd, _ := os.Getwd()
		base = cwd
	}
	if !filepath.IsAbs(base) {
		return "", fmt.Errorf("path must be absolute, got: %s", base)
	}
	if _, err := os.Stat(base); err != nil {
		return "", fmt.Errorf("path does not exist: %w", err)
	}

	mode := in.OutputMode
	if mode == "" {
		mode = "files_with_matches"
	}
	if mode != "content" && mode != "files_with_matches" && mode != "count" {
		return "", fmt.Errorf("invalid output_mode: %s", mode)
	}

	args := []string{"--hidden", "--max-columns", "500"}
	if in.Multiline {
		args = append(args, "-U", "--multiline-dotall")
	}
	if in.IgnoreCase {
		args = append(args, "-i")
	}
	switch mode {
	case "files_with_matches":
		args = append(args, "-l")
	case "count":
		args = append(args, "-c")
	case "content":
		showLine := true
		if in.LineNum != nil {
			showLine = *in.LineNum
		}
		if showLine {
			args = append(args, "-n")
		}
	}
	if strings.HasPrefix(in.Pattern, "-") {
		args = append(args, "-e", in.Pattern)
	} else {
		args = append(args, in.Pattern)
	}
	if strings.TrimSpace(in.Glob) != "" {
		for _, g := range strings.Fields(in.Glob) {
			args = append(args, "--glob", g)
		}
	}
	args = append(args, base)

	output, err := exec.Command("rg", args...).CombinedOutput()
	if err != nil {
		var ex *exec.ExitError
		if errors.As(err, &ex) && ex.ExitCode() == 1 {
			return noResultsForMode(mode), nil
		}
		if errors.As(err, &ex) && ex.ExitCode() == 2 {
			return "", fmt.Errorf("grep failed: %s", strings.TrimSpace(string(output)))
		}
		fallback, fbErr := grepFallback(base, in.Pattern, in.IgnoreCase, mode)
		if fbErr == nil {
			return fallback, nil
		}
		return "", fmt.Errorf("grep execution failed: %v (%s)", err, strings.TrimSpace(string(output)))
	}
	lines := splitNonEmptyLines(string(output))
	offset := 0
	if in.Offset != nil && *in.Offset > 0 {
		offset = *in.Offset
	}
	limit := 250
	if in.HeadLimit != nil {
		limit = *in.HeadLimit
	}
	if offset > len(lines) {
		lines = nil
	} else {
		lines = lines[offset:]
	}
	if limit > 0 && len(lines) > limit {
		lines = lines[:limit]
	}
	if len(lines) == 0 {
		return noResultsForMode(mode), nil
	}
	return strings.Join(lines, "\n"), nil
}

func noResultsForMode(mode string) string {
	if mode == "files_with_matches" {
		return "No files found"
	}
	return "No matches found"
}

func splitNonEmptyLines(raw string) []string {
	parts := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func grepFallback(base string, pattern string, ignoreCase bool, mode string) (string, error) {
	reFlags := ""
	if ignoreCase {
		reFlags = "(?i)"
	}
	re, err := regexp.Compile(reFlags + pattern)
	if err != nil {
		return "", err
	}
	type countEntry struct {
		file  string
		count int
	}
	content := make([]string, 0, 128)
	files := make([]string, 0, 64)
	counts := make([]countEntry, 0, 64)
	err = filepath.WalkDir(base, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		txt := strings.ReplaceAll(string(b), "\r\n", "\n")
		rel, _ := filepath.Rel(base, path)
		rel = filepath.ToSlash(rel)
		hits := 0
		for i, line := range strings.Split(txt, "\n") {
			if re.MatchString(line) {
				hits++
				if mode == "content" {
					content = append(content, fmt.Sprintf("%s:%d:%s", rel, i+1, line))
				}
			}
		}
		if hits > 0 {
			files = append(files, rel)
			counts = append(counts, countEntry{file: rel, count: hits})
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	switch mode {
	case "content":
		if len(content) == 0 {
			return "No matches found", nil
		}
		return strings.Join(content, "\n"), nil
	case "count":
		if len(counts) == 0 {
			return "No matches found", nil
		}
		rows := make([]string, 0, len(counts))
		for _, c := range counts {
			rows = append(rows, fmt.Sprintf("%s:%d", c.file, c.count))
		}
		return strings.Join(rows, "\n"), nil
	default:
		if len(files) == 0 {
			return "No files found", nil
		}
		return strings.Join(files, "\n"), nil
	}
}
