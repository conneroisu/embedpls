package parsers

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
	"go.lsp.dev/protocol"
)

// State is the state of the parser given a position in a source string.
// int
type State int

const (
	// StateUnknown is the state for an unknown position.
	StateUnknown State = iota
	// StateInComment is the state for a position in a comment.
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
	line := lines[position.Line]
	log.Debugf("current line: %s", line)
	if len(line) == 0 {
		return "", StateUnknown, nil
	}
	if strings.HasPrefix(line, "//g") {
		log.Debugf("found //go")
		return "", StateInComment, nil
	}
	return "", StateUnknown, nil
}
