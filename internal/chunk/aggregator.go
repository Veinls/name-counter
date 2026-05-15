package chunk

import (
	"fmt"
	"io"
	"strings"

	"namefreq/internal/counter"
)

type Counts map[string]int64

func Aggregate(r io.Reader, maxChunkBytes int64, handle func(Counts) error) error {
	if maxChunkBytes <= 0 {
		return fmt.Errorf("max chunk size must be positive")
	}

	current := make(Counts)
	var currentBytes int64

	flush := func() error {
		if len(current) == 0 {
			currentBytes = 0
			return nil
		}

		if err := handle(current); err != nil {
			return fmt.Errorf("handle chunk: %w", err)
		}

		current = make(Counts)
		currentBytes = 0
		return nil
	}

	if err := counter.ReadLines(r, func(line string) error {
		name := strings.TrimSpace(line)
		if name == "" {
			return nil
		}

		if currentBytes > 0 && currentBytes+int64(len(name)) > maxChunkBytes {
			if err := flush(); err != nil {
				return err
			}
		}

		current[name]++
		currentBytes += int64(len(name))

		if currentBytes >= maxChunkBytes {
			return flush()
		}

		return nil
	}); err != nil {
		return fmt.Errorf("aggregate input: %w", err)
	}

	if err := flush(); err != nil {
		return err
	}

	return nil
}
