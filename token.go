package main

import "fmt"

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
)

// Token is a lexical token
type Token struct {
	tt    TokenType // type of token
	value string    // value of token
}

func (tt TokenType) String() string {
	switch tt {
	case tEOF:
		return "EOF"
	case tError:
		return "ERROR"
	case tComment:
		return "COMMENT"
	case tInt:
		return "INT"
	case tFloat:
		return "FLOAT"
	case tIdent:
		return "IDENT"
	case tString:
		return "STRING"
	default:
		panic(fmt.Errorf("invalid token type"))
	}
}

func (t Token) String() string {
	return fmt.Sprintf("{%s %s}", t.tt, t.value)
}
