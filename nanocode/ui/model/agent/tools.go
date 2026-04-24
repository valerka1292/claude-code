package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

const (
	fileReadToolName  = "Read"
	fileWriteToolName = "Write"
	globToolName      = "Glob"
	grepToolName      = "Grep"
)

const (
	fileReadMaxLines = 2000
	fileReadMaxBytes = 256 * 1024 // 256 KB
)

type ToolDefinition struct {
	Name        string
	Description string
	ReadOnly   bool
	Parameters map[string]any
	Handler    func(context.Context, json.RawMessage) (string, error)
}

type ToolRegistry struct {
	byName map[string]ToolDefinition
	order  []ToolDefinition

	readState   map[string]time.Time
	readStateMu sync.RWMutex
}

func NewDefaultToolRegistry() *ToolRegistry {
	r := &ToolRegistry{
		readState: make(map[string]time.Time),
	}

	tools := []ToolDefinition{
		r.newFileReadTool(),
		r.newFileWriteTool(),
		r.newGlobTool(),
		r.newGrepTool(),
	}

	byName := make(map[string]ToolDefinition, len(tools))
	for _, tool := range tools {
		byName[tool.Name] = tool
	}
	r.byName = byName
	r.order = tools

	return r
}

func (r *ToolRegistry) GetAllTools() []ToolDefinition {
	return r.order
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

func (r *ToolRegistry) Execute(ctx context.Context, call APIToolCall, readOnly bool) (string, error) {
	def, ok := r.byName[call.Function.Name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", call.Function.Name)
	}

	if readOnly && !def.ReadOnly {
		return "", fmt.Errorf("ERROR: Mode 'ask' (read-only) is active. Tool '%s' is BLOCKED. Do not retry this tool. Politely ask the user to press 'shift+tab' to switch to 'code' mode if writing is required.", def.Name)
	}

	return def.Handler(ctx, json.RawMessage(call.Function.Arguments))
}

type fileReadInput struct {
	FilePath string `json:"file_path"`
	Offset *int   `json:"offset,omitempty"`
	Limit  *int   `json:"limit,omitempty"`
}

func (r *ToolRegistry) newFileReadTool() ToolDefinition {
	return ToolDefinition{
		Name:     fileReadToolName,
		ReadOnly: true,
		Description: `Reads a file from the local filesystem. You can access any file directly by using this tool.
Assume this tool is able to read all files on the machine. If the User provides a path to a file assume that path is valid. It is okay to read a file that does not exist; an error will be returned.

Usage:
- The file_path parameter must be an absolute path, not a relative path
- By default, it reads up to 2000 lines starting from the beginning of the file. Files larger than 256 KB will return an error; use offset and limit for larger files
- You can optionally specify a line offset and limit (especially handy for long files), but it's recommended to read the whole file by not providing these parameters
- Results are returned using cat -n format, with line numbers starting at 1
- This tool can only read files, not directories. To read a directory, use an ls command via the Bash tool.
- You will regularly be asked to read screenshots. If the user provides a path to a screenshot, ALWAYS use this tool to view the file at the path. This tool will work with all temporary file paths.
- If you read a file that exists but has empty contents you will receive a system reminder warning in place of file contents.`,
		Parameters: map[string]any{
			"type":                 "object",
			"additionalProperties":   false,
			"required":             []string{"file_path"},
			"properties": map[string]any{
				"file_path": map[string]any{"type": "string", "description": "The absolute path to the file to read"},
				"offset":  map[string]any{"type": "integer", "minimum": 1, "description": "The line number to start reading from. Only provide if the file is too large to read at once"},
				"limit":   map[string]any{"type": "integer", "minimum": 1, "description": "The number of lines to read. Only provide if the file is too large to read at once."},
			},
		},
		Handler: r.runFileRead,
	}
}

func (r *ToolRegistry) runFileRead(_ context.Context, raw json.RawMessage) (string, error) {
	var in fileReadInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if strings.TrimSpace(in.FilePath) == "" {
		return "", errors.New("file_path is required")
	}

	absPath := in.FilePath
	if !filepath.IsAbs(absPath) {
		cwd, _ := os.Getwd()
		absPath = filepath.Join(cwd, absPath)
	}

	if isBlockedDevicePath(absPath) {
		return "", fmt.Errorf("Cannot read '%s': this device file would block or produce infinite output.", in.FilePath)
	}

	info, err := os.Lstat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			cwd, _ := os.Getwd()
			return "", fmt.Errorf("File does not exist. Note: The current working directory is %s.", cwd)
		}
		return "", fmt.Errorf("%v", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("This tool can only read files, not directories. To read a directory, use an ls command via the Bash tool.")
	}

	offsetLine := 1
	if in.Offset != nil && *in.Offset > 0 {
		offsetLine = *in.Offset
	}

	limitLines := 0
	if in.Limit != nil {
		limitLines = *in.Limit
		if limitLines <= 0 {
			return "", errors.New("limit must be > 0")
		}
	}

	if limitLines == 0 && info.Size() > int64(fileReadMaxBytes) {
		return "", fmt.Errorf("File content (%d bytes) exceeds maximum allowed size (%d bytes). Use offset and limit parameters to read specific portions of the file.", info.Size(), fileReadMaxBytes)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("Error opening file: %v", err)
	}
	defer file.Close()

	head := make([]byte, 512)
	n, _ := file.Read(head)
	if isBinary(head[:n]) {
		return "", fmt.Errorf("This tool cannot read binary files. The file appears to be a binary file. Please use appropriate tools for binary file analysis.")
	}

	file.Seek(0, io.SeekStart)

	reader := bufio.NewReader(file)
	var sb strings.Builder
	lineNum := 1
	linesRead := 0

	for {
		line, err := reader.ReadString('\n')

		if lineNum >= offsetLine && (line != "" || err == nil) {
			sb.WriteString(fmt.Sprintf("%d\t%s", lineNum, line))

			if !strings.HasSuffix(line, "\n") {
				sb.WriteString("\n")
			}

			linesRead++
			if limitLines > 0 && linesRead >= limitLines {
				break
			}
			if limitLines == 0 && linesRead >= fileReadMaxLines {
				break
			}
		}

		lineNum++
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("error reading file: %v", err)
		}
	}

	if sb.Len() == 0 {
		if info.Size() == 0 {
			r.markAsRead(absPath, info.ModTime())
			return "<system-reminder>Warning: the file exists but the contents are empty.</system-reminder>", nil
		}
		return fmt.Sprintf("<system-reminder>Warning: the file exists but is shorter than the provided offset (%d).</system-reminder>", offsetLine), nil
	}

	r.markAsRead(absPath, info.ModTime())
	return sb.String(), nil
}

func isBlockedDevicePath(filePath string) bool {
	blocked := map[string]bool{
		"/dev/zero":     true,
		"/dev/random":   true,
		"/dev/urandom": true,
		"/dev/full":     true,
		"/dev/stdin":    true,
		"/dev/tty":    true,
		"/dev/console": true,
		"/dev/stdout":  true,
		"/dev/stderr":  true,
		"/dev/fd/0":   true,
		"/dev/fd/1":   true,
		"/dev/fd/2":   true,
	}
	if blocked[filePath] {
		return true
	}
	if strings.HasPrefix(filePath, "/proc/") {
		if strings.HasSuffix(filePath, "/fd/0") || strings.HasSuffix(filePath, "/fd/1") || strings.HasSuffix(filePath, "/fd/2") {
			return true
		}
	}
	return false
}

func isBinary(data []byte) bool {
	for i := 0; i < len(data); i++ {
		if data[i] == 0 {
			return true
		}
	}
	return !utf8.Valid(data)
}

func (r *ToolRegistry) markAsRead(path string, mtime time.Time) {
	r.readStateMu.Lock()
	defer r.readStateMu.Unlock()
	r.readState[path] = mtime
}

type fileWriteInput struct {
	FilePath string `json:"file_path"`
	Content string `json:"content"`
}

func (r *ToolRegistry) newFileWriteTool() ToolDefinition {
	return ToolDefinition{
		Name:     fileWriteToolName,
		ReadOnly: false,
		Description: `Writes a file to the local filesystem.

Usage:
- This tool will overwrite the existing file if there is one at the provided path.
- If this is an existing file, you MUST use the Read tool first to read the file's contents. This tool will fail if you did not read the file first.
- Prefer the Edit tool for modifying existing files — it only sends the diff. Only use this tool to create new files or for complete rewrites.
- NEVER create documentation files (*.md) or README files unless explicitly requested by the User.
- Only use emojis if the user explicitly requests it. Avoid writing emojis to files unless asked.`,
		Parameters: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"file_path", "content"},
			"properties": map[string]any{
				"file_path": map[string]any{"type": "string", "description": "The absolute path to the file to write (must be absolute, not relative)"},
				"content": map[string]any{"type": "string", "description": "The content to write to the file"},
			},
		},
		Handler: r.runFileWrite,
	}
}

func (r *ToolRegistry) runFileWrite(_ context.Context, raw json.RawMessage) (string, error) {
	var in fileWriteInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	absPath := in.FilePath
	if !filepath.IsAbs(absPath) {
		cwd, _ := os.Getwd()
		absPath = filepath.Join(cwd, absPath)
	}

	if strings.HasPrefix(absPath, `\\`) || strings.HasPrefix(absPath, `//`) {
		return "Write operation skipped for UNC paths due to security restrictions.", nil
	}

	isCreate := false
	fileStat, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			isCreate = true
		} else {
			return "", err
		}
	}

	var oldContent string

	if !isCreate {
		r.readStateMu.RLock()
		readTime, hasRead := r.readState[absPath]
		r.readStateMu.RUnlock()

		if !hasRead {
			return "", errors.New("File has not been read yet. Read it first before writing to it.")
		}

		if fileStat.ModTime().Unix() > readTime.Unix() {
			return "", errors.New("File has been modified since read, either by the user or by a linter. Read it again before attempting to write it.")
		}

		oldBytes, _ := os.ReadFile(absPath)
		oldContent = string(oldBytes)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directories: %w", err)
	}

	if err := os.WriteFile(absPath, []byte(in.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	newStat, _ := os.Stat(absPath)
	if newStat != nil {
		r.markAsRead(absPath, newStat.ModTime())
	}

	lines := strings.Count(in.Content, "\n")
	if len(in.Content) > 0 && !strings.HasSuffix(in.Content, "\n") {
		lines++
	}

	var diffText string
	if !isCreate {
		edits := myers.ComputeEdits(span.URIFromPath(absPath), oldContent, in.Content)
		diffText = fmt.Sprintf("%s", gotextdiff.ToUnified(filepath.Base(absPath), filepath.Base(absPath), oldContent, edits))
	} else {
		diffText = "@@ -0,0 +1," + strconv.Itoa(lines) + " @@\n"
		for _, line := range strings.Split(in.Content, "\n") {
			diffText += "+" + line + "\n"
		}
	}

	diffJSON, _ := json.Marshal(diffText)

	if isCreate {
		return fmt.Sprintf(`{"type":"create","filePath":%q,"lines":%d,"diff":%s}`, absPath, lines, string(diffJSON)), nil
	}
	return fmt.Sprintf(`{"type":"update","filePath":%q,"lines":%d,"diff":%s}`, absPath, lines, string(diffJSON)), nil
}

type globInput struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
}

func (r *ToolRegistry) newGlobTool() ToolDefinition {
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
				"path": map[string]any{
					"type":        "string",
					"description": "The directory to search in. If not specified, the current working directory will be used. IMPORTANT: Omit this field to use the default directory. DO NOT enter \"undefined\" or \"null\" - simply omit it for the default behavior. Must be a valid directory path if provided.",
				},
			},
		},
		Handler: r.runGlob,
	}
}

func (r *ToolRegistry) runGlob(_ context.Context, raw json.RawMessage) (string, error) {
	var in globInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if strings.TrimSpace(in.Pattern) == "" {
		return "", errors.New("pattern is required")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	basePath := in.Path
	if strings.TrimSpace(basePath) == "" {
		basePath = cwd
	}
	if !filepath.IsAbs(basePath) {
		basePath = filepath.Join(cwd, basePath)
	}

	if strings.HasPrefix(basePath, `\\`) || strings.HasPrefix(basePath, `//`) {
		return "No files found", nil
	}

	st, err := os.Stat(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("Directory does not exist: %s. Note: The current working directory is %s.", in.Path, cwd)
		}
		return "", err
	}
	if !st.IsDir() {
		return "", fmt.Errorf("Path is not a directory: %s", in.Path)
	}

	matches, err := globSearch(basePath, in.Pattern)
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return "No files found", nil
	}

	truncated := false
	limit := 100
	if len(matches) > limit {
		matches = matches[:limit]
		truncated = true
	}

	relMatches := make([]string, 0, len(matches))
	for _, m := range matches {
		rel, err := filepath.Rel(cwd, m)
		if err == nil {
			relMatches = append(relMatches, filepath.ToSlash(rel))
		} else {
			relMatches = append(relMatches, filepath.ToSlash(m))
		}
	}

	var sb strings.Builder
	sb.WriteString(strings.Join(relMatches, "\n"))
	if truncated {
		sb.WriteString("\n(Results are truncated. Consider using a more specific path or pattern.)")
	}

	return sb.String(), nil
}

func globSearch(base string, pattern string) ([]string, error) {
	fsys := os.DirFS(base)

	matches, err := doublestar.Glob(fsys, pattern)
	if err != nil {
		return nil, err
	}

	type fileInfo struct {
		path  string
		mtime int64
	}

	var files []fileInfo
	for _, match := range matches {
		fullPath := filepath.Join(base, match)
		info, err := os.Stat(fullPath)
		if err == nil && !info.IsDir() {
			files = append(files, fileInfo{path: fullPath, mtime: info.ModTime().UnixNano()})
		}
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].mtime == files[j].mtime {
			return files[i].path < files[j].path
		}
		return files[i].mtime > files[j].mtime
	})

	var results []string
	for _, f := range files {
		results = append(results, f.path)
	}

	return results, nil
}

type grepInput struct {
	Pattern    string `json:"pattern"`
	Path       string `json:"path,omitempty"`
	Glob       string `json:"glob,omitempty"`
	OutputMode string `json:"output_mode,omitempty"`
	ContextB   *int   `json:"-B,omitempty"`
	ContextA   *int   `json:"-A,omitempty"`
	ContextC1  *int   `json:"-C,omitempty"`
	Context    *int   `json:"context,omitempty"`
	LineNum    *bool  `json:"-n,omitempty"`
	IgnoreCase *bool  `json:"-i,omitempty"`
	Type       string `json:"type,omitempty"`
	HeadLimit  *int   `json:"head_limit,omitempty"`
	Offset     *int   `json:"offset,omitempty"`
	Multiline  *bool  `json:"multiline,omitempty"`
}

const defaultHeadLimit = 250

func (r *ToolRegistry) newGrepTool() ToolDefinition {
	return ToolDefinition{
		Name:     grepToolName,
		ReadOnly: true,
		Description: `A powerful search tool built on ripgrep

Usage:
- ALWAYS use Grep for search tasks. NEVER invoke grep or rg as a Bash command. The Grep tool has been optimized for correct permissions and access.
- Supports full regex syntax (e.g., "log.*Error", "function\\s+\w+")
- Filter files with glob parameter (e.g. "*.js", "*.{ts,tsx}") or type parameter (e.g., "js", "py", "rust")
- Output modes: "content" shows matching lines, "files_with_matches" shows only file paths (default), "count" shows match counts
- Use Agent tool for open-ended searches requiring multiple rounds
- Pattern syntax: Uses ripgrep (not grep) - literal braces need escaping (use Backslash + Backtick interface OpenBrace CloseBrace Backtick to find interface{} in Go code)
- Multiline matching: By default patterns match within single lines only. For cross-line patterns, use multiline: true`,
		Parameters: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"required":             []string{"pattern"},
			"properties": map[string]any{
				"pattern":     map[string]any{"type": "string", "description": "The regular expression pattern to search for in file contents"},
				"path":        map[string]any{"type": "string", "description": "File or directory to search in (rg PATH). Defaults to current working directory."},
				"glob":        map[string]any{"type": "string", "description": "Glob pattern to filter files (e.g. \"*.js\", \"*.{ts,tsx}\") - maps to rg --glob"},
				"output_mode": map[string]any{"type": "string", "enum": []string{"content", "files_with_matches", "count"}, "description": "Output mode: \"content\" shows matching lines, \"files_with_matches\" shows file paths, \"count\" shows match counts. Defaults to \"files_with_matches\"."},
				"-B":          map[string]any{"type": "integer", "description": "Number of lines to show before each match (rg -B)."},
				"-A":          map[string]any{"type": "integer", "description": "Number of lines to show after each match (rg -A)."},
				"-C":          map[string]any{"type": "integer", "description": "Alias for context."},
				"context":     map[string]any{"type": "integer", "description": "Number of lines to show before and after each match (rg -C)."},
				"-n":          map[string]any{"type": "boolean", "description": "Show line numbers in output (rg -n). Defaults to true."},
				"-i":          map[string]any{"type": "boolean", "description": "Case insensitive search (rg -i)"},
				"type":        map[string]any{"type": "string", "description": "File type to search (rg --type). Common types: js, py, rust, go, java, etc."},
				"head_limit":  map[string]any{"type": "integer", "description": "Limit output to first N lines/entries. Defaults to 250. Pass 0 for unlimited."},
				"offset":      map[string]any{"type": "integer", "description": "Skip first N lines/entries before applying head_limit. Defaults to 0."},
				"multiline":   map[string]any{"type": "boolean", "description": "Enable multiline mode. Default: false."},
			},
		},
		Handler: r.runGrep,
	}
}

func (r *ToolRegistry) runGrep(_ context.Context, raw json.RawMessage) (string, error) {
	var in grepInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if strings.TrimSpace(in.Pattern) == "" {
		return "", errors.New("pattern is required")
	}

	cwd, _ := os.Getwd()
	absolutePath := in.Path
	if absolutePath == "" {
		absolutePath = cwd
	} else if !filepath.IsAbs(absolutePath) {
		absolutePath = filepath.Join(cwd, absolutePath)
	}

	if strings.HasPrefix(absolutePath, `\\`) || strings.HasPrefix(absolutePath, `//`) {
		return "No files found", nil
	}

	if _, err := os.Stat(absolutePath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("Path does not exist: %s. Note: The current working directory is %s.", in.Path, cwd)
		}
		return "", err
	}

	mode := in.OutputMode
	if mode == "" {
		mode = "files_with_matches"
	}

	args := []string{"--hidden"}

	vcsDirs := []string{".git", ".svn", ".hg", ".bzr", ".jj", ".sl"}
	for _, dir := range vcsDirs {
		args = append(args, "--glob", "!"+dir)
	}

	args = append(args, "--max-columns", "500")

	if in.Multiline != nil && *in.Multiline {
		args = append(args, "-U", "--multiline-dotall")
	}
	if in.IgnoreCase != nil && *in.IgnoreCase {
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
		if in.Context != nil {
			args = append(args, "-C", strconv.Itoa(*in.Context))
		} else if in.ContextC1 != nil {
			args = append(args, "-C", strconv.Itoa(*in.ContextC1))
		} else {
			if in.ContextB != nil {
				args = append(args, "-B", strconv.Itoa(*in.ContextB))
			}
			if in.ContextA != nil {
				args = append(args, "-A", strconv.Itoa(*in.ContextA))
			}
		}
	}

	if strings.HasPrefix(in.Pattern, "-") {
		args = append(args, "-e", in.Pattern)
	} else {
		args = append(args, in.Pattern)
	}

	if in.Type != "" {
		args = append(args, "--type", in.Type)
	}

	if strings.TrimSpace(in.Glob) != "" {
		for _, g := range splitGlobPatterns(in.Glob) {
			args = append(args, "--glob", g)
		}
	}

	args = append(args, absolutePath)

	output, err := exec.Command("rg", args...).CombinedOutput()
	if err != nil {
		var ex *exec.ExitError
		if errors.As(err, &ex) && ex.ExitCode() == 1 {
			if mode == "files_with_matches" {
				return "No files found", nil
			}
			return "No matches found", nil
		}
		return "", fmt.Errorf("grep execution failed: %v\nMake sure 'rg' (ripgrep) is installed.", err)
	}

	rawLines := splitNonEmptyLines(string(output))

	offsetVal := 0
	if in.Offset != nil && *in.Offset > 0 {
		offsetVal = *in.Offset
	}

	limitedResults, appliedLimit := applyHeadLimit(rawLines, in.HeadLimit, offsetVal)
	limitInfo := formatLimitInfo(appliedLimit, offsetVal)

	switch mode {
	case "content":
		if len(limitedResults) == 0 {
			return "No matches found", nil
		}
		for i, line := range limitedResults {
			colonIdx := strings.Index(line, ":")
			if colonIdx > 0 {
				relPath, _ := filepath.Rel(cwd, line[:colonIdx])
				limitedResults[i] = filepath.ToSlash(relPath) + line[colonIdx:]
			}
		}
		resultContent := strings.Join(limitedResults, "\n")
		if limitInfo != "" {
			resultContent += fmt.Sprintf("\n\n[Showing results with pagination = %s]", limitInfo)
		}
		return resultContent, nil

	case "count":
		totalMatches := 0
		fileCount := 0
		for i, line := range limitedResults {
			colonIdx := strings.LastIndex(line, ":")
			if colonIdx > 0 {
				relPath, _ := filepath.Rel(cwd, line[:colonIdx])
				limitedResults[i] = filepath.ToSlash(relPath) + line[colonIdx:]
				count, err := strconv.Atoi(strings.TrimSpace(line[colonIdx+1:]))
				if err == nil {
					totalMatches += count
					fileCount++
				}
			}
		}

		summary := fmt.Sprintf("\n\nFound %d total occurrences across %d files.", totalMatches, fileCount)
		if limitInfo != "" {
			summary += fmt.Sprintf(" with pagination = %s", limitInfo)
		}
		return strings.Join(limitedResults, "\n") + summary, nil

	default:
		if len(limitedResults) == 0 {
			return "No files found", nil
		}
		for i, line := range limitedResults {
			relPath, _ := filepath.Rel(cwd, line)
			limitedResults[i] = filepath.ToSlash(relPath)
		}

		header := fmt.Sprintf("Found %d files", len(limitedResults))
		if limitInfo != "" {
			header += fmt.Sprintf(" %s", limitInfo)
		}
		return header + "\n" + strings.Join(limitedResults, "\n"), nil
	}
}

func applyHeadLimit(lines []string, limit *int, offset int) ([]string, *int) {
	if limit != nil && *limit == 0 {
		if offset < len(lines) {
			return lines[offset:], nil
		}
		return nil, nil
	}

	effectiveLimit := defaultHeadLimit
	if limit != nil {
		effectiveLimit = *limit
	}

	if offset >= len(lines) {
		return nil, nil
	}

	sliced := lines[offset:]
	wasTruncated := len(sliced) > effectiveLimit
	var appliedLimit *int

	if wasTruncated {
		sliced = sliced[:effectiveLimit]
		appliedLimit = &effectiveLimit
	}
	return sliced, appliedLimit
}

func formatLimitInfo(appliedLimit *int, offset int) string {
	var parts []string
	if appliedLimit != nil {
		parts = append(parts, fmt.Sprintf("limit: %d", *appliedLimit))
	}
	if offset > 0 {
		parts = append(parts, fmt.Sprintf("offset: %d", offset))
	}
	return strings.Join(parts, ", ")
}

func splitGlobPatterns(globRaw string) []string {
	fields := strings.Fields(globRaw)
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		if strings.Contains(field, "{") && strings.Contains(field, "}") {
			out = append(out, field)
			continue
		}
		for _, chunk := range strings.Split(field, ",") {
			chunk = strings.TrimSpace(chunk)
			if chunk != "" {
				out = append(out, chunk)
			}
		}
	}
	return out
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