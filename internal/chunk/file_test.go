package chunk

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWriteTempFileWritesSortedRecords(t *testing.T) {
	tmpDir := t.TempDir()

	path, err := WriteTempFile(tmpDir, Counts{
		"Миша":  2,
		"Алёна": 3,
		"Дима":  1,
	})
	if err != nil {
		t.Fatalf("WriteTempFile returned error: %v", err)
	}
	defer os.Remove(path)

	if filepath.Dir(path) != tmpDir {
		t.Fatalf("chunk file dir = %q, want %q", filepath.Dir(path), tmpDir)
	}

	gotBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read chunk file: %v", err)
	}

	want := "Алёна\t3\nДима\t1\nМиша\t2\n"
	if string(gotBytes) != want {
		t.Fatalf("chunk file = %q, want %q", string(gotBytes), want)
	}
}

func TestReadFileReadsRecords(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chunk.txt")
	if err := os.WriteFile(path, []byte("Alice\t2\nBob\t1\n"), 0o644); err != nil {
		t.Fatalf("write chunk fixture: %v", err)
	}

	got, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	want := []Record{
		{Name: "Alice", Count: 2},
		{Name: "Bob", Count: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("records = %#v, want %#v", got, want)
	}
}

func TestReadFileRejectsInvalidRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chunk.txt")
	if err := os.WriteFile(path, []byte("Alice two\n"), 0o644); err != nil {
		t.Fatalf("write chunk fixture: %v", err)
	}

	if _, err := ReadFile(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestSortedRecordsSortsByName(t *testing.T) {
	got := SortedRecords(Counts{
		"Charlie": 1,
		"Alice":   3,
		"Bob":     2,
	})

	want := []Record{
		{Name: "Alice", Count: 3},
		{Name: "Bob", Count: 2},
		{Name: "Charlie", Count: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("records = %#v, want %#v", got, want)
	}
}

func TestRemoveFilesRemovesExistingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "chunk.txt")
	if err := os.WriteFile(path, []byte("Alice\t1\n"), 0o644); err != nil {
		t.Fatalf("write chunk fixture: %v", err)
	}

	if err := RemoveFiles([]string{path}); err != nil {
		t.Fatalf("RemoveFiles returned error: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected file to be removed, stat error: %v", err)
	}
}
