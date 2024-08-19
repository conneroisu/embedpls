package parsers

import (
	"testing"

	"go.lsp.dev/protocol"
)

// TestParseSourcePosition tests the ParseSourcePosition function.
func TestParseSourcePosition(t *testing.T) {
	tests := []struct {
		name      string
		source    *string
		position  protocol.Position
		wantStr   string
		wantState State
		wantErr   bool
	}{
		{
			name:      "nil source",
			source:    nil,
			position:  protocol.Position{Line: 1, Character: 0},
			wantStr:   "",
			wantState: StateUnknown,
			wantErr:   false,
		},
		{
			name:      "empty line",
			source:    ptrToStr(""),
			position:  protocol.Position{Line: 1, Character: 0},
			wantStr:   "",
			wantState: StateUnknown,
			wantErr:   false,
		},
		{
			name:      "line is a comment with go:embed directive",
			source:    ptrToStr("// go:embed file.txt"),
			position:  protocol.Position{Line: 1, Character: 0},
			wantStr:   "file.txt",
			wantState: StateInComment,
			wantErr:   false,
		},
		{
			name:      "line is a comment without go:embed directive",
			source:    ptrToStr("// This is a comment"),
			position:  protocol.Position{Line: 1, Character: 0},
			wantStr:   "",
			wantState: StateInComment,
			wantErr:   false,
		},
		{
			name:      "line is a comment with go:embed in block comment",
			source:    ptrToStr("/* go:embed file.txt */"),
			position:  protocol.Position{Line: 1, Character: 0},
			wantStr:   "file.txt",
			wantState: StateInComment,
			wantErr:   false,
		},
		{
			name:      "line is a comment block without go:embed directive",
			source:    ptrToStr("/* This is a comment */"),
			position:  protocol.Position{Line: 1, Character: 0},
			wantStr:   "",
			wantState: StateInComment,
			wantErr:   false,
		},
		{
			name:      "line is code, not a comment",
			source:    ptrToStr("fmt.Println(\"Hello, world!\")"),
			position:  protocol.Position{Line: 1, Character: 0},
			wantStr:   "",
			wantState: StateUnknown,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotState, err := ParseSourcePosition(tt.source, tt.position)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSourcePosition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotStr != tt.wantStr {
				t.Errorf("ParseSourcePosition() gotStr = %v, want %v", gotStr, tt.wantStr)
			}
			if gotState != tt.wantState {
				t.Errorf("ParseSourcePosition() gotState = %v, want %v", gotState, tt.wantState)
			}
		})
	}
}

func ptrToStr(s string) *string {
	return &s
}
