package main

import "fmt"

type TokenType int

const (
	tEOF TokenType = iota
	tError
	tComment
	tInt
	tFloat
	tIdent
	tString
)

type Token struct {
	tt    TokenType
	value string
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
