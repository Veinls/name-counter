package counter

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

func ReadLines(r io.Reader, handle func(string) error) error {
	reader := bufio.NewReader(r)

	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			if err := handle(trimLineBreak(line)); err != nil {
				return fmt.Errorf("handle input line: %w", err)
			}
		}

		if err == nil {
			continue
		}
		if errors.Is(err, io.EOF) {
			return nil
		}

		return fmt.Errorf("read input line: %w", err)
	}
}

func trimLineBreak(line string) string {
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	return line
}
