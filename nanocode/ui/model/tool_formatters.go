package model

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"nanocode/ui/components/messages"
)

type toolArgsFormatter func(args map[string]any, cwd string) string
type toolResultFormatter func(raw string, width int) string

var toolArgsFormatters = map[string]toolArgsFormatter{
	"Write": formatWriteArgs,
	"Read":  formatReadArgs,
	"Glob":  formatGlobArgs,
	"Grep":  formatGrepArgs,
}

var toolResultFormatters = map[string]toolResultFormatter{
	"Write": formatWriteResult,
	"Edit":  formatEditResult,
	"Read":  formatReadResult,
	"Glob":  formatSearchResult,
	"Grep":  formatSearchResult,
}

func formatToolArgs(name string, raw string, cwd string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	var args map[string]any
	if err := json.Unmarshal([]byte(trimmed), &args); err != nil {
		return clampString(trimmed, 70)
	}

	formatter, ok := toolArgsFormatters[name]
	if !ok {
		return formatDefaultArgs(args)
	}
	return formatter(args, cwd)
}

func formatToolResult(name string, raw string, isErr bool, width int) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "empty result"
	}

	if isErr {
		firstLine := strings.Split(trimmed, "\n")[0]
		firstLine = strings.TrimPrefix(firstLine, "tool execution failed: ")
		return clampString(firstLine, 120)
	}

	formatter, ok := toolResultFormatters[name]
	if !ok {
		return formatDefaultResult(trimmed)
	}
	return formatter(trimmed, width)
}

func formatWriteArgs(args map[string]any, cwd string) string {
	path, _ := args["file_path"].(string)
	content, _ := args["content"].(string)
	base := relativePath(cwd, path)
	lines := strings.Count(content, "\n")
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		lines++
	}
	return fmt.Sprintf("%s · %d lines", base, lines)
}

func formatReadArgs(args map[string]any, cwd string) string {
	path, _ := args["file_path"].(string)
	offsetFloat, hasOffset := args["offset"].(float64)
	limitFloat, hasLimit := args["limit"].(float64)

	base := relativePath(cwd, path)
	if hasOffset || hasLimit {
		start := 1
		if hasOffset {
			start = int(offsetFloat)
		}
		if hasLimit {
			return fmt.Sprintf("%s · lines %d-%d", base, start, start+int(limitFloat)-1)
		}
		return fmt.Sprintf("%s · from line %d", base, start)
	}
	return base
}

func formatGlobArgs(args map[string]any, cwd string) string {
	pattern, _ := args["pattern"].(string)
	path, ok := args["path"].(string)
	if ok && path != "" && path != cwd {
		return fmt.Sprintf("%q in %s", pattern, relativePath(cwd, path))
	}
	return fmt.Sprintf("%q", pattern)
}

func formatGrepArgs(args map[string]any, cwd string) string {
	pattern, _ := args["pattern"].(string)
	path, ok := args["path"].(string)
	glob, _ := args["glob"].(string)
	mode, _ := args["output_mode"].(string)

	res := fmt.Sprintf("%q", pattern)
	if ok && path != "" && path != cwd {
		res += fmt.Sprintf(" in %s", relativePath(cwd, path))
	}
	if glob != "" {
		res += fmt.Sprintf(" [%s]", glob)
	}
	if mode == "content" {
		res += " (content)"
	} else if mode == "count" {
		res += " (count)"
	}
	return res
}

func formatDefaultArgs(args map[string]any) string {
	parts := make([]string, 0, len(args))
	for k, v := range args {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return clampString(strings.Join(parts, " "), 70)
}

func formatWriteResult(raw string, width int) string {
	var writeData struct {
		Type     string `json:"type"`
		FilePath string `json:"filePath"`
		Diff     string `json:"diff"`
	}
	if json.Unmarshal([]byte(raw), &writeData) == nil && writeData.Diff != "" {
		return messages.RenderDiff(writeData.FilePath, writeData.Diff, width, "write")
	}
	if strings.Contains(raw, `"type":"create"`) {
		return "File created"
	}
	return "File updated"
}

func formatEditResult(raw string, width int) string {
	var editData struct {
		Type       string `json:"type"`
		FilePath   string `json:"filePath"`
		ReplaceAll bool   `json:"replaceAll"`
		Diff       string `json:"diff"`
	}
	if json.Unmarshal([]byte(raw), &editData) == nil && editData.Diff != "" {
		op := "edit"
		if editData.ReplaceAll {
			op = "replace"
		}
		return messages.RenderDiff(editData.FilePath, editData.Diff, width, op)
	}
	return "File edited"
}

func formatReadResult(raw string, _ int) string {
	lines := strings.Count(raw, "\n") + 1
	bytes := len(raw)
	return fmt.Sprintf("Read %d lines (%s)", lines, formatBytes(bytes))
}

func formatSearchResult(raw string, _ int) string {
	lines := strings.Count(raw, "\n") + 1
	firstLine := strings.SplitN(raw, "\n", 2)[0]
	if strings.HasPrefix(firstLine, "Found ") || strings.HasPrefix(firstLine, "No ") {
		return firstLine
	}
	return fmt.Sprintf("%d results", lines)
}

func formatDefaultResult(raw string) string {
	lines := strings.Count(raw, "\n") + 1
	bytes := len(raw)
	if lines == 1 && bytes < 80 {
		return raw
	}
	return fmt.Sprintf("%d lines, %s", lines, formatBytes(bytes))
}

func clampString(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}

func relativePath(cwd string, path string) string {
	if path == "" {
		return ""
	}
	if rel, err := filepath.Rel(cwd, path); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return path
}

func formatBytes(b int) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	return fmt.Sprintf("%.1f KB", float64(b)/1024.0)
}
