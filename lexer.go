package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// stateFn is a state function executes an action and returns the next state
type stateFn func(*Lexer) stateFn

// Lexer is a struct that will lex a script for tokens
type Lexer struct {
	input  string     // input string
	length int        // length of input string
	tokens chan Token // channel of tokens
	state  stateFn    // current state function
	pos    int        // lexer's current position in the input
	start  int        // starting position of the current item
	line   int        // current line
	width  int        // width of the last rune
}

var eof = rune(0)

// NewLexer returns a new lexer
func NewLexer() *Lexer {
	return &Lexer{
		tokens: make(chan Token),
		state:  lexRoot,
	}
}

// Lex lexes an io.Reader for tokens
func (l *Lexer) Lex(input string) {
	l.input = input
	l.length = len(input)
	go l.run()
}

// accept consumes a rune if it is in the valid charset
func (l *Lexer) accept(s string) bool {
	r := l.next()
	if strings.IndexRune(s, r) >= 0 {
		return true
	}
	l.unread()
	return false
}

// acceptRun consumes all consecutive runes in a valid charset
func (l *Lexer) acceptRun(s string) {
	for l.accept(s) {
	}
}

// emit passes the current token into the token channel
func (l *Lexer) emit(tt TokenType) {
	l.tokens <- Token{
		tt:    tt,
		value: l.input[l.start:l.pos],
	}
	l.start = l.pos
}

// ignore passes over the current token
func (l *Lexer) ignore() {
	l.start = l.pos
}

// next consumes and returns the next rune
func (l *Lexer) next() rune {
	if l.pos >= l.length {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += w
	if r == '\n' {
		l.line++
	}
	return r
}

// unread steps back one rune
func (l *Lexer) unread() {
	l.pos -= l.width
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// peek returns the next rune without consuming it
func (l *Lexer) peek() rune {
	r := l.next()
	l.unread()
	return r
}

// run lexes the input and executes all state functions
func (l *Lexer) run() {
	defer close(l.tokens)
	for l.state != nil {
		l.state = l.state(l)
	}
}

// lexRoot is the main state function
func lexRoot(l *Lexer) stateFn {
	r := l.next()
	switch {
	case r == eof:
		l.emit(tEOF)
		return nil
	case r == '#':
		return lexComment
	case r == '\n' || r == '\r':
		l.ignore()
		return lexRoot
	case r == ' ' || r == '\t':
		l.ignore()
		return lexRoot
	case strings.IndexRune(".+-0123456789", r) >= 0:
		l.unread()
		return lexNumber
	case unicode.IsPrint(r):
		return lexString
	default:
		return l.error(fmt.Sprintf("unexpected rune '%c'", r))
	}
}

// error emits a lex error
func (l *Lexer) error(s string) stateFn {
	l.tokens <- Token{
		tt:    tError,
		value: fmt.Sprintf("%d: syntax error: %s", l.line, s),
	}
	return nil
}

// lexComment lexes a comment
func lexComment(l *Lexer) stateFn {
	r := l.next()
	switch r {
	case '\n':
		l.ignore()
		return lexRoot
	case eof:
		l.ignore()
		return nil
	default:
		return lexComment
	}
}

// lexNumber lexes a number
func lexNumber(l *Lexer) stateFn {
	// accept an optional sign
	l.accept("+-")

	l.acceptRun("0123456789")
	// accept floating points
	if l.accept(".") {
		l.acceptRun("0123456789")
	}
	next := l.peek()
	// The next character must be numeric
	if unicode.IsLetter(next) {
		return l.error("invalid number")
	}
	if strings.ContainsRune(l.input[l.start:l.pos], '.') {
		l.emit(tFloat)
	} else {
		l.emit(tInt)
	}
	return lexRoot
}

// lexString lexes a string
func lexString(l *Lexer) stateFn {
	r := l.next()
	for unicode.IsPrint(r) && !unicode.IsSpace(r) {
		r = l.next()
	}
	l.unread()
	if Lookup(l.input[l.start:l.pos]) == tIllegal {
		// Not a legal identifier, so treat it as a string
		l.emit(tString)
	} else {
		l.emit(tIdent)
	}
	return lexRoot
}
