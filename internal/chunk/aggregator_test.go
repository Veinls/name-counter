package chunk

import (
	"errors"
	"strings"
	"testing"
)

func TestAggregateCountsNames(t *testing.T) {
	chunks, err := collectChunks("Alice\nBob\nAlice\n", 1024)
	if err != nil {
		t.Fatalf("Aggregate returned error: %v", err)
	}

	if len(chunks) != 1 {
		t.Fatalf("got %d chunks, want 1", len(chunks))
	}

	assertCounts(t, chunks[0], Counts{
		"Alice": 2,
		"Bob":   1,
	})
}

func TestAggregateTrimsSpacesAndIgnoresEmptyLines(t *testing.T) {
	input := "  Алёна  \n\n\tМиша\t\n   \r\nАлёна\n"

	chunks, err := collectChunks(input, 1024)
	if err != nil {
		t.Fatalf("Aggregate returned error: %v", err)
	}

	if len(chunks) != 1 {
		t.Fatalf("got %d chunks, want 1", len(chunks))
	}

	assertCounts(t, chunks[0], Counts{
		"Алёна": 2,
		"Миша":  1,
	})
}

func TestAggregateFlushesByChunkSize(t *testing.T) {
	chunks, err := collectChunks("aa\nbb\ncc\n", 4)
	if err != nil {
		t.Fatalf("Aggregate returned error: %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("got %d chunks, want 2: %#v", len(chunks), chunks)
	}

	assertCounts(t, chunks[0], Counts{
		"aa": 1,
		"bb": 1,
	})
	assertCounts(t, chunks[1], Counts{
		"cc": 1,
	})
}

func TestAggregateFlushesNameLargerThanChunkSize(t *testing.T) {
	chunks, err := collectChunks("long-name\nx\n", 4)
	if err != nil {
		t.Fatalf("Aggregate returned error: %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("got %d chunks, want 2: %#v", len(chunks), chunks)
	}

	assertCounts(t, chunks[0], Counts{
		"long-name": 1,
	})
	assertCounts(t, chunks[1], Counts{
		"x": 1,
	})
}

func TestAggregateDoesNotEmitEmptyChunk(t *testing.T) {
	chunks, err := collectChunks("\n  \n\t\n", 1024)
	if err != nil {
		t.Fatalf("Aggregate returned error: %v", err)
	}

	if len(chunks) != 0 {
		t.Fatalf("got %d chunks, want 0", len(chunks))
	}
}

func TestAggregateRejectsInvalidChunkSize(t *testing.T) {
	err := Aggregate(strings.NewReader("Alice\n"), 0, func(Counts) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAggregateReturnsHandlerError(t *testing.T) {
	wantErr := errors.New("stop")

	err := Aggregate(strings.NewReader("Alice\n"), 1024, func(Counts) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped handler error, got %v", err)
	}
}

func collectChunks(input string, maxChunkBytes int64) ([]Counts, error) {
	var chunks []Counts

	err := Aggregate(strings.NewReader(input), maxChunkBytes, func(counts Counts) error {
		copyCounts := make(Counts, len(counts))
		for name, count := range counts {
			copyCounts[name] = count
		}
		chunks = append(chunks, copyCounts)
		return nil
	})

	return chunks, err
}

func assertCounts(t *testing.T, got, want Counts) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %d names, want %d: %#v", len(got), len(want), got)
	}

	for name, wantCount := range want {
		if got[name] != wantCount {
			t.Fatalf("count for %q = %d, want %d", name, got[name], wantCount)
		}
	}
}
