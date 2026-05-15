package counter

import (
	"errors"
	"strings"
	"testing"
)

func TestReadLinesReadsAllLines(t *testing.T) {
	input := "Алёна\nМиша\nДима\n"

	got, err := collectLines(input)
	if err != nil {
		t.Fatalf("ReadLines returned error: %v", err)
	}

	want := []string{"Алёна", "Миша", "Дима"}
	assertLines(t, got, want)
}

func TestReadLinesReadsLastLineWithoutLineBreak(t *testing.T) {
	input := "one\ntwo\nthree"

	got, err := collectLines(input)
	if err != nil {
		t.Fatalf("ReadLines returned error: %v", err)
	}

	want := []string{"one", "two", "three"}
	assertLines(t, got, want)
}

func TestReadLinesSupportsCRLF(t *testing.T) {
	input := "one\r\ntwo\r\nthree\r\n"

	got, err := collectLines(input)
	if err != nil {
		t.Fatalf("ReadLines returned error: %v", err)
	}

	want := []string{"one", "two", "three"}
	assertLines(t, got, want)
}

func TestReadLinesSupportsLongLines(t *testing.T) {
	longName := strings.Repeat("а", 1024*1024)
	input := longName + "\nshort\n"

	got, err := collectLines(input)
	if err != nil {
		t.Fatalf("ReadLines returned error: %v", err)
	}

	want := []string{longName, "short"}
	assertLines(t, got, want)
}

func TestReadLinesReturnsHandlerError(t *testing.T) {
	wantErr := errors.New("stop")

	err := ReadLines(strings.NewReader("one\ntwo\n"), func(line string) error {
		if line == "two" {
			return wantErr
		}

		return nil
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped handler error, got %v", err)
	}
}

func collectLines(input string) ([]string, error) {
	var lines []string

	err := ReadLines(strings.NewReader(input), func(line string) error {
		lines = append(lines, line)
		return nil
	})

	return lines, err
}

func assertLines(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d: %#v", len(got), len(want), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}
