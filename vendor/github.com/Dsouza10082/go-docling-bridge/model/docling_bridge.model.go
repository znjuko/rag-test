package model

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"iter"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

//go:embed docling_wrapper.py
var doclingWrapper string

//go:embed docling_wrapper_advanced.py
var doclingWrapperAdvanced string

//go:embed docling_wrapper_audio.py
var doclingWrapperAudio string

type DoclingModel struct {
	wrapper              string
	mutex                sync.RWMutex
	textCache            map[string]string
	pythonPath           string
	DocumentPath         string
	filePath             string
	OutputPath           string
	FileList             []string
	ErrorList            []string
	SuccessList          []string
	threadCount          int
	errorMutex           sync.Mutex
	successMutex         sync.Mutex
	maxTokens            int
	isAdvanced           bool
	isAudio              bool
	ConvertToMarkdown    *Converter
	ConvertToURLMarkdown *ConverterURL
}

type OutputData struct {
	File   string `json:"file"`
	Format string `json:"format"`
	Status string `json:"status"`
	Error  string `json:"error"`
}

type ProcessResult struct {
	File   string
	Output *OutputData
	Error  error
}

func NewDoclingModel() *DoclingModel {

	config := DefaultConfigForURLMarkdown()

	doc := &DoclingModel{
		wrapper:           doclingWrapper,
		mutex:             sync.RWMutex{},
		textCache:         make(map[string]string),
		pythonPath:        "",
		DocumentPath:      "",
		OutputPath:        "",
		FileList:          []string{},
		ErrorList:         []string{},
		SuccessList:       []string{},
		threadCount:       4,
		maxTokens:         512,
		isAdvanced:        false,
		isAudio:           false,
		ConvertToMarkdown: NewConverter(),
		ConvertToURLMarkdown: NewConverterURL(config),
	}
	doc.findPythonPath()
	doc.checkDependencies()
	return doc
}

func (pe *DoclingModel) StartConvertToMarkdownProcess(filePath string) string {
	md, err := pe.ConvertToMarkdown.ConvertFile(filePath)
	if err != nil {
		pe.ErrorList = append(pe.ErrorList, fmt.Sprintf("%s: %v", filePath, err))
		return ""
	}
	pe.SuccessList = append(pe.SuccessList, filePath)
	return md
}

func (pe *DoclingModel) SetFilePath(path string) *DoclingModel {
	pe.mutex.Lock()
	pe.filePath = path
	pe.mutex.Unlock()
	return pe
}

func (pe *DoclingModel) SetIsAudio(isAudio bool) *DoclingModel {
	pe.mutex.Lock()
	pe.isAudio = isAudio
	pe.mutex.Unlock()
	return pe
}

func (pe *DoclingModel) SetDocumentPath(path string) *DoclingModel {
	pe.mutex.Lock()
	pe.DocumentPath = path
	pe.mutex.Unlock()
	return pe
}

func (pe *DoclingModel) SetOutputPath(path string) *DoclingModel {
	pe.mutex.Lock()
	pe.OutputPath = path
	pe.mutex.Unlock()
	return pe
}

func (pe *DoclingModel) SetThreadCount(count int) *DoclingModel {
	pe.mutex.Lock()
	if count < 1 {
		count = 1
	}
	pe.threadCount = count
	pe.mutex.Unlock()
	return pe
}

func (pe *DoclingModel) SetMaxTokens(tokens int) *DoclingModel {
	pe.mutex.Lock()
	pe.maxTokens = tokens
	pe.mutex.Unlock()
	return pe
}

func (pe *DoclingModel) SetIsAdvanced(isAdvanced bool) *DoclingModel {
	pe.mutex.Lock()
	pe.isAdvanced = isAdvanced
	pe.mutex.Unlock()
	return pe
}

func (pe *DoclingModel) CreateDoclingWrapper() (string, error) {
	wrapper := doclingWrapper
	if pe.isAdvanced {
		wrapper = doclingWrapperAdvanced
	}
	if pe.isAudio {
		wrapper = doclingWrapperAudio
	}
	tmpFile, err := os.CreateTemp("", "docling_wrapper_*.py")
	if err != nil {
		return "", fmt.Errorf("error creating temporary file: %v", err)
	}
	if _, err := tmpFile.WriteString(wrapper); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("error writing script: %v", err)
	}
	tmpFile.Close()
	return tmpFile.Name(), nil
}

func (pe *DoclingModel) extractDataToMD(path string) (*OutputData, error) {
	pe.mutex.RLock()
	if cachedText, exists := pe.textCache[path]; exists {
		pe.mutex.RUnlock()
		return &OutputData{
			Status: "success",
			Format: cachedText,
			File:   filepath.Base(path),
		}, nil
	}
	pe.mutex.RUnlock()

	scriptPath, err := pe.CreateDoclingWrapper()
	if err != nil {
		return nil, err
	}
	defer os.Remove(scriptPath)
	var cmd *exec.Cmd
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("⚠️ Warning: Could not resolve absolute path for %s: %v", path, err)
		return nil, err
	}
	absOutputPath, err := filepath.Abs(pe.OutputPath)
	if err != nil {
		log.Printf("⚠️ Warning: Could not resolve absolute path for %s: %v", path, err)
		return nil, err
	}
	if pe.isAdvanced {
		cmd = exec.Command(pe.pythonPath, scriptPath, absPath, absOutputPath, strconv.Itoa(pe.maxTokens))
	} else if pe.isAudio {
		cmd = exec.Command(pe.pythonPath, scriptPath, absPath, absOutputPath)
	} else {
		cmd = exec.Command(pe.pythonPath, scriptPath, absPath, absOutputPath)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &OutputData{
			Status: "failed",
			Error:  fmt.Sprintf("Error executing script: %v", err),
		}, nil
	}

	outputStr := string(output)
	jsonStart := strings.Index(outputStr, "JSON_RESULT_START")
	jsonEnd := strings.Index(outputStr, "JSON_RESULT_END")
	if jsonStart == -1 || jsonEnd == -1 {
		return &OutputData{
			Status: "failed",
			Error:  "Invalid output format from Python script",
		}, nil
	}

	jsonStr := outputStr[jsonStart+len("JSON_RESULT_START") : jsonEnd]
	jsonStr = strings.TrimSpace(jsonStr)

	var result OutputData
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return &OutputData{
			Status: "failed",
			Error:  fmt.Sprintf("Error decoding JSON: %v", err),
		}, nil
	}

	if result.Status == "success" {
		pe.successMutex.Lock()
		pe.SuccessList = append(pe.SuccessList, path)
		pe.successMutex.Unlock()

		pe.mutex.Lock()
		pe.textCache[path] = result.Format
		pe.mutex.Unlock()
	}

	return &result, nil
}

func (pe *DoclingModel) checkDependencies() error {
	dependencies := []string{"docling"}
	fmt.Printf("🔍 Checking Docling dependencies...\n")
	for _, dep := range dependencies {
		cmd := exec.Command(pe.pythonPath, "-c", fmt.Sprintf("import %s", strings.ToLower(dep)))
		if err := cmd.Run(); err != nil {
			fmt.Printf("Dependency %s not found. Trying to install...\n", dep)
			installCmd := exec.Command(pe.pythonPath, "-m", "pip", "install", dep)
			if err := installCmd.Run(); err != nil {
				return fmt.Errorf("error installing %s: %v\nTry manually: pip install %s", dep, err, dep)
			}
			fmt.Printf("✅ %s installed successfully\n", dep)
		} else {
			fmt.Printf("✅ %s already installed\n", dep)
		}
	}
	cmd := exec.Command("python3", "-c", "import docling")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docling library not found")
	}
	return nil
}

func (pe *DoclingModel) findPythonPath() {
	pythonCommands := []string{"python3", "python"}
	for _, cmd := range pythonCommands {
		if path, err := exec.LookPath(cmd); err == nil {
			pe.pythonPath = path
			fmt.Printf("Python found in: %s\n", path)
			return
		}
	}
	log.Fatal("❌ Python not found in system. Please ensure Python is installed and in the PATH.")
}

func (pe *DoclingModel) ReadDirectory() *DoclingModel {
	pe.mutex.Lock()
	defer pe.mutex.Unlock()
	err := filepath.WalkDir(pe.DocumentPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			pe.ErrorList = append(pe.ErrorList, err.Error())
			return nil
		}
		if !d.IsDir() {
			pe.FileList = append(pe.FileList, path)
		}
		return nil
	})
	if err != nil {
		pe.ErrorList = append(pe.ErrorList, err.Error())
	}
	return pe
}

func (pe *DoclingModel) filesIterator() iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		for i, file := range pe.FileList {
			if !yield(i, file) {
				return
			}
		}
	}
}

func (pe *DoclingModel) processFileWorker(jobs <-chan string, results chan<- ProcessResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for file := range jobs {
		output, err := pe.extractDataToMD(file)
		results <- ProcessResult{
			File:   file,
			Output: output,
			Error:  err,
		}
	}
}

func (pe *DoclingModel) Start() *DoclingModel {
	totalFiles := len(pe.FileList)
	if totalFiles == 0 {
		fmt.Println("⚠️ No files to process")
		return pe
	}

	fmt.Printf("🚀 Starting processing of %d files with %d workers...\n", totalFiles, pe.threadCount)

	jobs := make(chan string, totalFiles)
	results := make(chan ProcessResult, totalFiles)

	var wg sync.WaitGroup

	for w := 0; w < pe.threadCount; w++ {
		wg.Add(1)
		go pe.processFileWorker(jobs, results, &wg)
	}

	go func() {
		for i, file := range pe.filesIterator() {
			fmt.Printf("📄 Queuing [%d/%d]: %s\n", i+1, totalFiles, filepath.Base(file))
			jobs <- file
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	processed := 0
	for result := range results {
		processed++
		if result.Error != nil {
			pe.errorMutex.Lock()
			pe.ErrorList = append(pe.ErrorList, fmt.Sprintf("%s: %v", result.File, result.Error))
			pe.errorMutex.Unlock()
		}
		fmt.Printf("✓ Processed [%d/%d]: %s\n", processed, totalFiles, filepath.Base(result.File))
	}
	return pe
}

func (pe *DoclingModel) StartWithProgress() *DoclingModel {
	totalFiles := len(pe.FileList)
	if totalFiles == 0 {
		fmt.Println("⚠️ No files to process")
		return pe
	}

	fmt.Printf("🚀 Starting processing of %d files with %d workers...\n", totalFiles, pe.threadCount)

	jobs := make(chan string, totalFiles)
	results := make(chan ProcessResult, totalFiles)

	var wg sync.WaitGroup

	for w := 0; w < pe.threadCount; w++ {
		wg.Add(1)
		go pe.processFileWorker(jobs, results, &wg)
	}

	go func() {
		for _, file := range pe.filesIterator() {
			jobs <- file
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	processed := 0
	for result := range results {
		processed++
		percentage := (processed * 100) / totalFiles

		if result.Error != nil {
			pe.errorMutex.Lock()
			pe.ErrorList = append(pe.ErrorList, fmt.Sprintf("%s: %v", result.File, result.Error))
			pe.errorMutex.Unlock()
		}

		fmt.Printf("\r⏳ Progress: %d/%d (%d%%) | Latest: %s",
			processed, totalFiles, percentage, filepath.Base(result.File))
	}

	return pe
}

func (pe *DoclingModel) ProcessBatch(startIdx, endIdx int) *DoclingModel {
	if startIdx < 0 || endIdx > len(pe.FileList) || startIdx >= endIdx {
		fmt.Println("⚠️ Invalid batch range")
		return pe
	}

	batchFiles := pe.FileList[startIdx:endIdx]
	totalFiles := len(batchFiles)

	fmt.Printf("🚀 Processing batch: files %d-%d (%d files) with %d workers...\n",
		startIdx+1, endIdx, totalFiles, pe.threadCount)

	jobs := make(chan string, totalFiles)
	results := make(chan ProcessResult, totalFiles)

	var wg sync.WaitGroup

	for w := 0; w < pe.threadCount; w++ {
		wg.Add(1)
		go pe.processFileWorker(jobs, results, &wg)
	}

	go func() {
		for _, file := range batchFiles {
			jobs <- file
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	processed := 0
	for result := range results {
		processed++
		if result.Error != nil {
			pe.errorMutex.Lock()
			pe.ErrorList = append(pe.ErrorList, fmt.Sprintf("%s: %v", result.File, result.Error))
			pe.errorMutex.Unlock()
		}
		fmt.Printf("✓ Batch processed [%d/%d]: %s\n", processed, totalFiles, filepath.Base(result.File))
	}

	return pe
}

func (pe *DoclingModel) GetResults() (success []string, errors []string) {
	pe.successMutex.Lock()
	pe.errorMutex.Lock()
	defer pe.successMutex.Unlock()
	defer pe.errorMutex.Unlock()

	return pe.SuccessList, pe.ErrorList
}

func (pe *DoclingModel) GetCachedContent(path string) (string, bool) {
	pe.mutex.RLock()
	defer pe.mutex.RUnlock()
	content, exists := pe.textCache[path]
	return content, exists
}

func (pe *DoclingModel) ClearCache() {
	pe.mutex.Lock()
	defer pe.mutex.Unlock()
	pe.textCache = make(map[string]string)
}

// ConvertOneFileToMarkdown converts a file to a markdown string without the Result struct
func (pe *DoclingModel) ConvertOneFileToMarkdown(filePath string) (string, error) {
	result, err := pe.ConvertToMarkdown.ConvertFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to convert file to markdown: %w", err)
	}
	return result, nil
}

// ConvertURLToMarkdown converts a URL to a markdown string without the ResultURL struct
func (pe *DoclingModel) ConvertURLToMarkdown(url string) (string, error) {
	result, err := pe.ConvertToURLMarkdown.Convert(url)
	if err != nil {
		return "", err
	}
	return result.Markdown, nil
}

// ConvertURLToMarkdownComplete converts a URL to a complete ResultURL struct
func (pe *DoclingModel) ConvertURLToMarkdownComplete(url string) (*ResultURL, error) {
	result, err := pe.ConvertToURLMarkdown.Convert(url)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ConvertMultipleURLsToMarkdownComplete converts multiple URLs to a complete ResultURL struct
func (pe *DoclingModel) ConvertMultipleURLsToMarkdownComplete(urls []string) ([]*ResultURL, error) {
	if pe.threadCount <= 0 {
		pe.threadCount = 4
	}

	results := make([]*ResultURL, len(urls))
	var wg sync.WaitGroup

	sem := make(chan struct{}, pe.threadCount)

	config := DefaultConfigForURLMarkdown()

	for i, url := range urls {
		wg.Add(1)
		go func(idx int, u string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			converter := NewConverterURL(config)
			result, err := converter.Convert(u)
			if err != nil {
				results[idx] = nil
			} else {
				results[idx] = result
			}
		}(i, url)
	}

	wg.Wait()
	return results, nil
}

// ConvertMultipleURLsToMarkdown converts multiple URLs to a markdown string without the ResultURL struct
func (pe *DoclingModel) ConvertMultipleURLsToMarkdown(urls []string) ([]string, error) {
	if pe.threadCount <= 0 {
		pe.threadCount = 4
	}

	results := make([]string, len(urls))
	var wg sync.WaitGroup

	sem := make(chan struct{}, pe.threadCount)

	config := DefaultConfigForURLMarkdown()

	for i, url := range urls {
		wg.Add(1)
		go func(idx int, u string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			converter := NewConverterURL(config)
			result, err := converter.Convert(u)
			if err != nil {
				results[idx] = ""
			} else {
				results[idx] = result.Markdown
			}
		}(i, url)
	}

	wg.Wait()
	return results, nil
}