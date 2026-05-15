package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"namefreq/internal/app"
	"namefreq/internal/config"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "namefreq: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("namefreq", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	cfg := config.Default()
	fs.StringVar(&cfg.SortMode, "sort", cfg.SortMode, "sort result by name or freq")
	fs.StringVar(&cfg.ChunkSizeText, "chunk-size", cfg.ChunkSizeText, "maximum in-memory chunk size, for example 128MB")
	fs.StringVar(&cfg.TempDir, "tmp-dir", cfg.TempDir, "directory for temporary chunk files")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return errors.New("usage: namefreq [--sort=name|freq] [--chunk-size=128MB] [--tmp-dir=/tmp/namefreq] input.txt")
	}

	cfg.InputPath = fs.Arg(0)
	if err := cfg.Validate(); err != nil {
		return err
	}

	return app.Run(cfg, os.Stdout)
}
