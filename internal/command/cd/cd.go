package cd

import (
	"asa/shell/utils"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNoFileDir = errors.New("no such file or directory")
)

type CDCommand struct {
	rootDir string
}

func NewCDCommand(rootDir string) *CDCommand {
	return &CDCommand{
		rootDir: rootDir,
	}
}

func (c *CDCommand) Name() string {
	return "cd"
}

func (c *CDCommand) Execute(args []string, stdout io.Writer) error {
	var dir string

	switch len(args) {
	case 0:
		dir = c.rootDir
	case 1:
		switch args[0] {
		case "~":
			dir = os.Getenv("HOME")
		default:
			dir = args[0]
			if len(dir) > 0 && dir[0] == '~' {
				home := os.Getenv("HOME")
				dir = filepath.Join(home, dir[1:])
				break
			}
			if dir == ".." {
				currentDir, err := utils.CurrentPwd()
				if err != nil {
					return err
				}
				currentCleanDir := filepath.Clean(currentDir)
				dirArgs := strings.Split(currentCleanDir, "/")
				dirArgs = dirArgs[:len(dirArgs)-1]
				newDir := strings.Join(dirArgs, "/")
				if utils.IsValidDirectory(newDir) {
					dir = newDir
					break

				}
			}
		}
	default:
		return utils.ErrTooManyArgs
	}

	if err := os.Chdir(dir); err != nil {
		return ErrNoFileDir
	}

	return nil
}
