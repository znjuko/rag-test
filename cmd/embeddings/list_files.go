package main

import (
	"io/fs"
	"path/filepath"
)

type documentFile struct {
	Name string
	Path string
}

func listDocumentFiles(documentsDir string) ([]documentFile, error) {
	absDir, err := filepath.Abs(documentsDir)
	if err != nil {
		return nil, err
	}

	files := make([]documentFile, 0)
	err = filepath.WalkDir(absDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		files = append(files, documentFile{
			Name: entry.Name(),
			Path: path,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
