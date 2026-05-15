package merge

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"namefreq/internal/chunk"
)

func TestMergeOneChunk(t *testing.T) {
	paths := writeChunks(t, []chunk.Counts{
		{"Alice": 2, "Bob": 1},
	})

	got, err := collectMerge(paths)
	if err != nil {
		t.Fatalf("Merge returned error: %v", err)
	}

	want := []chunk.Record{
		{Name: "Alice", Count: 2},
		{Name: "Bob", Count: 1},
	}
	assertRecords(t, got, want)
}

func TestMergeTwoChunks(t *testing.T) {
	paths := writeChunks(t, []chunk.Counts{
		{"Alice": 2, "Charlie": 1},
		{"Alice": 3, "Bob": 4},
	})

	got, err := collectMerge(paths)
	if err != nil {
		t.Fatalf("Merge returned error: %v", err)
	}

	want := []chunk.Record{
		{Name: "Alice", Count: 5},
		{Name: "Bob", Count: 4},
		{Name: "Charlie", Count: 1},
	}
	assertRecords(t, got, want)
}

func TestMergeMultipleChunks(t *testing.T) {
	paths := writeChunks(t, []chunk.Counts{
		{"Алёна": 1, "Миша": 2},
		{"Дима": 4, "Миша": 1},
		{"Алёна": 3, "Яна": 5},
	})

	got, err := collectMerge(paths)
	if err != nil {
		t.Fatalf("Merge returned error: %v", err)
	}

	want := []chunk.Record{
		{Name: "Алёна", Count: 4},
		{Name: "Дима", Count: 4},
		{Name: "Миша", Count: 3},
		{Name: "Яна", Count: 5},
	}
	assertRecords(t, got, want)
}

func TestMergeEmptyInput(t *testing.T) {
	got, err := collectMerge(nil)
	if err != nil {
		t.Fatalf("Merge returned error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("got %#v, want empty result", got)
	}
}

func TestMergeReturnsHandlerError(t *testing.T) {
	paths := writeChunks(t, []chunk.Counts{
		{"Alice": 1},
	})
	wantErr := errors.New("stop")

	err := Merge(paths, func(chunk.Record) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped handler error, got %v", err)
	}
}

func TestMergeRejectsInvalidChunk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.chunk")
	if err := os.WriteFile(path, []byte("Alice 1\n"), 0o644); err != nil {
		t.Fatalf("write bad chunk: %v", err)
	}

	if _, err := collectMerge([]string{path}); err == nil {
		t.Fatal("expected error")
	}
}

func writeChunks(t *testing.T, chunks []chunk.Counts) []string {
	t.Helper()

	tmpDir := t.TempDir()
	paths := make([]string, 0, len(chunks))
	for _, counts := range chunks {
		path, err := chunk.WriteTempFile(tmpDir, counts)
		if err != nil {
			t.Fatalf("WriteTempFile returned error: %v", err)
		}
		paths = append(paths, path)
	}

	return paths
}

func collectMerge(paths []string) ([]chunk.Record, error) {
	var records []chunk.Record

	err := Merge(paths, func(record chunk.Record) error {
		records = append(records, record)
		return nil
	})

	return records, err
}

func assertRecords(t *testing.T, got, want []chunk.Record) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("records = %#v, want %#v", got, want)
	}
}
