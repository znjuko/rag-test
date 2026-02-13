package docling_bridge

import (
	"fmt"
	"os"

	"github.com/Dsouza10082/go-docling-bridge/model"
)

type DoclingBridge struct {
	Instance *model.DoclingModel
}

func NewDoclingBridge() *DoclingBridge {
	inst := model.NewDoclingModel()
	return &DoclingBridge{
		Instance: inst,
	}
}

func (db *DoclingBridge) SetThreadCount(count int) *DoclingBridge {
	db.Instance.SetThreadCount(count)
	return db
}

func (db *DoclingBridge) SetDocumentPath(path string) *DoclingBridge {
	db.Instance.SetDocumentPath(path)
	return db
}

func (db *DoclingBridge) ReadDirectory() *DoclingBridge {
	db.Instance.ReadDirectory()
	return db
}

func (db *DoclingBridge) Start() *DoclingBridge {
	db.Instance.Start()
	return db
}

func (db *DoclingBridge) GetResults() *DoclingBridge {
	db.Instance.GetResults()
	return db
}

func (db *DoclingBridge) StartWithProgress() *DoclingBridge {
	db.Instance.StartWithProgress()
	return db
}

func (db *DoclingBridge) SetIsAdvanced(isAdvanced bool) *DoclingBridge {
	db.Instance.SetIsAdvanced(isAdvanced)
	return db
}

func (db *DoclingBridge) SetMaxTokens(tokens int) *DoclingBridge {
	db.Instance.SetMaxTokens(tokens)
	return db
}

func (db *DoclingBridge) ProcessBatch(startIdx, endIdx int) *DoclingBridge {
	db.Instance.ProcessBatch(startIdx, endIdx)
	return db
}

func (db *DoclingBridge) SetOutputPath(path string) *DoclingBridge {
	db.Instance.SetOutputPath(path)
	return db
}

func (db *DoclingBridge) SetIsAudio(isAudio bool) *DoclingBridge {
	db.Instance.SetIsAudio(isAudio)
	return db
}

func (db *DoclingBridge) ConvertOneFileToMarkdown(filePath string) (string, error) {
	return db.Instance.ConvertOneFileToMarkdown(filePath)
}

func (db *DoclingBridge) ConvertMultipleURLsToMarkdown(urls []string) ([]string, error) {
	return db.Instance.ConvertMultipleURLsToMarkdown(urls)
}

func (db *DoclingBridge) ConvertURLToMarkdown(url string) (string, error) {
	return db.Instance.ConvertURLToMarkdown(url)
}

func (db *DoclingBridge) ConvertURLToMarkdownComplete(url string) (*model.ResultURL, error) {
	return db.Instance.ConvertURLToMarkdownComplete(url)
}

func (db *DoclingBridge) ConvertMultipleURLsToMarkdownComplete(urls []string) ([]*model.ResultURL, error) {
	return db.Instance.ConvertMultipleURLsToMarkdownComplete(urls)
}

func main() {

	// Check if the documents directory exists otherwise create it
	if _, err := os.Stat("documents"); os.IsNotExist(err) {
		fmt.Println("Documents directory does not exist")
		os.Mkdir("documents", 0755)
	}

	// Check if the output directory exists otherwise create it
	if _, err := os.Stat("output"); os.IsNotExist(err) {
		fmt.Println("Output directory does not exist")
		os.Mkdir("output", 0755)
	}

	fmt.Println("Initializing Docling Handler")
	fmt.Println("--------------------------------")
	NewDoclingBridge().
		SetDocumentPath("./documents").
		SetOutputPath("./output/transcript.md").
		SetIsAudio(true).
		SetMaxTokens(512).
		SetThreadCount(8).
		ReadDirectory().
		StartWithProgress()

}
