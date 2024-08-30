package server

import (
	"fmt"
	"os"
	"path/filepath"

	"go.lsp.dev/uri"
)

type embeddableResp struct {
	embeddables []embeddable
}
type embeddable struct {
	name string
	data []byte
}

func getEmbbeddables(
	uri uri.URI,
	curVal string,
	errCh chan<- error,
) <-chan embeddableResp {
	respCh := make(chan embeddableResp)
	dir := filepath.Dir(uri.Filename())
	entries, err := os.ReadDir(dir)
	if err != nil {
		errCh <- fmt.Errorf("error reading directory: %w", err)
	}
	embeddables := make([]embeddable, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			errCh <- fmt.Errorf("error reading file: %w", err)
		}
		embeddables = append(embeddables, embeddable{
			name: entry.Name(),
			data: data,
		})
	}
	respCh <- embeddableResp{
		embeddables: embeddables,
	}
	return respCh
}
