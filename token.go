package main

import (
	"fmt"
)

// TokenType is used for representing different tokens
type TokenType int

const (
	tEOF     TokenType = iota // end of file
	tError                    // error has occurred
	tComment                  // comment
	tInt                      // integer
	tFloat                    // floating point
	tIdent                    // identifier
	tString                   // string
	tNewline                  // new line
	tIllegal

	keywordBeginning
	LINE
	SCALE
	MOVE
	ROTATE
	XAXIS
	YAXIS
	ZAXIS
	SAVE
	DISPLAY
	CIRCLE
	HERMITE
	BEZIER
	BOX
	CLEAR
	SPHERE
	TORUS
	PUSH
	POP
	VARY
	BASENAME
	FRAMES
	SET
	SETKNOBS
	keywordEnd
)

var tokens = map[TokenType]string{
	tEOF:     "EOF",
	tError:   "ERROR",
	tComment: "COMMENT",
	tInt:     "INT",
	tFloat:   "FLOAT",
	tIdent:   "IDENTIFIER",
	tString:  "STRING",
	tIllegal: "ILLEGAL",
	tNewline: "NEWLINE",

	LINE:     "line",
	SCALE:    "scale",
	MOVE:     "move",
	ROTATE:   "rotate",
	XAXIS:    "x",
	YAXIS:    "y",
	ZAXIS:    "z",
	SAVE:     "save",
	DISPLAY:  "display",
	CIRCLE:   "circle",
	HERMITE:  "hermite",
	BEZIER:   "bezier",
	BOX:      "box",
	CLEAR:    "clear",
	SPHERE:   "sphere",
	TORUS:    "torus",
	PUSH:     "push",
	POP:      "pop",
	VARY:     "vary",
	BASENAME: "basename",
	FRAMES:   "frames",
	SET:      "set",
	SETKNOBS: "setknobs",
}

var keywords map[string]TokenType

func init() {
	keywords = make(map[string]TokenType)
	for i := keywordBeginning; i < keywordEnd; i++ {
		keywords[tokens[i]] = i
	}
}

// Token is a lexical token
type Token struct {
	tt    TokenType // type of token
	value string    // value of token
}

func (tt TokenType) String() string {
	if s, isToken := tokens[tt]; isToken {
		return s
	}
	return fmt.Sprint("token(", tt, ")")
}

func (t Token) String() string {
	return fmt.Sprintf("{%s %s}", t.tt, t.value)
}

// LookupIdent returns the corresponding token type for an identifier
func LookupIdent(ident string) TokenType {
	if tok, isKeyword := keywords[ident]; isKeyword {
		return tok
	}
	return tIllegal
}
