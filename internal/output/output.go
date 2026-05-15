package output

import (
	"bufio"
	"fmt"
	"io"
	"sort"

	"namefreq/internal/chunk"
	"namefreq/internal/config"
)

func WriteNameSorted(w io.Writer, records []chunk.Record) error {
	writer := bufio.NewWriter(w)

	for _, record := range records {
		if err := WriteRecord(writer, record); err != nil {
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush output: %w", err)
	}

	return nil
}

func WriteFreqSorted(w io.Writer, records []chunk.Record) error {
	sorted := append([]chunk.Record(nil), records...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Count != sorted[j].Count {
			return sorted[i].Count > sorted[j].Count
		}

		return sorted[i].Name < sorted[j].Name
	})

	return WriteNameSorted(w, sorted)
}

func WriteByMode(w io.Writer, records []chunk.Record, sortMode string) error {
	switch sortMode {
	case config.SortByName:
		return WriteNameSorted(w, records)
	case config.SortByFreq:
		return WriteFreqSorted(w, records)
	default:
		return fmt.Errorf("unsupported sort mode %q", sortMode)
	}
}

func WriteRecord(w io.Writer, record chunk.Record) error {
	if _, err := fmt.Fprintf(w, "%s\t%d\n", record.Name, record.Count); err != nil {
		return fmt.Errorf("write output record %q: %w", record.Name, err)
	}

	return nil
}
