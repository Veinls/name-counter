package app

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"namefreq/internal/chunk"
	"namefreq/internal/config"
	"namefreq/internal/merge"
	"namefreq/internal/output"
)

func Run(cfg config.Config, out io.Writer) (runErr error) {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	input, err := os.Open(cfg.InputPath)
	if err != nil {
		return fmt.Errorf("open input file %q: %w", cfg.InputPath, err)
	}
	defer input.Close()

	var chunkPaths []string
	defer func() {
		if cleanupErr := chunk.RemoveFiles(chunkPaths); cleanupErr != nil {
			if runErr != nil {
				runErr = fmt.Errorf("%w; cleanup failed: %v", runErr, cleanupErr)
				return
			}
			runErr = cleanupErr
		}
	}()

	if err := chunk.Aggregate(input, cfg.ChunkSize, func(counts chunk.Counts) error {
		path, err := chunk.WriteTempFile(cfg.TempDir, counts)
		if err != nil {
			return fmt.Errorf("write chunk: %w", err)
		}
		chunkPaths = append(chunkPaths, path)
		return nil
	}); err != nil {
		return fmt.Errorf("aggregate chunks: %w", err)
	}

	switch cfg.SortMode {
	case config.SortByName:
		writer := bufio.NewWriter(out)
		if err := merge.Merge(chunkPaths, func(record chunk.Record) error {
			return output.WriteRecord(writer, record)
		}); err != nil {
			return fmt.Errorf("merge chunks: %w", err)
		}
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("flush output: %w", err)
		}
	case config.SortByFreq:
		var records []chunk.Record
		if err := merge.Merge(chunkPaths, func(record chunk.Record) error {
			records = append(records, record)
			return nil
		}); err != nil {
			return fmt.Errorf("merge chunks: %w", err)
		}
		if err := output.WriteFreqSorted(out, records); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	default:
		return fmt.Errorf("unsupported sort mode %q", cfg.SortMode)
	}

	return nil
}
