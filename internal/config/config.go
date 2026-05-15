package config

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	SortByName = "name"
	SortByFreq = "freq"

	defaultChunkSizeText = "128MB"
	defaultTempDir       = "/tmp/namefreq"
)

type Config struct {
	InputPath     string
	SortMode      string
	ChunkSizeText string
	ChunkSize     int64
	TempDir       string
}

func Default() Config {
	return Config{
		SortMode:      SortByName,
		ChunkSizeText: defaultChunkSizeText,
		TempDir:       defaultTempDir,
	}
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.InputPath) == "" {
		return fmt.Errorf("input path is required")
	}

	switch c.SortMode {
	case SortByName, SortByFreq:
	default:
		return fmt.Errorf("invalid sort mode %q: expected %q or %q", c.SortMode, SortByName, SortByFreq)
	}

	chunkSize, err := ParseSize(c.ChunkSizeText)
	if err != nil {
		return fmt.Errorf("invalid chunk size %q: %w", c.ChunkSizeText, err)
	}
	if chunkSize <= 0 {
		return fmt.Errorf("chunk size must be positive")
	}
	c.ChunkSize = chunkSize

	if strings.TrimSpace(c.TempDir) == "" {
		return fmt.Errorf("temporary directory is required")
	}

	return nil
}

func ParseSize(value string) (int64, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return 0, fmt.Errorf("size is empty")
	}

	unitMultiplier := int64(1)
	numberText := text
	upper := strings.ToUpper(text)

	units := []struct {
		suffix     string
		multiplier int64
	}{
		{suffix: "GB", multiplier: 1024 * 1024 * 1024},
		{suffix: "MB", multiplier: 1024 * 1024},
		{suffix: "KB", multiplier: 1024},
		{suffix: "B", multiplier: 1},
	}

	for _, unit := range units {
		suffix := unit.suffix
		if strings.HasSuffix(upper, suffix) {
			unitMultiplier = unit.multiplier
			numberText = strings.TrimSpace(text[:len(text)-len(suffix)])
			break
		}
	}

	number, err := strconv.ParseInt(numberText, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("expected integer size with optional B, KB, MB, or GB suffix")
	}
	if number <= 0 {
		return 0, fmt.Errorf("size must be positive")
	}
	if number > (1<<63-1)/unitMultiplier {
		return 0, fmt.Errorf("size is too large")
	}

	return number * unitMultiplier, nil
}
