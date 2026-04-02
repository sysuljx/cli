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

// Save writes body to path atomically after validating the output path.
// Parent directories are created as needed.
func (l *LocalFileIO) Save(path string, _ fileio.SaveOptions, body io.Reader) error {
	safePath, err := validate.SafeOutputPath(path)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(safePath), 0755); err != nil {
		return err
	}
	return validate.AtomicWrite(safePath, data, 0644)
}
