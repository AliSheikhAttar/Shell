package ls

import (
	"asa/shell/utils"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"
)

// LSCommand implements the 'ls' built-in command
type LSCommand struct {
	showAll    bool
	longFormat bool
}

// NewLSCommand creates a new ls command
func NewLSCommand() *LSCommand {
	return &LSCommand{}
}

// Name returns the name of the command
func (c *LSCommand) Name() string {
	return "ls"
}

// Execute handles the ls command execution
func (c *LSCommand) Execute(args []string, stdout io.Writer) error {
	// Parse arguments and options
	dirPath, err := c.parseArgs(args)
	if err != nil {
		return err
	}

	// If no directory specified, use current directory
	if dirPath == "" {
		dirPath = "."
	}

	return c.listDirectory(dirPath, stdout)
}

// parseArgs processes command line arguments and returns the target directory
func (c *LSCommand) parseArgs(args []string) (string, error) {
	var dirPath string
	c.showAll = false
	c.longFormat = false
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			// Process options
			for _, opt := range arg[1:] {
				switch opt {
				case 'a':
					c.showAll = true
				case 'l':
					c.longFormat = true
				default:
					return "", fmt.Errorf("invalid option: %c", opt)
				}
			}
		} else {
			// Process directory path
			if dirPath != "" {
				return "", utils.ErrTooManyArgs
			}
			dirPath = arg
		}
	}

	return dirPath, nil
}

// listDirectory lists the contents of the specified directory
func (c *LSCommand) listDirectory(dirPath string, stdout io.Writer) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("cannot open directory %s: %v", dirPath, err)
	}

	// Sort entries (optional, as ReadDir returns sorted entries by default)
	for _, entry := range entries {
		// Skip hidden files unless -a flag is set
		if !c.showAll && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if c.longFormat {
			if err := c.printLongFormat(entry, stdout); err != nil {
				return err
			}
		} else {
			fmt.Fprintln(stdout, entry.Name())
		}
	}

	return nil
}

// printLongFormat prints detailed file information
func (c *LSCommand) printLongFormat(entry fs.DirEntry, stdout io.Writer) error {
	info, err := entry.Info()
	if err != nil {
		return fmt.Errorf("error getting info for %s: %v", entry.Name(), err)
	}

	// Format: permissions size modified_time name
	mode := info.Mode().String()
	size := info.Size()
	modTime := info.ModTime().Format(time.RFC3339[:19]) // Use shorter time format
	name := entry.Name()
	fmt.Fprintf(stdout, "%s %8d %s %s\n", mode, size, modTime, name)
	return nil
}
