package main

import (
	"context"
	"log/slog"
	"rag-test/internal/helpers"
	milvusrepo "rag-test/internal/repository/milvus"
	"strings"
)

func processAllFiles(ctx context.Context) error {
	if !needMigration {
		return nil
	}

	var counter = 0
	files, err := listDocumentFiles("documents")
	if err != nil {
		slog.Error("failed to list documents", slog.String("error", err.Error()))
		return err
	}

	for _, file := range files {
		lgr := slog.With(
			slog.String("path", file.Path),
			slog.String("name", file.Name),
		)
		lgr.Info("❗❗processing file❗❗")

		markdown, err := docling.ConvertOneFileToMarkdown(file.Path)
		if err != nil {
			lgr.Error(
				"failed to convert docx to markdown",
				slog.String("error", err.Error()),
			)
			return err
		}

		markdown = strings.ToLower(markdown)

		split, err := helpers.SplitTextByChunks(markdown)
		if err != nil {
			lgr.Error(
				"failed to split text by chunks",
				slog.String("error", err.Error()),
			)
			return err
		}

		embs, err := embedRepo.CreateEmbeddings(ctx, split...)
		if err != nil {
			lgr.Error(
				"failed to create embeddings",
				slog.String("error", err.Error()),
			)
			return err
		}

		var items []milvusrepo.VectorItem
		for i, emb := range embs {
			vi := milvusrepo.VectorItem{
				ID:         int64(counter),
				Embedding:  emb,
				Payload:    split[i],
				DataSource: file.Path,
			}

			items = append(items, vi)
			counter++
		}

		if err = vectorRepo.Upsert(ctx, collectionName, items); err != nil {
			lgr.Error(
				"failed to upsert documents",
				slog.String("error", err.Error()),
			)
			return err
		}

		lgr.Info("✅✅processed file✅✅")
	}

	return nil
}
