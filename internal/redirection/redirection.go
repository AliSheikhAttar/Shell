// internal/redirection/redirection.go
package redirection

import (
	"fmt"
	"os"
)

type RedirectionType int

var (
	ErrMissingFileForRedirection = fmt.Errorf("missing file for redirection")
)

const (
	NoRedirection  RedirectionType = iota
	OutputRedirect                 // >
	OutputAppend                   // >>
	ErrorRedirect                  // 2>
	ErrorAppend                    // 2>>
)

type Redirection struct {
	Type RedirectionType
	File string
}

// ParseRedirection parses command arguments for redirection operators
func ParseRedirection(args []string) ([]string, *Redirection, error) {
	if len(args) == 0 {
		return args, nil, nil
	}

	for i, arg := range args {
		switch {
		case arg == ">":
			if i+1 >= len(args) {
				return nil, nil, ErrMissingFileForRedirection
			}
			return args[:i], &Redirection{
				Type: OutputRedirect,
				File: args[i+1],
			}, nil

		case arg == ">>":
			if i+1 >= len(args) {
				return nil, nil, ErrMissingFileForRedirection
			}
			return args[:i], &Redirection{
				Type: OutputAppend,
				File: args[i+1],
			}, nil

		case arg == "2>":
			if i+1 >= len(args) {
				return nil, nil, ErrMissingFileForRedirection
			}
			return args[:i], &Redirection{
				Type: ErrorRedirect,
				File: args[i+1],
			}, nil

		case arg == "2>>":
			if i+1 >= len(args) {
				return nil, nil, ErrMissingFileForRedirection
			}
			return args[:i], &Redirection{
				Type: ErrorAppend,
				File: args[i+1],
			}, nil
		}
	}

	return args, nil, nil
}

// SetupRedirection configures the redirection for a command
func SetupRedirection(redir *Redirection) (*os.File, error) {
	if redir == nil {
		return nil, nil
	}

	flags := os.O_WRONLY | os.O_CREATE
	if redir.Type == OutputAppend || redir.Type == ErrorAppend {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	file, err := os.OpenFile(redir.File, flags, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open redirection file: %v", err)
	}
	return file, nil
}
