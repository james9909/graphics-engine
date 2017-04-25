package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// stateFn is a state function executes an action and returns the next state
type stateFn func(*Lexer) stateFn

// Lexer is a struct that will lex a script for tokens
type Lexer struct {
	in    *bufio.Reader // input stream
	out   chan Token    // channel of tokens
	buf   []rune        // value of the current token
	state stateFn       // current state function
	line  int           // current line
}

var eof = rune(0)
var commands = []string{
	"push",
	"pop",
	"move",
	"translate",
	"rotate",
	"scale",
	"box",
	"sphere",
	"torus",
	"line",
	"curve",
	"circle",
	"save",
	"display",
	"ident",
	"clear",
}

// NewLexer returns a new lexer
func NewLexer() *Lexer {
	return &Lexer{
		out:   make(chan Token),
		buf:   make([]rune, 0, 10),
		state: lexRoot,
	}
}

// Lex lexes an io.Reader for tokens
func (l *Lexer) Lex(r io.Reader) {
	l.in = bufio.NewReader(r)
	go l.run()
}

// accept consumes a rune if it is in the valid charset
func (l *Lexer) accept(s string) bool {
	r := l.next()
	if strings.IndexRune(s, r) >= 0 {
		l.keep(r)
		return true
	}
	l.unread()
	return false
}

// acceptAll consumes all consecutive runes in a valid charset
func (l *Lexer) acceptAll(s string) {
	for l.accept(s) {
	}
}

// emit passes the current token into the token channel
func (l *Lexer) emit(tt TokenType) {
	l.out <- Token{
		tt:    tt,
		value: string(l.buf),
	}
	l.buf = l.buf[0:0]
}

// next consumes and returns the next rune
func (l *Lexer) next() rune {
	r, _, err := l.in.ReadRune()
	if err != nil {
		return eof
	}
	return r
}

// unread steps back one rune
func (l *Lexer) unread() {
	l.in.UnreadRune()
}

// peek returns the next rune without consuming it
func (l *Lexer) peek() rune {
	r := l.next()
	l.unread()
	return r
}

// keep adds the given rune into the buffer
func (l *Lexer) keep(r rune) {
	if l.buf == nil {
		l.buf = make([]rune, 0, 10)
	}
	l.buf = append(l.buf, r)
}

// run lexes the input and executes all state functions
func (l *Lexer) run() {
	defer close(l.out)
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
		l.line++
		return lexRoot
	case r == ' ' || r == '\t':
		return lexRoot
	case strings.IndexRune("+-0123456789", r) >= 0:
		l.unread()
		return lexNumber
	case unicode.IsPrint(r):
		l.keep(r)
		return lexString
	default:
		return l.errorf("line %d: unexpected rune '%c'", l.line, r)
	}
}

// errorf emits a lex error
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.out <- Token{
		tt:    tError,
		value: fmt.Sprintf(format, args),
	}
	return nil
}

// lexComment lexes a comment
func lexComment(l *Lexer) stateFn {
	r := l.next()
	switch r {
	case '\n':
		l.emit(tComment)
		return lexRoot
	case eof:
		l.emit(tComment)
		return nil
	default:
		l.keep(r)
		return lexComment
	}
}

// lexNumber lexes a number
func lexNumber(l *Lexer) stateFn {
	// accept an optional sign
	l.accept("+-")

	l.acceptAll("0123456789")
	// accept floating points
	if l.accept(".") {
		l.acceptAll("0123456789")
	}
	next := l.peek()
	// if the next character is not numeric, then treat it as a string
	if unicode.IsLetter(next) {
		return lexString
	}
	if strings.ContainsRune(string(l.buf), '.') {
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
		l.keep(r)
		r = l.next()
	}
	l.unread()
	if isIdent(string(l.buf)) {
		l.emit(tIdent)
	} else {
		l.emit(tString)
	}
	return lexRoot
}

// isIdent returns true if a string is an identifier, false otherwise
func isIdent(s string) bool {
	for i := 0; i < len(commands); i++ {
		if commands[i] == s {
			return true
		}
	}
	return false
}
