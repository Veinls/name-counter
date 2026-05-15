package app

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"namefreq/internal/config"
)

func TestRunSortsByName(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := writeInput(t, tmpDir, "Миша\nАлёна\nМиша\n  Дима  \n\n")

	var out bytes.Buffer
	if err := Run(testConfig(inputPath, tmpDir, config.SortByName, "8B"), &out); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := "Алёна\t1\nДима\t1\nМиша\t2\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
	assertNoChunkFiles(t, tmpDir)
}

func TestRunSortsByFreq(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := writeInput(t, tmpDir, "Bob\nAlice\nBob\nCharlie\nAlice\nBob\n")

	var out bytes.Buffer
	if err := Run(testConfig(inputPath, tmpDir, config.SortByFreq, "5B"), &out); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	want := "Bob\t3\nAlice\t2\nCharlie\t1\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
	assertNoChunkFiles(t, tmpDir)
}

func TestRunOutputsNothingForEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := writeInput(t, tmpDir, "\n  \n")

	var out bytes.Buffer
	if err := Run(testConfig(inputPath, tmpDir, config.SortByName, "128MB"), &out); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if out.String() != "" {
		t.Fatalf("output = %q, want empty output", out.String())
	}
	assertNoChunkFiles(t, tmpDir)
}

func TestRunReturnsInvalidInputPathError(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := testConfig(filepath.Join(tmpDir, "missing.txt"), tmpDir, config.SortByName, "128MB")

	err := Run(cfg, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunCleansChunksAfterOutputError(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := writeInput(t, tmpDir, "Alice\nBob\n")
	wantErr := errors.New("write failed")

	err := Run(testConfig(inputPath, tmpDir, config.SortByName, "1B"), failingWriter{err: wantErr})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped writer error, got %v", err)
	}
	assertNoChunkFiles(t, tmpDir)
}

func testConfig(inputPath, tmpDir, sortMode, chunkSize string) config.Config {
	return config.Config{
		InputPath:     inputPath,
		SortMode:      sortMode,
		ChunkSizeText: chunkSize,
		TempDir:       tmpDir,
	}
}

func writeInput(t *testing.T, tmpDir, content string) string {
	t.Helper()

	path := filepath.Join(tmpDir, "input.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write input fixture: %v", err)
	}

	return path
}

func assertNoChunkFiles(t *testing.T, tmpDir string) {
	t.Helper()

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("read temp dir: %v", err)
	}

	var chunkFiles []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "namefreq-") && strings.HasSuffix(entry.Name(), ".chunk") {
			chunkFiles = append(chunkFiles, entry.Name())
		}
	}
	if len(chunkFiles) > 0 {
		t.Fatalf("temporary chunk files were not removed: %v", chunkFiles)
	}
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}
