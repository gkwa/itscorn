package itscorn

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/spf13/afero"
)

var opts struct {
	LogFormat string `long:"log-format" choice:"text" choice:"json" default:"text" description:"Log format"`
	Verbose   []bool `short:"v" long:"verbose" description:"Show verbose debug information, each -v bumps log level"`
	logLevel  slog.Level

	FilesystemType string `short:"f" long:"filesystem" default:"os" description:"Type of filesystem (os, mem)"`
}

var parser *flags.Parser

func Execute() int {
	parser = flags.NewParser(&opts, flags.HelpFlag)

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			parser.WriteHelp(os.Stdout)
			return 0
		}

		parser.WriteHelp(os.Stderr)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		return 1
	}

	if err := setLogLevel(); err != nil {
		slog.Error("error setting log level", "error", err)
		return 1
	}

	if err := setupLogger(); err != nil {
		slog.Error("error setting up logger", "error", err)
		return 1
	}

	if err := run(); err != nil {
		slog.Error("run failed", "error", err)
		return 1
	}

	return 0
}

// Filesystem interface defines the methods for filesystem operations
type Filesystem interface {
	WriteFile(string, []byte, os.FileMode) error
	ReadFile(string) ([]byte, error)
}

// RealFilesystem implements the Filesystem interface using the real OS filesystem
type RealFilesystem struct {
	aferoFs afero.Fs
}

func (r *RealFilesystem) WriteFile(filePath string, content []byte, perm os.FileMode) error {
	return afero.WriteFile(r.aferoFs, filePath, content, perm)
}

func (r *RealFilesystem) ReadFile(filePath string) ([]byte, error) {
	return afero.ReadFile(r.aferoFs, filePath)
}

func handleError(operation, filePath string, err error) {
	fmt.Printf("Error %s to %s: %v\n", operation, filePath, err)
}

func run() error {
	var fs Filesystem
	switch opts.FilesystemType {
	case "os":
		realFs := afero.NewOsFs()
		aferoFs := afero.Afero{Fs: realFs}
		fs = &RealFilesystem{aferoFs}
	case "mem":
		memFs := afero.NewMemMapFs()
		fs = &RealFilesystem{memFs}
	default:
		fmt.Println("Invalid filesystem type. Supported types: os, mem")
		return fmt.Errorf("invalid filesystem type: %s", opts.FilesystemType)
	}

	// Define a file path
	filePath := "example.txt"

	// Write content to the file
	content := []byte("Hello, Afero with DI and go-flags!")
	err := fs.WriteFile(filePath, content, 0o644)
	if err != nil {
		handleError("writing", filePath, err)
		return fmt.Errorf("write file failed: %w", err)
	}

	// Read content from the file
	readContent, err := fs.ReadFile(filePath)
	if err != nil {
		handleError("reading", filePath, err)
		return fmt.Errorf("read file failed: %w", err)
	}

	fmt.Println(string(readContent))

	return nil
}
