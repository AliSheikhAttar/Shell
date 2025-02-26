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

type LSCommand struct {
	showAll    bool
	longFormat bool
}

func NewLSCommand() *LSCommand {
	return &LSCommand{}
}

func (c *LSCommand) Name() string {
	return "ls"
}

func (c *LSCommand) Execute(args []string, stdout io.Writer) error {
	dirPath, err := c.parseArgs(args)
	if err != nil {
		return err
	}

	if dirPath == "" {
		dirPath = "."
	}

	return c.listDirectory(dirPath, stdout)
}

func (c *LSCommand) parseArgs(args []string) (string, error) {
	var dirPath string
	c.showAll = false
	c.longFormat = false
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
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
			if dirPath != "" {
				return "", utils.ErrTooManyArgs
			}
			dirPath = arg
		}
	}

	return dirPath, nil
}

func (c *LSCommand) listDirectory(dirPath string, stdout io.Writer) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("cannot open directory %s: %v", dirPath, err)
	}

	for _, entry := range entries {
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

func (c *LSCommand) printLongFormat(entry fs.DirEntry, stdout io.Writer) error {
	info, err := entry.Info()
	if err != nil {
		return fmt.Errorf("error getting info for %s: %v", entry.Name(), err)
	}

	mode := info.Mode().String()
	size := info.Size()
	modTime := info.ModTime().Format(time.RFC3339[:19]) // Use shorter time format
	name := entry.Name()
	fmt.Fprintf(stdout, "%s %8d %s %s\n", mode, size, modTime, name)
	return nil
}
