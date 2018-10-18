package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	TokenEOF   TokenType = -1
	TokenError TokenType = -2

	EOF = -1
)

type TokenType int

type Token struct {
	Type  TokenType
	Value string
}

func (t Token) String() string {
	switch t.Type {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return "Error: " + t.Value
	}
	if len(t.Value) > 50 {
		return fmt.Sprintf("%d:%.50q...", t.Type, t.Value)
	}
	return fmt.Sprintf("%d:%q", t.Type, t.Value)
}

type StateFunc func(*Lexer) StateFunc

// Lexer
type Lexer struct {
	input  string
	start  int
	pos    int
	width  int
	state  StateFunc
	tokens chan Token
}

// No copying, just a slice
// Channel usage adds some overhead could use ring buffer
func NewLexer(input string, initialState StateFunc) *Lexer {
	l := &Lexer{
		input:  input,
		tokens: make(chan Token, 2),
		state:  initialState,
	}
	return l
}

func (t *Lexer) NextToken() Token {
	for {
		select {
		case item := <-t.tokens:
			return item
		default:
			t.state = t.state(t)
		}
	}
	panic("should never get here")
}

func (t *Lexer) Emit(i TokenType) {
	if t.pos > len(t.input) {
		t.tokens <- Token{TokenError, "Reached end of input unexpectantly"}
		return
	}

	fmt.Printf("E '%s'\n", t.input[t.start:t.pos])

	t.tokens <- Token{i, t.input[t.start:t.pos]}
	t.start = t.pos
}

func (l *Lexer) Peek() rune {
	r := l.Next()
	l.Backup()
	return r
}

func (t *Lexer) Next() rune {
	r, w := utf8.DecodeRuneInString(t.input[t.pos:])
	t.width = w
	t.pos += t.width

	if int(t.pos) >= len(t.input) {
		t.width = 0
		return EOF
	}

	fmt.Println("R", string(r))
	return r
}

func (t *Lexer) Skip() {
	t.Next()
	t.Ignore()
}

// ignore skips over the pending input before this point.
func (l *Lexer) Ignore() {
	l.start = l.pos
}

// backup steps back one rune.
// Can be called only once per call of next.
func (l *Lexer) Backup() {
	l.pos -= l.width
}

// accept consumes the next rune
// if it's from the valid set.
func (l *Lexer) Accept(valid string) bool {
	if strings.IndexRune(valid, l.Next()) >= 0 {
		return true
	}
	l.Backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *Lexer) AcceptRun(valid string) {
	for strings.IndexRune(valid, l.Next()) >= 0 {
	}
	l.Backup()
}

func (l *Lexer) Matches(str string) bool {
	if strings.HasPrefix(l.input[l.pos:], str) {
		l.pos += len(str)
	}
	return false
}

func (t *Lexer) Errorf(format string, args ...interface{}) StateFunc {
	t.tokens <- Token{TokenError, fmt.Sprintf(format, args...)}
	return nil
}

func IsAlphaNumeric(r rune) bool {
	return 'A' <= r && r <= 'Z' || 'a' <= r && r <= 'z' || '0' <= r && r <= '9'
}
