package merge

import (
	"bufio"
	"container/heap"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"namefreq/internal/chunk"
)

func Merge(paths []string, handle func(chunk.Record) error) error {
	readers := make([]*chunkReader, 0, len(paths))
	defer func() {
		for _, reader := range readers {
			_ = reader.close()
		}
	}()

	items := make(recordHeap, 0, len(paths))
	for _, path := range paths {
		reader, err := openChunkReader(path)
		if err != nil {
			return err
		}
		readers = append(readers, reader)

		record, ok, err := reader.next()
		if err != nil {
			return err
		}
		if ok {
			heap.Push(&items, heapItem{record: record, readerIndex: len(readers) - 1})
		}
	}

	var current chunk.Record
	hasCurrent := false

	emit := func(record chunk.Record) error {
		if err := handle(record); err != nil {
			return fmt.Errorf("handle merged record %q: %w", record.Name, err)
		}
		return nil
	}

	for items.Len() > 0 {
		item := heap.Pop(&items).(heapItem)
		record := item.record

		if !hasCurrent {
			current = record
			hasCurrent = true
		} else if current.Name == record.Name {
			current.Count += record.Count
		} else {
			if err := emit(current); err != nil {
				return err
			}
			current = record
		}

		next, ok, err := readers[item.readerIndex].next()
		if err != nil {
			return err
		}
		if ok {
			heap.Push(&items, heapItem{record: next, readerIndex: item.readerIndex})
		}
	}

	if hasCurrent {
		if err := emit(current); err != nil {
			return err
		}
	}

	return nil
}

type chunkReader struct {
	path       string
	file       *os.File
	reader     *bufio.Reader
	lineNumber int
}

func openChunkReader(path string) (*chunkReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open chunk file %q: %w", path, err)
	}

	return &chunkReader{
		path:   path,
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}

func (r *chunkReader) next() (chunk.Record, bool, error) {
	line, err := r.reader.ReadString('\n')
	if len(line) > 0 {
		r.lineNumber++
		record, parseErr := chunk.ParseRecord(trimLineBreak(line))
		if parseErr != nil {
			return chunk.Record{}, false, fmt.Errorf("parse chunk file %q line %d: %w", r.path, r.lineNumber, parseErr)
		}
		return record, true, nil
	}

	if err == nil {
		return chunk.Record{}, false, nil
	}
	if errors.Is(err, io.EOF) {
		return chunk.Record{}, false, nil
	}

	return chunk.Record{}, false, fmt.Errorf("read chunk file %q: %w", r.path, err)
}

func (r *chunkReader) close() error {
	if r.file == nil {
		return nil
	}
	return r.file.Close()
}

type heapItem struct {
	record      chunk.Record
	readerIndex int
}

type recordHeap []heapItem

func (h recordHeap) Len() int {
	return len(h)
}

func (h recordHeap) Less(i, j int) bool {
	return h[i].record.Name < h[j].record.Name
}

func (h recordHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *recordHeap) Push(value any) {
	*h = append(*h, value.(heapItem))
}

func (h *recordHeap) Pop() any {
	old := *h
	item := old[len(old)-1]
	*h = old[:len(old)-1]
	return item
}

func trimLineBreak(line string) string {
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	return line
}
