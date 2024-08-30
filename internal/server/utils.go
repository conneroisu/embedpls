package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/conneroisu/embedpls/internal/lsp"
	"github.com/conneroisu/embedpls/internal/parsers"
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

func (l *lspHandler) getHoverResp(req lsp.HoverRequest, errCh chan<- error) <-chan lsp.HoverResult {
	respCh := make(chan lsp.HoverResult)
	go func() {
		doc, ok := l.documents.Get(req.Params.TextDocument.URI)
		if !ok {
			errCh <- fmt.Errorf("document not found")
			return
		}
		curVal, state, err := parsers.ParseSourcePosition(
			doc,
			req.Params.Position,
		)
		if err != nil {
			errCh <- err
		}
		if state == parsers.StateUnknown {
			errCh <- nil
			return
		}
		content, err := relativeReadFile(req.Params.TextDocument.URI, curVal)
		if err != nil {
			errCh <- err
			return
		}
		respCh <- lsp.HoverResult{
			Contents: content,
		}
	}()
	return respCh
}

func relativeReadFile(uri uri.URI, embedPath string) (string, error) {
	dir := filepath.Dir(uri.Filename())
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("error reading directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), embedPath) {
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				return "", fmt.Errorf("error reading file: %w", err)
			}
			log.Debugf("found file: %s", entry.Name())
			log.Debugf("file content: %s", string(data))
			return string(data), nil
		}
	}
	return "", fmt.Errorf("file not found")
}
