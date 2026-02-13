# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.8.0] - 2025-01-23

### Added

- **New `Converter` type** for streamlined document-to-Markdown conversion
  - `NewConverter()` - Creates a new Converter instance
  - `WithPythonPath(path)` - Fluent API for custom Python path
  - `ConvertFile(path)` - One-liner conversion returning Markdown string
  - `ToMarkdown(path)` - Alias with built-in caching
  - `ToMarkdownResult(path)` - Returns structured `Result` with metadata
  - `GetCached(path)` - Retrieve cached conversion results
  - `ClearCache()` - Clear all cached conversions

- **New `Result` struct** for structured conversion results
  - `Status` - "success" or "failed"
  - `Markdown` - Converted content
  - `File` - Original filename
  - `Error` - Error message (on failure)

- **`StartConvertToMarkdownProcess(path)`** method in DoclingModel
  - Direct integration of Converter into existing workflow
  - Returns Markdown string directly
  - Automatically updates SuccessList/ErrorList

- **Embedded Python script** via `//go:embed`
  - `docling_convert_one_file_to_markdown.py` embedded at compile time
  - No external script files required at runtime
  - Cleaner deployment and distribution

- **Comprehensive test suite** for Converter
  - Unit tests for all methods
  - Benchmark tests for performance
  - Concurrent access tests
  - Integration tests (requires docling)

### Changed

- DoclingModel now includes `ConvertToMarkdown *Converter` field
- Improved thread-safety for cache operations
- Updated documentation with new API examples

### Technical Details

- Converter uses `sync.RWMutex` for thread-safe cache access
- Python script output uses JSON markers for reliable parsing
- Cache key is the file path (supports both relative and absolute)

## [0.1.0] - 2024-10-09

### Added

- Initial release
- `DoclingBridge` high-level interface
- `DoclingModel` core processing engine
- Worker pool for concurrent document processing
- Advanced hybrid chunking support via `docling_wrapper_advanced.py`
- Audio file support via `docling_wrapper_audio.py`
- Batch processing with `ProcessBatch()`
- Progress tracking with `StartWithProgress()`
- Smart caching to avoid reprocessing
- Python dependency auto-installation

### Features

- Process multiple documents concurrently
- Configurable thread count
- Configurable token limits for chunking
- Multiple output formats support
- Cache management
- Error tracking and reporting

---

## Migration Guide

### From v0.1.0 to v0.8.0

**New simpler API for single file conversion:**

```go
// Old way (still works)
bridge := docling_bridge.NewDoclingBridge()
bridge.SetDocumentPath("./").SetOutputPath("output").ReadDirectory().Start()
content, _ := bridge.Instance.GetCachedContent("file.pdf")

// New way (v0.8.0)
converter := model.NewConverter()
markdown, err := converter.ConvertFile("file.pdf")
```

**Using Converter with DoclingModel:**

```go
// New integrated method
model := model.NewDoclingModel()
markdown := model.StartConvertToMarkdownProcess("document.pdf")
```

**Structured results:**

```go
converter := model.NewConverter()
result, err := converter.ToMarkdownResult("document.pdf")

fmt.Printf("Status: %s\n", result.Status)
fmt.Printf("File: %s\n", result.File)
fmt.Printf("Content: %s\n", result.Markdown)
```

---

[0.8.0]: https://github.com/Dsouza10082/go-docling-bridge/compare/v0.1.0...v0.8.0
[0.1.0]: https://github.com/Dsouza10082/go-docling-bridge/releases/tag/v0.1.0