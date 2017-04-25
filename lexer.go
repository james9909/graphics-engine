package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type stateFn func(*Lexer) stateFn

type Lexer struct {
	in    *bufio.Reader
	out   chan Token
	buf   []rune
	state stateFn
	line  int
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

func NewLexer() *Lexer {
	return &Lexer{
		out: make(chan Token),
		buf: make([]rune, 0, 10),
	}
}

func (l *Lexer) Lex(r io.Reader) {
	l.in = bufio.NewReader(r)
	go l.run()
}

func (l *Lexer) accept(s string) bool {
	r := l.next()
	if strings.IndexRune(s, r) >= 0 {
		l.keep(r)
		return true
	}
	l.unread()
	return false
}

func (l *Lexer) acceptAll(s string) {
	for l.accept(s) {
	}
}

func (l *Lexer) emit(tt TokenType) {
	l.out <- Token{
		tt:    tt,
		value: string(l.buf),
	}
	l.buf = l.buf[0:0]
}

func (l *Lexer) next() rune {
	r, _, err := l.in.ReadRune()
	if err != nil {
		return eof
	}
	return r
}

func (l *Lexer) unread() {
	l.in.UnreadRune()
}

func (l *Lexer) peek() rune {
	r := l.next()
	l.unread()
	return r
}

func (l *Lexer) keep(r rune) {
	if l.buf == nil {
		l.buf = make([]rune, 0, 10)
	}
	l.buf = append(l.buf, r)
}

func (l *Lexer) run() {
	defer close(l.out)
	for l.state = lexRoot; l.state != nil; {
		l.state = l.state(l)
	}
}

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

func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.out <- Token{
		tt:    tError,
		value: fmt.Sprintf(format, args),
	}
	return nil
}

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

func lexNumber(l *Lexer) stateFn {
	l.accept("+-")

	l.acceptAll("0123456789")
	if l.accept(".") {
		l.acceptAll("0123456789")
	}
	next := l.peek()
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

func isIdent(s string) bool {
	for i := 0; i < len(commands); i++ {
		if commands[i] == s {
			return true
		}
	}
	return false
}
