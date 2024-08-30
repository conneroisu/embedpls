package methods

import (
	"encoding/json"
	"strings"
)

// Method represents an LSP method
type Method string

// String returns the string representation of the method
func (m Method) String() string {
	return string(m)
}

// Prefix returns true if the method starts with the input
func (m Method) Prefix(input string) bool {
	return strings.HasPrefix(string(m), input)
}

// Decode decodes the method to a byte array
func (m Method) Decode() ([]byte, error) {
	return json.Marshal(m)
}
