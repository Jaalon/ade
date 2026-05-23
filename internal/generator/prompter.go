package generator

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Prompter interface {
	Confirm(message string) (bool, error)
}

type StdPrompter struct{}

func (p *StdPrompter) Confirm(message string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s [y/N] ", message)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	line = strings.TrimSpace(line)
	line = strings.ToLower(line)
	return line == "y" || line == "yes", nil
}
