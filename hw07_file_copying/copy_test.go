package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopy(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	content := []byte("This is a test file content for copying")

	if err := os.WriteFile(srcPath, content, 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		offset  int64
		limit   int64
		want    []byte
		wantErr error
		setup   func() // Дополнительная настройка
	}{
		{
			name: "successful copy full file",
			want: content,
		},
		{
			name:   "copy with offset",
			offset: 5,
			want:   content[5:],
		},
		{
			name:  "copy with limit",
			limit: 10,
			want:  content[:10],
		},
		{
			name:   "copy with offset and limit",
			offset: 5,
			limit:  10,
			want:   content[5:15],
		},
		{
			name:    "offset exceeds file size",
			offset:  int64(len(content) + 1),
			wantErr: ErrOffsetExceedsFileSize,
		},
		{
			name:    "negative offset",
			offset:  -1,
			wantErr: ErrNegativeOffset,
		},
		{
			name:    "negative limit",
			limit:   -1,
			wantErr: ErrNegativeLimit,
		},
		{
			name:    "nonexistent source file",
			setup:   func() { srcPath = filepath.Join(tmpDir, "nonexistent.txt") },
			wantErr: os.ErrNotExist,
		},
		{
			name:    "copy to the same file",
			setup:   func() { srcPath = filepath.Join(tmpDir, "copy_"+"copy to the same file"+".txt") },
			wantErr: ErrCopyToSameFile,
		},
		{
			name: "unsupported file type",
			setup: func() {
				srcPath = filepath.Join(tmpDir, "testdir")
				os.Mkdir(srcPath, 0o755)
			},
			wantErr: ErrUnsupportedFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			dstPath := filepath.Join(tmpDir, "copy_"+tt.name+".txt")
			err := Copy(srcPath, dstPath, tt.offset, tt.limit)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Copy() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Copy() unexpected error = %v", err)
				return
			}

			copied, err := os.ReadFile(dstPath)
			if err != nil {
				t.Fatalf("Failed to read copied file: %v", err)
			}

			if !bytes.Equal(tt.want, copied) {
				t.Errorf("Content mismatch, got %q, want %q", copied, tt.want)
			}
		})
	}
}
