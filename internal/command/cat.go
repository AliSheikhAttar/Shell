package command

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

type CatCommand struct{
    stdout io.Writer
}

func NewCatCommand(stdout io.Writer) *CatCommand {
    return &CatCommand{
        stdout: stdout,
    }
}

func (c *CatCommand) Name() string {
    return "cat"
}

func (c *CatCommand) Execute(args []string) error {
    if len(args) == 0 {
        return ErrCatNoArgs
    }

    for _, filename := range args {
        if err := c.displayFile(filename); err != nil {
            return fmt.Errorf("cat: %s -> %v", filename, err)
        }
    }

    return nil
}

func (c *CatCommand) displayFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        fmt.Fprintln(c.stdout, scanner.Text())
    }

    return scanner.Err()
}