# Go-Docling-Bridge 🚀

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-1.0.0-green.svg)](https://github.com/Dsouza10082/go-docling-bridge/releases/tag/v1.0.0)

A high-performance Go wrapper for the [Docling](https://github.com/docling-project/docling) document processing library, enabling intelligent document and URL conversion to Markdown with concurrent processing capabilities.

![go-docling-logo](https://raw.githubusercontent.com/Dsouza10082/go-docling-bridge/main/assets/logo.png)

## 🙏 Acknowledgments

This project builds upon the excellent work of:

- **[Docling Project](https://github.com/docling-project/docling)** - The amazing document understanding library that makes intelligent document processing possible
- **[Cole Medin (@coleam00)](https://github.com/coleam00/ottomator-agents/tree/main/docling-rag-agent)** - Special thanks for the Python examples and RAG agent implementation that inspired this bridge

## 📋 Table of Contents

- [What's New in v1.0.0](#-whats-new-in-v100)
- [Features](#-features)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [URL Conversion](#-url-conversion)
- [API Reference](#-api-reference)
- [Advanced Usage](#-advanced-usage)
- [Testing](#-testing)
- [Benchmarks](#-benchmarks)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🆕 What's New in v1.0.0

### URL to Markdown Conversion

Version 1.0.0 introduces powerful URL-to-Markdown conversion capabilities with full HTTP handling, concurrent processing, and comprehensive metadata:

```go
bridge := docling_bridge.NewDoclingBridge()

// Single URL conversion
markdown, err := bridge.ConvertURLToMarkdown("https://example.com/article")

// Multiple URLs with concurrent processing
urls := []string{"https://example.com/page1", "https://example.com/page2"}
results, err := bridge.ConvertMultipleURLsToMarkdown(urls)

// Complete conversion with metadata
result, err := bridge.ConvertURLToMarkdownComplete("https://example.com")
fmt.Printf("Title: %s, Processing Time: %v\n", result.Title, result.ProcessingTime)
```

**Key improvements in v1.0.0:**

- 🌐 **URL to Markdown** - Convert web pages directly to Markdown
- ⚡ **Concurrent URL Processing** - Process multiple URLs simultaneously with configurable thread pools
- 📊 **Rich Metadata** - Get page title, content length, processing time, and final URL (after redirects)
- 🔧 **Configurable HTTP Client** - Customize timeouts, redirects, user agent, and content limits
- 🔒 **Thread-safe** - Safe for concurrent use across all operations
- 📁 **Direct File Conversion** - Simplified API with `ConvertOneFileToMarkdown()`

---

## ✨ Features

- **URL to Markdown Conversion**: Convert web pages to clean Markdown with metadata
- **Direct File Conversion**: Convert documents to Markdown strings with a single function call
- **Concurrent Processing**: Process multiple documents/URLs simultaneously with configurable worker pools
- **Hybrid Chunking**: Intelligent document chunking that respects structure and token limits
- **Smart Caching**: Built-in cache system to avoid reprocessing documents
- **Progress Tracking**: Real-time progress monitoring for batch operations
- **Flexible Processing**: Standard, progressive, and batch processing modes
- **Thread-Safe**: Concurrent-safe operations with mutex protection
- **Configurable HTTP Client**: Full control over timeouts, redirects, and request headers
- **Audio Processing**: Support for audio file transcription
- **Python Integration**: Seamless bridge to Docling's Python library

---

## 🔧 Installation

### Prerequisites

- Go 1.21 or higher
- Python 3.10+ with pip
- Docling library

### Install Go Package

```bash
go get github.com/Dsouza10082/go-docling-bridge@v1.0.0
```

### Python Dependencies

The library will automatically check and install required Python dependencies:

```bash
pip install docling
```

For advanced chunking features:
```bash
pip install docling transformers sentence-transformers
```

---

## 🚀 Quick Start

### Convert a Single File to Markdown

```go
package main

import (
    "fmt"
    "log"

    docling_bridge "github.com/Dsouza10082/go-docling-bridge"
)

func main() {
    bridge := docling_bridge.NewDoclingBridge()

    // Convert document to Markdown string
    markdown, err := bridge.ConvertOneFileToMarkdown("document.pdf")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(markdown)
}
```

### Convert a URL to Markdown

```go
package main

import (
    "fmt"
    "log"

    docling_bridge "github.com/Dsouza10082/go-docling-bridge"
)

func main() {
    bridge := docling_bridge.NewDoclingBridge()

    // Convert URL to Markdown
    markdown, err := bridge.ConvertURLToMarkdown("https://example.com/article")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(markdown)
}
```

### Basic Batch Processing

```go
package main

import (
    "fmt"

    docling_bridge "github.com/Dsouza10082/go-docling-bridge"
)

func main() {
    bridge := docling_bridge.NewDoclingBridge()

    bridge.
        SetDocumentPath("documents").
        SetOutputPath("output").
        SetThreadCount(4).
        ReadDirectory().
        StartWithProgress()

    success, errors := bridge.Instance.GetResults()
    fmt.Printf("Processed: %d, Errors: %d\n", len(success), len(errors))
}
```

### Audio Processing

```go
bridge := docling_bridge.NewDoclingBridge()
bridge.
    SetDocumentPath("./audio_files").
    SetOutputPath("./output/transcript.md").
    SetIsAudio(true).
    SetThreadCount(8).
    ReadDirectory().
    StartWithProgress()
```

---

## 🌐 URL Conversion

### Single URL Conversion

Convert a single URL to Markdown:

```go
bridge := docling_bridge.NewDoclingBridge()

// Simple conversion - returns only Markdown string
markdown, err := bridge.ConvertURLToMarkdown("https://example.com/article")
if err != nil {
    log.Fatal(err)
}
fmt.Println(markdown)
```

### Complete URL Conversion with Metadata

Get full conversion results including metadata:

```go
bridge := docling_bridge.NewDoclingBridge()

result, err := bridge.ConvertURLToMarkdownComplete("https://example.com/article")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Source URL: %s\n", result.SourceURL)
fmt.Printf("Final URL: %s\n", result.FinalURL)  // After redirects
fmt.Printf("Title: %s\n", result.Title)
fmt.Printf("Content Length: %d bytes\n", result.ContentLength)
fmt.Printf("Processing Time: %v\n", result.ProcessingTime)
fmt.Printf("Markdown:\n%s\n", result.Markdown)
```

### Multiple URLs Conversion (Concurrent)

Process multiple URLs simultaneously:

```go
bridge := docling_bridge.NewDoclingBridge().
    SetThreadCount(8)  // Configure concurrent workers

urls := []string{
    "https://example.com/page1",
    "https://example.com/page2",
    "https://example.com/page3",
    "https://blog.example.com/article",
}

// Simple conversion - returns slice of Markdown strings
markdowns, err := bridge.ConvertMultipleURLsToMarkdown(urls)
if err != nil {
    log.Fatal(err)
}

for i, md := range markdowns {
    if md != "" {
        fmt.Printf("URL %d: %d characters\n", i+1, len(md))
    } else {
        fmt.Printf("URL %d: conversion failed\n", i+1)
    }
}
```

### Multiple URLs with Complete Metadata

```go
bridge := docling_bridge.NewDoclingBridge().
    SetThreadCount(8)

urls := []string{
    "https://example.com/page1",
    "https://example.com/page2",
}

results, err := bridge.ConvertMultipleURLsToMarkdownComplete(urls)
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    if result != nil {
        fmt.Printf("✅ %s\n", result.Title)
        fmt.Printf("   URL: %s\n", result.FinalURL)
        fmt.Printf("   Size: %d bytes\n", result.ContentLength)
        fmt.Printf("   Time: %v\n", result.ProcessingTime)
    } else {
        fmt.Println("❌ Conversion failed")
    }
}
```

### ResultURL Struct

The `ResultURL` struct provides comprehensive metadata about URL conversions:

```go
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
```

### URL Converter Configuration

For advanced use cases, you can configure the URL converter directly:

```go
import "github.com/Dsouza10082/go-docling-bridge/model"

// Create custom configuration
config := &model.Config{
    Timeout:         60 * time.Second,      // HTTP request timeout
    MaxContentSize:  20 * 1024 * 1024,      // 20MB max content
    UserAgent:       "MyBot/1.0",           // Custom user agent
    FollowRedirects: true,                  // Follow HTTP redirects
    MaxRedirects:    15,                    // Max redirect hops
    TempDir:         "/tmp/docling",        // Temp directory for processing
    CleanupTemp:     true,                  // Auto-cleanup temp files
    PythonPath:      "/usr/bin/python3.11", // Custom Python path
}

// Create converter with custom config
converter := model.NewConverterURL(config)

// Convert with context for cancellation support
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := converter.ConvertWithContext(ctx, "https://example.com")
```

### Default Configuration

The default configuration provides sensible defaults:

```go
config := model.DefaultConfigForURLMarkdown()
// Returns:
// - Timeout: 30 seconds
// - MaxContentSize: 10MB
// - UserAgent: Mozilla/5.0 compatible
// - FollowRedirects: true
// - MaxRedirects: 10
// - CleanupTemp: true
// - PythonPath: "python3"
```

---

## 📚 API Reference

### DoclingBridge Methods

| Method | Description | Returns |
|--------|-------------|---------|
| `NewDoclingBridge()` | Creates a new DoclingBridge instance | `*DoclingBridge` |
| `SetDocumentPath(path)` | Sets the document directory | `*DoclingBridge` |
| `SetOutputPath(path)` | Sets the output directory | `*DoclingBridge` |
| `SetThreadCount(count)` | Configures concurrent workers | `*DoclingBridge` |
| `SetMaxTokens(tokens)` | Sets maximum tokens per chunk | `*DoclingBridge` |
| `SetIsAdvanced(bool)` | Enables advanced hybrid chunking | `*DoclingBridge` |
| `SetIsAudio(bool)` | Enables audio file processing | `*DoclingBridge` |
| `ReadDirectory()` | Scans and populates file list | `*DoclingBridge` |
| `Start()` | Begins processing | `*DoclingBridge` |
| `StartWithProgress()` | Processes with progress tracking | `*DoclingBridge` |
| `ProcessBatch(start, end)` | Processes a range of files | `*DoclingBridge` |
| `GetResults()` | Returns success and error lists | `*DoclingBridge` |

### File Conversion Methods

| Method | Description | Returns |
|--------|-------------|---------|
| `ConvertOneFileToMarkdown(path)` | Converts a single file to Markdown | `(string, error)` |

### URL Conversion Methods

| Method | Description | Returns |
|--------|-------------|---------|
| `ConvertURLToMarkdown(url)` | Converts URL to Markdown string | `(string, error)` |
| `ConvertURLToMarkdownComplete(url)` | Converts URL with full metadata | `(*ResultURL, error)` |
| `ConvertMultipleURLsToMarkdown(urls)` | Converts multiple URLs concurrently | `([]string, error)` |
| `ConvertMultipleURLsToMarkdownComplete(urls)` | Converts multiple URLs with metadata | `([]*ResultURL, error)` |

### Converter (Direct File Conversion)

| Method | Description | Returns |
|--------|-------------|---------|
| `NewConverter()` | Creates a new Converter instance | `*Converter` |
| `WithPythonPath(path)` | Sets custom Python executable path | `*Converter` |
| `ConvertFile(path)` | Converts document to Markdown string | `(string, error)` |
| `ToMarkdown(path)` | Alias for ConvertFile with caching | `(string, error)` |
| `ToMarkdownResult(path)` | Returns structured Result with metadata | `(*Result, error)` |
| `GetCached(path)` | Retrieves cached conversion result | `(string, bool)` |
| `ClearCache()` | Clears all cached conversions | `void` |

### ConverterURL (URL Conversion)

| Method | Description | Returns |
|--------|-------------|---------|
| `NewConverterURL(config)` | Creates URL converter with config | `*ConverterURL` |
| `Convert(url)` | Converts URL to Markdown | `(*ResultURL, error)` |
| `ConvertWithContext(ctx, url)` | Converts with context support | `(*ResultURL, error)` |
| `ConvertToString(url)` | Returns only Markdown string | `(string, error)` |

### DoclingModel Methods

| Method | Description | Returns |
|--------|-------------|---------|
| `NewDoclingModel()` | Creates a new DoclingModel instance | `*DoclingModel` |
| `SetDocumentPath(path)` | Sets the directory containing documents | `*DoclingModel` |
| `SetOutputPath(path)` | Sets the output directory | `*DoclingModel` |
| `SetFilePath(path)` | Sets a single file path | `*DoclingModel` |
| `SetThreadCount(count)` | Sets number of concurrent workers | `*DoclingModel` |
| `SetMaxTokens(tokens)` | Sets maximum tokens per chunk | `*DoclingModel` |
| `SetIsAdvanced(bool)` | Enables advanced hybrid chunking | `*DoclingModel` |
| `SetIsAudio(bool)` | Enables audio file processing | `*DoclingModel` |
| `StartConvertToMarkdownProcess(path)` | Converts single file to Markdown string | `string` |
| `ReadDirectory()` | Scans document path and populates file list | `*DoclingModel` |
| `Start()` | Begins processing all files | `*DoclingModel` |
| `StartWithProgress()` | Processes files with progress tracking | `*DoclingModel` |
| `ProcessBatch(start, end)` | Processes a specific range of files | `*DoclingModel` |
| `GetResults()` | Returns success and error lists | `([]string, []string)` |
| `GetCachedContent(path)` | Retrieves cached content for a file | `(string, bool)` |
| `ClearCache()` | Clears the internal document cache | `void` |
| `ConvertOneFileToMarkdown(path)` | Converts file to Markdown | `(string, error)` |
| `ConvertURLToMarkdown(url)` | Converts URL to Markdown | `(string, error)` |
| `ConvertURLToMarkdownComplete(url)` | Converts URL with metadata | `(*ResultURL, error)` |
| `ConvertMultipleURLsToMarkdown(urls)` | Converts multiple URLs | `([]string, error)` |
| `ConvertMultipleURLsToMarkdownComplete(urls)` | Converts multiple URLs with metadata | `([]*ResultURL, error)` |

---

## 🔬 Advanced Usage

### Direct Markdown Conversion with Caching

```go
package main

import (
    "fmt"
    "log"

    "github.com/Dsouza10082/go-docling-bridge/model"
)

func main() {
    converter := model.NewConverter()

    files := []string{"doc1.pdf", "doc2.docx", "doc3.html"}

    for _, file := range files {
        // First call processes the document
        md, err := converter.ToMarkdown(file)
        if err != nil {
            log.Printf("Error converting %s: %v", file, err)
            continue
        }
        fmt.Printf("Converted %s: %d characters\n", file, len(md))

        // Second call uses cache (instant)
        if cached, exists := converter.GetCached(file); exists {
            fmt.Printf("Cache hit for %s: %d characters\n", file, len(cached))
        }
    }
}
```

### Concurrent URL Processing with Error Handling

```go
package main

import (
    "fmt"
    "log"

    docling_bridge "github.com/Dsouza10082/go-docling-bridge"
)

func main() {
    bridge := docling_bridge.NewDoclingBridge().
        SetThreadCount(10)

    urls := []string{
        "https://example.com/page1",
        "https://invalid-url.notadomain",
        "https://example.com/page2",
    }

    results, err := bridge.ConvertMultipleURLsToMarkdownComplete(urls)
    if err != nil {
        log.Fatal(err)
    }

    successful := 0
    failed := 0
    
    for i, result := range results {
        if result != nil {
            successful++
            fmt.Printf("✅ [%d] %s - %d chars in %v\n", 
                i+1, result.Title, len(result.Markdown), result.ProcessingTime)
        } else {
            failed++
            fmt.Printf("❌ [%d] %s - conversion failed\n", i+1, urls[i])
        }
    }
    
    fmt.Printf("\nSummary: %d successful, %d failed\n", successful, failed)
}
```

### Advanced Chunking for Large Documents

```go
bridge := docling_bridge.NewDoclingBridge()
bridge.Instance.
    SetIsAdvanced(true).
    SetMaxTokens(512).
    SetDocumentPath("documents").
    SetOutputPath("output").
    SetThreadCount(4).
    ReadDirectory().
    Start()
```

### Custom Python Environment

```go
// Using environment variable
os.Setenv("PYTHON_PATH", "/path/to/venv/bin/python")
converter := model.NewConverter()

// Or using fluent API
converter := model.NewConverter().
    WithPythonPath("/usr/local/bin/python3.11")
```

### Context-Aware URL Conversion

```go
import (
    "context"
    "time"
    
    "github.com/Dsouza10082/go-docling-bridge/model"
)

func fetchWithTimeout(url string) (*model.ResultURL, error) {
    config := model.DefaultConfigForURLMarkdown()
    converter := model.NewConverterURL(config)

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    return converter.ConvertWithContext(ctx, url)
}
```

---

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with race detector
go test -race ./...

# Run short tests only
go test -short ./...

# Run specific test
go test -v -run TestConverter_ToMarkdown

# Run URL conversion tests
go test -v -run TestConvertURL

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Test Categories

```bash
# Unit tests (fast)
go test -short -v ./...

# Integration tests (requires docling)
go test -v -run Integration ./...

# URL conversion tests
go test -v -run URL ./...

# Performance tests
go test -v -run Performance ./...
```

---

## ⚡ Benchmarks

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run with memory stats
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkConverter ./...
go test -bench=BenchmarkURLConversion ./...

# CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -bench=. -memprofile=mem.prof ./...
go tool pprof mem.prof
```

### Expected Performance

| Operation | Time | Memory |
|-----------|------|--------|
| Single file (small PDF) | ~200-500ms | ~50MB |
| Single URL conversion | ~500ms-2s | ~30MB |
| Batch files (10 files, 4 workers) | ~1-2s | ~100MB |
| Batch URLs (10 URLs, 8 workers) | ~2-4s | ~80MB |
| Batch (100 files, 8 workers) | ~10-20s | ~200MB |
| Cache hit | <1ms | ~0 |

---

## 📁 Project Structure

```
go-docling-bridge/
├── docling_bridge.go                    # Main bridge interface
├── model/
│   ├── docling_bridge.model.go          # Core DoclingModel
│   ├── converter.go                     # File Converter type
│   ├── converter_url_model.go           # URL Converter type (v1.0.0)
│   ├── docling_wrapper.py               # Basic Python wrapper
│   ├── docling_wrapper_advanced.py      # Advanced chunking wrapper
│   ├── docling_wrapper_audio.py         # Audio processing wrapper
│   ├── docling_convert_one_file_to_markdown.py  # Direct conversion script
│   └── docling_convert_url_to_md.py     # URL conversion script (v1.0.0)
├── tests/
│   ├── docling_bridge_test.go
│   ├── converter_test.go
│   ├── converter_url_test.go
│   └── benchmark_test.go
├── docs/
│   ├── API.md
│   └── TESTING.md
└── examples/
    └── main.go
```

---

## 📄 Supported Formats

### Document Formats

Docling supports the following document formats:

| Format | Extensions | Notes |
|--------|------------|-------|
| PDF | `.pdf` | Full layout understanding |
| Word | `.docx` | Tables, formatting preserved |
| PowerPoint | `.pptx` | Slides to sections |
| Excel | `.xlsx` | Tables extraction |
| HTML | `.html` | Structure preserved |
| Markdown | `.md` | Pass-through with normalization |
| CSV | `.csv` | Table format |
| Images | `.png`, `.jpg`, `.jpeg`, `.tiff` | OCR supported |
| Audio | `.wav`, `.mp3` | ASR transcription |

### URL Support

| Protocol | Support |
|----------|---------|
| HTTP | ✅ Full support |
| HTTPS | ✅ Full support |
| Redirects | ✅ Configurable (up to 10 by default) |
| Gzip | ✅ Automatic decompression |

---

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Add tests for new functionality
4. Ensure all tests pass (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🐛 Issues and Support

Found a bug or need help? Please [open an issue](https://github.com/Dsouza10082/go-docling-bridge/issues) on GitHub.

---

## 🔗 Related Projects

- [Docling](https://github.com/docling-project/docling) - Core document processing library
- [Ottomator Agents](https://github.com/coleam00/ottomator-agents) - AI agent implementations using Docling

---

## 📈 Changelog

### v1.0.0 (Latest)

- 🌐 **New**: URL to Markdown conversion with `ConvertURLToMarkdown()`
- 🌐 **New**: Complete URL conversion with metadata via `ConvertURLToMarkdownComplete()`
- ⚡ **New**: Concurrent URL processing with `ConvertMultipleURLsToMarkdown()`
- ⚡ **New**: Concurrent URL processing with metadata via `ConvertMultipleURLsToMarkdownComplete()`
- 📊 **New**: `ResultURL` struct with comprehensive metadata (title, content length, processing time, final URL)
- 🔧 **New**: `ConverterURL` type with configurable HTTP client
- 🔧 **New**: `Config` struct for URL conversion customization
- 📁 **New**: Direct file conversion via `ConvertOneFileToMarkdown()` in DoclingBridge
- 🎧 **New**: Audio processing support with `SetIsAudio()`
- 🔒 **Improved**: Thread-safe concurrent processing for URLs
- 📚 **Improved**: Comprehensive documentation

### v0.8.0

- ✨ `Converter` type for direct Markdown conversion
- ✨ `ConvertFile()` and `ToMarkdown()` methods
- ✨ `ToMarkdownResult()` for structured results
- ✨ `StartConvertToMarkdownProcess()` in DoclingModel
- ✨ Embedded Python script via `//go:embed`
- 🔧 Thread-safe caching mechanism
- 📚 Comprehensive documentation

### v0.1.0

- Initial release with batch processing
- Worker pool management
- Advanced hybrid chunking support

---

**Made with ❤️ by the Go-Docling-Bridge team**
