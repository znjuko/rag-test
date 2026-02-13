package model

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed docling_convert_url_to_md.py
var doclingConvertUrlToMd string

type Config struct {
	// Timeout for HTTP requests (default: 30 seconds)
	Timeout time.Duration
	// MaxContentSize maximum HTML content size in bytes (default: 10MB)
	MaxContentSize int64
	// UserAgent for HTTP requests
	UserAgent string
	// FollowRedirects whether to follow HTTP redirects (default: true)
	FollowRedirects bool
	// MaxRedirects maximum number of redirects to follow (default: 10)
	MaxRedirects int
	// TempDir directory for temporary files used by docling
	TempDir string
	// CleanupTemp remove temporary files after conversion (default: true)
	CleanupTemp bool
	// PythonPath path to Python executable (default: "python3")
	PythonPath string
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfigForURLMarkdown() *Config {
	return &Config{
		Timeout:         30 * time.Second,
		MaxContentSize:  10 * 1024 * 1024, // 10MB
		UserAgent:       "Mozilla/5.0 (compatible; HTMLToMarkdown/1.0; +https://github.com/example/htmltomd)",
		FollowRedirects: true,
		MaxRedirects:    10,
		TempDir:         os.TempDir(),
		CleanupTemp:     true,
		PythonPath:      "python3",
	}
}

// Result holds the conversion result
type ResultURL struct {
	// Markdown is the converted content
	Markdown string
	// SourceURL is the original URL
	SourceURL string
	// Title is the page title if detected
	Title string
	// ContentLength is the size of the original HTML
	ContentLength int64
	// ProcessingTime is how long the conversion took
	ProcessingTime time.Duration
	// FinalURL is the URL after redirects (may differ from SourceURL)
	FinalURL string
}

// Converter handles URL to Markdown conversions
type ConverterURL struct {
	config *Config
	client *http.Client
}

// NewConverter creates a new Converter with the given configuration
// If config is nil, default configuration will be used
func NewConverterURL(config *Config) *ConverterURL {
	if config == nil {
		config = DefaultConfigForURLMarkdown()
	}

	// Create HTTP client with redirect policy
	client := &http.Client{
		Timeout: config.Timeout,
	}

	if config.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", config.MaxRedirects)
			}
			return nil
		}
	} else {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return &ConverterURL{
		config: config,
		client: client,
	}
}

// Convert fetches the URL and converts its HTML content to Markdown
// The Markdown is returned in memory without writing to disk (except for temp processing)
func (c *ConverterURL) Convert(urlStr string) (*ResultURL, error) {
	return c.ConvertWithContext(context.Background(), urlStr)
}

// ConvertWithContext fetches the URL and converts its HTML content to Markdown
// with context support for cancellation
func (c *ConverterURL) ConvertWithContext(ctx context.Context, urlStr string) (*ResultURL, error) {
	startTime := time.Now()

	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s (only http and https are supported)", parsedURL.Scheme)
	}

	// Fetch HTML content
	htmlContent, finalURL, err := c.fetchHTML(ctx, urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	// Convert HTML to Markdown using docling
	markdown, title, err := c.htmlToMarkdown(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to convert HTML to Markdown: %w", err)
	}

	return &ResultURL{
		Markdown:       markdown,
		SourceURL:      urlStr,
		Title:          title,
		ContentLength:  int64(len(htmlContent)),
		ProcessingTime: time.Since(startTime),
		FinalURL:       finalURL,
	}, nil
}

// ConvertToString is a convenience method that returns only the Markdown string
func (c *ConverterURL) ConvertToString(urlStr string) (string, error) {
	result, err := c.Convert(urlStr)
	if err != nil {
		return "", err
	}
	return result.Markdown, nil
}

// fetchHTML fetches the HTML content from the URL
func (c *ConverterURL) fetchHTML(ctx context.Context, urlStr string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to mimic a browser
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "text/html") &&
		!strings.Contains(strings.ToLower(contentType), "application/xhtml") {
		return "", "", fmt.Errorf("unexpected content type: %s (expected HTML)", contentType)
	}

	// Handle gzip encoding
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", "", fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Read with size limit
	limitedReader := io.LimitReader(reader, c.config.MaxContentSize+1)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response body: %w", err)
	}

	if int64(len(body)) > c.config.MaxContentSize {
		return "", "", fmt.Errorf("content too large: exceeds %d bytes", c.config.MaxContentSize)
	}

	finalURL := resp.Request.URL.String()

	return string(body), finalURL, nil
}

// htmlToMarkdown converts HTML content to Markdown using docling via Python
func (c *ConverterURL) htmlToMarkdown(htmlContent string) (string, string, error) {
	hash := sha256.Sum256([]byte(htmlContent + time.Now().String()))
	tempDir := filepath.Join(c.config.TempDir, fmt.Sprintf("htmltomd_%x", hash[:8]))

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	if c.config.CleanupTemp {
		defer os.RemoveAll(tempDir)
	}

	htmlPath := filepath.Join(tempDir, "input.html")
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write temp HTML file: %w", err)
	}

	pythonScript := doclingConvertUrlToMd

	scriptPath := filepath.Join(tempDir, "convertToUrl.py")
	if err := os.WriteFile(scriptPath, []byte(pythonScript), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write Python script: %w", err)
	}

	cmd := exec.Command(c.config.PythonPath, scriptPath, htmlPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("docling conversion failed: %v\nstderr: %s", err, stderr.String())
	}

	var result struct {
		Markdown string `json:"markdown"`
		Title    string `json:"title"`
		Status   string `json:"status"`
		Error    string `json:"error"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return "", "", fmt.Errorf("failed to parse conversion output: %w", err)
	}

	if result.Status == "error" {
		return "", "", fmt.Errorf("conversion error: %s", result.Error)
	}

	return result.Markdown, result.Title, nil
}