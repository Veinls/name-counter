package config

import "testing"

func TestValidateAcceptsValidConfig(t *testing.T) {
	cfg := Config{
		InputPath:     "input.txt",
		SortMode:      SortByName,
		ChunkSizeText: "128MB",
		TempDir:       "/tmp/namefreq",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if cfg.ChunkSize != 128*1024*1024 {
		t.Fatalf("ChunkSize = %d, want %d", cfg.ChunkSize, 128*1024*1024)
	}
}

func TestValidateRejectsInvalidSortMode(t *testing.T) {
	cfg := Config{
		InputPath:     "input.txt",
		SortMode:      "unknown",
		ChunkSizeText: "128MB",
		TempDir:       "/tmp/namefreq",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateRejectsMissingInputPath(t *testing.T) {
	cfg := Config{
		SortMode:      SortByName,
		ChunkSizeText: "128MB",
		TempDir:       "/tmp/namefreq",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateRejectsInvalidChunkSize(t *testing.T) {
	cfg := Config{
		InputPath:     "input.txt",
		SortMode:      SortByName,
		ChunkSizeText: "zero",
		TempDir:       "/tmp/namefreq",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{name: "bytes without suffix", input: "42", want: 42},
		{name: "bytes with suffix", input: "42B", want: 42},
		{name: "kilobytes", input: "2KB", want: 2 * 1024},
		{name: "megabytes", input: "3MB", want: 3 * 1024 * 1024},
		{name: "gigabytes", input: "4GB", want: 4 * 1024 * 1024 * 1024},
		{name: "lowercase suffix", input: "5mb", want: 5 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSize(tt.input)
			if err != nil {
				t.Fatalf("ParseSize returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseSizeRejectsInvalidValues(t *testing.T) {
	tests := []string{"", "0", "-1", "1.5MB", "MB"}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if _, err := ParseSize(input); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
