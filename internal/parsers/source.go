package parsers

import (
	"regexp"
	"strings"

	"go.lsp.dev/protocol"
)

// State is the state of the parser given a position in a source string.
// int
type State int

const (
	StateUnknown State = iota
	StateInComment
)

var (
	embedRegex = regexp.MustCompile(`(?m)^\s*//\s*go:embed\s+(.+)\s*$|/\*\s*go:embed\s+(.+?)\s*\*/`)
)

// ParseSourcePosition parses a source position from a string.
func ParseSourcePosition(
	source *string,
	position protocol.Position,
) (string, State, error) {
	if source == nil {
		return "", StateUnknown, nil
	}
	// split the source string into lines
	lines := strings.Split(*source, "\n")
	line := lines[position.Line-1]
	if len(line) == 0 {
		return "", StateUnknown, nil
	}
	// check if the line is a comment
	if strings.HasPrefix(line, "//") {
		filepath := embedRegex.FindStringSubmatch(line)
		if len(filepath) > 1 && filepath[1] != "" {
			return filepath[1], StateInComment, nil
		}
		return "", StateInComment, nil
	}
	// check if the line is a comment with a star
	if strings.HasPrefix(line, "/*") {
		filepath := embedRegex.FindStringSubmatch(line)
		if len(filepath) > 2 && filepath[2] != "" {
			return filepath[2], StateInComment, nil
		}
		return "", StateInComment, nil
	}
	return "", StateUnknown, nil
}
