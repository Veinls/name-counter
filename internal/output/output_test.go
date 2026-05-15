package output

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"namefreq/internal/chunk"
	"namefreq/internal/config"
)

func TestWriteNameSortedKeepsInputOrder(t *testing.T) {
	records := []chunk.Record{
		{Name: "Алёна", Count: 2},
		{Name: "Дима", Count: 1},
		{Name: "Миша", Count: 3},
	}

	var out bytes.Buffer
	if err := WriteNameSorted(&out, records); err != nil {
		t.Fatalf("WriteNameSorted returned error: %v", err)
	}

	want := "Алёна\t2\nДима\t1\nМиша\t3\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestWriteFreqSortedSortsByCountDescAndNameAsc(t *testing.T) {
	records := []chunk.Record{
		{Name: "Charlie", Count: 2},
		{Name: "Alice", Count: 3},
		{Name: "Bob", Count: 3},
		{Name: "Dave", Count: 1},
	}

	var out bytes.Buffer
	if err := WriteFreqSorted(&out, records); err != nil {
		t.Fatalf("WriteFreqSorted returned error: %v", err)
	}

	want := "Alice\t3\nBob\t3\nCharlie\t2\nDave\t1\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestWriteByModeUsesNameSort(t *testing.T) {
	var out bytes.Buffer
	err := WriteByMode(&out, []chunk.Record{{Name: "Alice", Count: 1}}, config.SortByName)
	if err != nil {
		t.Fatalf("WriteByMode returned error: %v", err)
	}

	if out.String() != "Alice\t1\n" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestWriteByModeUsesFreqSort(t *testing.T) {
	records := []chunk.Record{
		{Name: "Alice", Count: 1},
		{Name: "Bob", Count: 2},
	}

	var out bytes.Buffer
	err := WriteByMode(&out, records, config.SortByFreq)
	if err != nil {
		t.Fatalf("WriteByMode returned error: %v", err)
	}

	if out.String() != "Bob\t2\nAlice\t1\n" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestWriteByModeRejectsUnsupportedSort(t *testing.T) {
	err := WriteByMode(io.Discard, nil, "unknown")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWriteRecordReturnsWriterError(t *testing.T) {
	wantErr := errors.New("write failed")
	writer := failingWriter{err: wantErr}

	err := WriteRecord(writer, chunk.Record{Name: "Alice", Count: 1})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped writer error, got %v", err)
	}
}

func TestWriteFreqSortedDoesNotMutateInput(t *testing.T) {
	records := []chunk.Record{
		{Name: "Alice", Count: 1},
		{Name: "Bob", Count: 2},
	}

	var out bytes.Buffer
	if err := WriteFreqSorted(&out, records); err != nil {
		t.Fatalf("WriteFreqSorted returned error: %v", err)
	}

	gotOrder := []string{records[0].Name, records[1].Name}
	wantOrder := []string{"Alice", "Bob"}
	if strings.Join(gotOrder, ",") != strings.Join(wantOrder, ",") {
		t.Fatalf("records were mutated: %#v", records)
	}
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}
