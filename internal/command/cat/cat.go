package cat

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	ErrCatNoArgs = errors.New("no argument")
)

type CatCommand struct{}

func NewCatCommand() *CatCommand {
	return &CatCommand{}
}

func (c *CatCommand) Name() string {
	return "cat"
}

func (c *CatCommand) Execute(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return ErrCatNoArgs
	}

	for _, filename := range args {
		if err := c.displayFile(filename, stdout); err != nil {
			return fmt.Errorf("cat: %s -> %v", filename, err)
		}
	}

	return nil
}

func (c *CatCommand) displayFile(filename string, stdout io.Writer) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Fprintln(stdout, scanner.Text())
	}

	return scanner.Err()
}
