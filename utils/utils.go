package utils

import (
	user "asa/shell/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrCommandNotFound      = errors.New("command not found")
	ErrEnvironmentVarNotSet = errors.New("PATH environment variable is not set")
	ErrPwdWentWrong         = errors.New("something went wrong while trying to identify current dir")
	ErrInvalidQuotedArg     = errors.New("invalid Quoted arg")
	ErrInvalidValue         = errors.New("corresponding value is not stored correctly")
	ErrUsernameRequired     = errors.New("username required")
	ErrUserAlreadyExist     = errors.New("user already exist")
	ErrLoggedin             = errors.New("a user is currently logged in to shell")
	ErrTooManyArgs          = errors.New("too many arguments")
	ErrInvalidArgs          = errors.New("invalid arguments")
	ErrEmptyHistory         = errors.New("empty command history")
	ErrNotEnoughArgs        = errors.New("not Enough Arguments")
	ErrUnvalidArg           = errors.New("unvalid Argument")
	ErrColorUnset           = errors.New("color is not set")
	ErrColorSet             = errors.New("color is already set")
	ErrMissingCommandName   = errors.New("type: missing command name")
)

const (
	// Text colors
	TextBlack   = "\033[30m"
	TextRed     = "\033[31m"
	TextGreen   = "\033[32m"
	TextYellow  = "\033[33m"
	TextBlue    = "\033[34m"
	TextMagenta = "\033[35m"
	TextCyan    = "\033[36m"
	TextWhite   = "\033[37m"

	// Background colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"

	// Special formatting
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
	Reverse   = "\033[7m"
)

func ColorText(text string, formats ...string) string {
	var combined string
	for _, format := range formats {
		combined += format
	}
	return combined + text + Reset
}

func IsColor() bool {
	_, ok := os.LookupEnv("SHELLCOLOR")
	return ok
}

func FindCommand(cmd string) (string, error) {

	if strings.Contains(cmd, "/") {
		return cmd, nil 
	}

	path := os.Getenv("PATH")
	if path == "" {
		return "", ErrEnvironmentVarNotSet
	}

	dirs := strings.Split(path, ":")
	for _, dir := range dirs {
		fullPath := filepath.Join(dir, cmd)
		if isExecutable(fullPath) {
			return fullPath, nil
		}
	}

	return "", ErrCommandNotFound
}

func IsQuoted(s string) bool {
	return (HasPrefix(s, "'") && HasSuffix(s, "'") && len(s) > 1) ||
		(HasPrefix(s, "\"") && HasSuffix(s, "\"") && len(s) > 1)
}

func WhichQuoted(s string) string {

	switch {
	case HasPrefix(s, "'") && HasSuffix(s, "'") && len(s) > 1:
		return "'"
	case HasPrefix(s, "\"") && HasSuffix(s, "\"") && len(s) > 1:
		return "\""
	default:
		return ""
	}

}

func IsAlphaNumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9')
}

func HasPrefix(s string, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	w := !(s[0:len(prefix)] == prefix)
	return len(s) >= len(prefix) && !w
}

func HasSuffix(s string, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	w := !(s[len(s)-len(suffix):] == suffix)
	return len(s) >= len(suffix) && !w
}

func TrimEdge(s string) string {
	if len(s) < 1 {
		return s
	}
	if len(s) == 1 {
		return ""
	}

	x := []rune(s)
	x = x[1 : len(x)-1]
	return string(x)
}

// func ExtractQuotes(input string) ([]string, error) {
// 	res := []string{}
// 	for i := 0; i < len(input); i++ {
// 		if input[i] == '"' {
// 			var result strings.Builder
// 			result.WriteByte(input[i])
// 			for i = i + 1; i < len(input) && input[i] != '"'; i++ {
// 				result.WriteByte(input[i])
// 			}
// 			if i == len(input) || input[i] != '"' {
// 				result.WriteByte('"') // last quote
// 				res = append(res, result.String())
// 				return res, ErrInvalidQuotedArg
// 			}
// 			result.WriteByte(input[i]) // last quote
// 			res = append(res, result.String())
// 		} else if input[i] == '\'' {
// 			var result strings.Builder
// 			result.WriteByte(input[i])
// 			for i = i + 1; i < len(input) && input[i] != '\''; i++ {
// 				result.WriteByte(input[i])
// 			}
// 			if i == len(input) || input[i] != '\'' {
// 				result.WriteByte('\'') // last quote
// 				res = append(res, result.String())
// 				return res, ErrInvalidQuotedArg
// 			}
// 			result.WriteByte(input[i]) // last quote
// 			res = append(res, result.String())
// 		}
// 	}
// 	return res, nil
// }

// func Seperate(input string, quotes []string) (res []string) {
// 	fields := strings.Fields(input)
// 	j := 0
// 	for i := 0; i < len(fields); i++ {
// 		if HasPrefix(fields[i], string("'")) {
// 			if IsQuoted(fields[i]) && len(fields[i]) != 1 {
// 				if j >= len(quotes) {
// 					return res
// 				}
// 				res = append(res, quotes[j])
// 				j++
// 				continue
// 			}
// 			i++
// 			for i < len(fields) && !HasSuffix(fields[i], string("'")) {
// 				i++
// 			}
// 			if j >= len(quotes) {
// 				return res
// 			}
// 			res = append(res, quotes[j])
// 			j++
// 			continue
// 		} else if HasPrefix(fields[i], string('"')) {
// 			if IsQuoted(fields[i]) && len(fields[i]) != 1 {
// 				if j >= len(quotes) {
// 					return res
// 				}
// 				res = append(res, quotes[j])
// 				j++
// 				continue
// 			}
// 			i++
// 			for i < len(fields) && !HasSuffix(fields[i], string('"')) {
// 				i++
// 			}
// 			if j >= len(quotes) {
// 				return res
// 			}
// 			res = append(res, quotes[j])
// 			j++
// 			continue
// 		}
// 		res = append(res, fields[i])
// 	}
// 	return res
// }

func HandleAdress(baseAddr string, currentDir string) string {
	cleanBaseAddr := strings.Trim(baseAddr, "\n")
	cleanCurrentDir := strings.Trim(currentDir, "\n")

	if HasPrefix(cleanCurrentDir, cleanBaseAddr) {
		if cleanBaseAddr == cleanCurrentDir {
			return "~"
		}
		return "~" + cleanCurrentDir[len(cleanBaseAddr):]
	}
	return cleanCurrentDir
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && (info.Mode()&0111 != 0)
}

func IsValidDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func CurrentPwd() (string, error) {
	if pwd, err := filepath.Abs("."); pwd != "" && err == nil {
		if IsValidDirectory(pwd) {
			return filepath.Clean(pwd), nil
		}
	}
	return "", ErrPwdWentWrong
}

func ParseArgs(input string) ([]string, error) {
	var err error
	escape := false
	space := false
	str := ""
	i := 0
	res := []string{}
	for i < len(input) {
		if input[i] == ' ' {
			escape = false
			if !space {
				res = append(res, str)
				str = ""
			}
			space = true
		} else if escape {
			space = false
			str += string(input[i])
			escape = false
		} else if !escape {
			if input[i] == '`' {
				return res, ErrInvalidQuotedArg

			} else if input[i] == '$' {
				space = false
				str, i = handleEnv(input, i, str)
			} else if input[i] == '\\' {
				escape = true
			} else if input[i] == '\'' {
				space = false
				str, i, err = handleSinglgQ(input, i, str)
				if err != nil {
					return res, err
				}
			} else if input[i] == '"' {
				space = false
				str, i, err = handleDoubleQ(input, i, str)
				if err != nil {
					return res, err
				}
			} else {
				space = false
				str += string(input[i])
			}
		} else {
			return res, fmt.Errorf("unhandled parsing situation %s", string(input[i]))
		}
		i++
	}
	if escape {
		return res, ErrInvalidQuotedArg
	}
	if str != "" {
		res = append(res, str)
	}
	return res, nil
}

func handleSinglgQ(input string, idx int, result string) (string, int, error) {
	if idx == len(input)-1 {
		return "", 0, ErrInvalidQuotedArg
	}
	for j := idx + 1; j < len(input); j++ {
		if input[j] == '\'' {
			return result, j, nil
		}
		result += string(input[j])
	}
	return "", 0, ErrInvalidQuotedArg
}

func handleDoubleQ(input string, idx int, result string) (string, int, error) {
	escape := false
	str := result
	if idx == len(input)-1 {
		return "", 0, ErrInvalidQuotedArg
	}
	i := idx + 1
	for i < len(input) {
		if escape {
			if input[i] == '"' || input[i] == '`' || input[i] == '$' || input[i] == '\\' {
				str = str[:len(str)-1]
			}
			str += string(input[i])
			escape = false
		} else if !escape {
			if input[i] == '`' {
				return str, 0, ErrInvalidQuotedArg

			} else if input[i] == '$' {
				str, i = handleEnv(input, i, str)
			} else if input[i] == '\\' {
				escape = true
				str += string(input[i])
			} else if input[i] == '"' {
				return str, i, nil
			} else {
				str += string(input[i])
			}
		} else {
			return "", 0, fmt.Errorf("unhandled parsing situation %s", string(input[i]))
		}
		i++
	}
	return "", 0, ErrInvalidQuotedArg
}

func handleEnv(input string, idx int, result string) (string, int) {
	res := result
	if idx == len(input)-1 {
		return "", idx
	}
	i := idx + 1
	for i < len(input) && (IsAlphaNumeric(input[i]) || input[i] == '_') {
		i++
	}
	if i > idx+1 {
		varName := input[idx+1 : i]
		varValue := os.Getenv(varName)
		res += varValue
	}
	return res, i - 1
}

func MockHistoryData() map[string]int {
	return map[string]int{
		"ls":         2,
		"ls -l":      3,
		"echo hello": 1,
		"pwd":        1,
		"cd ..":      2,
	}
}
func ClearAndFillHistoryWithMockData(db *gorm.DB) error {
	var users []user.User

	if err := db.Find(&users).Error; err != nil {
		return fmt.Errorf("failed to retrieve users: %w", err)
	}

	for _, obj := range users {
		fmt.Printf("Processing user: %s (ID: %d)\n", obj.Username, obj.ID)

		historyMap := MockHistoryData()

		historyJSON, err := json.Marshal(historyMap)
		if err != nil {
			return fmt.Errorf("failed to marshal history to JSON for user %s: %w", obj.Username, err)
		}
		fmt.Printf("History JSON to be saved for user %s: %s\n", obj.Username, string(historyJSON)) 

		obj.History = string(historyJSON)
		if err := user.Update(db, &obj); err != nil {
			return fmt.Errorf("failed to update history for user %s: %w", obj.Username, err)
		}
		fmt.Printf("History updated for user: %s\n", obj.Username)
	}

	fmt.Println("Successfully cleared and filled history for all users.")
	return nil
}

func PrintSortedMap(historyMap map[string]int, stdout io.Writer) {
	var sortedPairs []struct {
		Key   string
		Value int
	}
	for key, val := range historyMap {
		sortedPairs = append(sortedPairs, struct {
			Key   string
			Value int
		}{Key: key, Value: val})
	}

	sort.Slice(sortedPairs, func(i, j int) bool {
		return sortedPairs[j].Value < sortedPairs[i].Value
	})

	fmt.Fprintln(stdout, "------------------------------")
	fmt.Fprintln(stdout, "|      Command       | Count |")
	fmt.Fprintln(stdout, "------------------------------")
	for _, pair := range sortedPairs {
		fmt.Fprintf(stdout, "| %-18s | %-5d |\n", pair.Key, pair.Value)
	}
	fmt.Fprintln(stdout, "------------------------------")

}
