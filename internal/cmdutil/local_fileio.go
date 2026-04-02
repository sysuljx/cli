// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package cmdutil

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/larksuite/cli/extension/fileio"
	"github.com/larksuite/cli/internal/validate"
)

// localFileIOProvider is the default fileio.Provider backed by the local filesystem.
type localFileIOProvider struct{}

func (p *localFileIOProvider) Name() string { return "local" }

func (p *localFileIOProvider) ResolveFileIO(_ context.Context) fileio.FileIO {
	return &LocalFileIO{}
}

func init() {
	fileio.Register(&localFileIOProvider{})
}

// LocalFileIO implements fileio.FileIO using the local filesystem.
// Path validation (SafeInputPath/SafeOutputPath), directory creation,
// and atomic writes are handled internally.
type LocalFileIO struct{}

// Open opens a local file for reading after validating the path.
func (l *LocalFileIO) Open(name string) (fileio.File, error) {
	safePath, err := validate.SafeInputPath(name)
	if err != nil {
		return nil, err
	}
	return os.Open(safePath)
}

// Stat returns file metadata after validating the path.
func (l *LocalFileIO) Stat(name string) (os.FileInfo, error) {
	safePath, err := validate.SafeInputPath(name)
	if err != nil {
		return nil, err
	}
	return os.Stat(safePath)
}

// localSaveResult implements fileio.SaveResult.
type localSaveResult struct{ size int64 }

func (r *localSaveResult) Size() int64 { return r.size }

// Save writes body to path atomically after validating the output path.
// Parent directories are created as needed. The body is streamed directly
// to a temp file and renamed, avoiding full in-memory buffering.
func (l *LocalFileIO) Save(path string, _ fileio.SaveOptions, body io.Reader) (fileio.SaveResult, error) {
	safePath, err := validate.SafeOutputPath(path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(safePath), 0755); err != nil {
		return nil, err
	}
	n, err := validate.AtomicWriteFromReader(safePath, body, 0644)
	if err != nil {
		return nil, err
	}
	return &localSaveResult{size: n}, nil
}
