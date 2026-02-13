package model

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed docling_convert_one_file_to_markdown.py
var doclingConvertOneFileToMarkdown string

// Result represents the output from document conversion
type Result struct {
	Status   string `json:"status"`
	Markdown string `json:"markdown"`
	File     string `json:"file"`
	Error    string `json:"error,omitempty"`
}

// Converter handles document to markdown conversion using Docling
type Converter struct {
	PythonPath string
	Cache      map[string]string
	mutex      sync.RWMutex
}

// NewConverter creates a new Docling converter instance
func NewConverter() *Converter {
	pythonPath := "python3"
	if p := os.Getenv("PYTHON_PATH"); p != "" {
		pythonPath = p
	}

	return &Converter{
		PythonPath: pythonPath,
		Cache:      make(map[string]string),
	}
}

// WithPythonPath sets a custom Python executable path
func (c *Converter) WithPythonPath(path string) *Converter {
	c.PythonPath = path
	return c
}

// ToMarkdown converts a document file to markdown string
// Returns the markdown content or an error
func (c *Converter) ToMarkdown(filePath string) (string, error) {
	// Check cache first
	c.mutex.RLock()
	if cached, exists := c.Cache[filePath]; exists {
		c.mutex.RUnlock()
		return cached, nil
	}
	c.mutex.RUnlock()

	// Resolve absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path %s: %w", filePath, err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", absPath)
	}

	// Create temporary Python script
	scriptPath, err := c.createPythonScript()
	if err != nil {
		return "", fmt.Errorf("failed to create python script: %w", err)
	}
	defer os.Remove(scriptPath)

	// Execute Python script
	cmd := exec.Command(c.PythonPath, scriptPath, absPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docling execution failed: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON result from output
	result, err := c.parseOutput(string(output))
	if err != nil {
		return "", err
	}

	if result.Status != "success" {
		return "", fmt.Errorf("conversion failed: %s", result.Error)
	}

	// Cache the result
	c.mutex.Lock()
	c.Cache[filePath] = result.Markdown
	c.mutex.Unlock()

	return result.Markdown, nil
}

// ToMarkdownResult converts a document and returns the full Result struct
func (c *Converter) ToMarkdownResult(filePath string) (*Result, error) {
	md, err := c.ToMarkdown(filePath)
	if err != nil {
		return &Result{
			Status: "failed",
			File:   filepath.Base(filePath),
			Error:  err.Error(),
		}, err
	}

	return &Result{
		Status:   "success",
		Markdown: md,
		File:     filepath.Base(filePath),
	}, nil
}

// ClearCache removes all cached conversions
func (c *Converter) ClearCache() {
	c.mutex.Lock()
	c.Cache = make(map[string]string)
	c.mutex.Unlock()
}

// GetCached retrieves a cached markdown conversion if available
func (c *Converter) GetCached(filePath string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	md, exists := c.Cache[filePath]
	return md, exists
}

// createPythonScript creates a temporary Python script for conversion
func (c *Converter) createPythonScript() (string, error) {
	script := doclingConvertOneFileToMarkdown

	tmpFile, err := os.CreateTemp("", "docling_convert_*.py")
	if err != nil {
		return "", err
	}

	if _, err := tmpFile.WriteString(script); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", err
	}

	tmpFile.Close()
	return tmpFile.Name(), nil
}

// parseOutput extracts the JSON result from Python script output
func (c *Converter) parseOutput(output string) (*Result, error) {
	startMarker := "JSON_RESULT_START"
	endMarker := "JSON_RESULT_END"

	startIdx := strings.Index(output, startMarker)
	endIdx := strings.Index(output, endMarker)

	if startIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("invalid output format from Python script: %s", output)
	}

	jsonStr := output[startIdx+len(startMarker) : endIdx]
	jsonStr = strings.TrimSpace(jsonStr)

	var result Result
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON result: %w", err)
	}

	return &result, nil
}

// ConvertFile is a convenience method that wraps ToMarkdown
// This provides a simpler API for one-off conversions
func (c *Converter) ConvertFile(filePath string) (string, error) {
	return c.ToMarkdown(filePath)
}