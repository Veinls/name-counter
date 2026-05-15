package chunk

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"namefreq/internal/counter"
)

type Record struct {
	Name  string
	Count int64
}

func WriteTempFile(tmpDir string, counts Counts) (string, error) {
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", fmt.Errorf("create temporary directory: %w", err)
	}

	file, err := os.CreateTemp(tmpDir, "namefreq-*.chunk")
	if err != nil {
		return "", fmt.Errorf("create temporary chunk file: %w", err)
	}

	path := file.Name()
	removeOnError := true
	defer func() {
		if removeOnError {
			_ = os.Remove(path)
		}
	}()

	writer := bufio.NewWriter(file)
	records := SortedRecords(counts)
	for _, record := range records {
		if _, err := fmt.Fprintf(writer, "%s\t%d\n", record.Name, record.Count); err != nil {
			_ = file.Close()
			return "", fmt.Errorf("write temporary chunk file %q: %w", path, err)
		}
	}

	if err := writer.Flush(); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("flush temporary chunk file %q: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close temporary chunk file %q: %w", path, err)
	}

	removeOnError = false
	return path, nil
}

func ReadFile(path string) ([]Record, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open chunk file %q: %w", path, err)
	}
	defer file.Close()

	var records []Record
	lineNumber := 0
	if err := counter.ReadLines(file, func(line string) error {
		lineNumber++
		record, err := parseRecord(line)
		if err != nil {
			return fmt.Errorf("parse chunk file %q line %d: %w", path, lineNumber, err)
		}
		records = append(records, record)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("read chunk file %q: %w", path, err)
	}

	return records, nil
}

func SortedRecords(counts Counts) []Record {
	records := make([]Record, 0, len(counts))
	for name, count := range counts {
		records = append(records, Record{Name: name, Count: count})
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Name < records[j].Name
	})

	return records
}

func parseRecord(line string) (Record, error) {
	name, countText, ok := strings.Cut(line, "\t")
	if !ok {
		return Record{}, fmt.Errorf("expected name and count separated by tab")
	}
	if name == "" {
		return Record{}, fmt.Errorf("name is empty")
	}

	count, err := strconv.ParseInt(countText, 10, 64)
	if err != nil {
		return Record{}, fmt.Errorf("invalid count %q: %w", countText, err)
	}
	if count < 0 {
		return Record{}, fmt.Errorf("count must be non-negative")
	}

	return Record{Name: name, Count: count}, nil
}

func RemoveFiles(paths []string) error {
	var errs []error
	for _, path := range paths {
		if path == "" {
			continue
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("remove temporary chunk file %q: %w", filepath.Clean(path), err))
		}
	}

	if len(errs) == 1 {
		return errs[0]
	}
	if len(errs) > 1 {
		return fmt.Errorf("%d errors occurred while removing temporary chunk files: %v", len(errs), errs)
	}

	return nil
}
