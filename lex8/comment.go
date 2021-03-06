package lex8

import (
	"io"
)

// LexComment lexes a c style comment. It is not a complete LexFunc,
// where it assumes that there is already a "/" buffered in the lexer as a
// precondition.
func LexComment(x *Lexer) *Token {
	if x.Buffered() != "/" {
		panic("needs to buffer a '/' for lex comment")
	}

	if x.Rune() == '/' {
		return lexLineComment(x)
	} else if x.Rune() == '*' {
		return lexBlockComment(x)
	}
	x.Errorf("illegal char %q", x.Rune())
	return x.MakeToken(Illegal)
}

// lexComment is a LexFunc that parses only comments.
func lexComment(x *Lexer) *Token {
	r := x.Rune()
	if r == '/' {
		x.Next()
		return LexComment(x)
	}
	x.Next()
	x.Errorf("illegal rune %q", r)
	return x.MakeToken(Illegal)
}

// NewCommentLexer returns a lexer that parse only comments.
func NewCommentLexer(file string, r io.Reader) *Lexer {
	ret := MakeLexer(file, r, lexComment)
	ret.IsWhite = IsWhiteOrEndl
	return ret
}

func lexLineComment(x *Lexer) *Token {
	for {
		x.Next()
		if x.Ended() || x.Rune() == '\n' {
			break
		}
	}
	return x.MakeToken(Comment)
}

func lexBlockComment(x *Lexer) *Token {
	star := false
	for {
		x.Next()
		if x.Ended() {
			x.Errorf("unexpected eof in block comment")
			return x.MakeToken(Comment)
		}

		if star && x.Rune() == '/' {
			x.Next()
			break
		}

		star = x.Rune() == '*'
	}

	return x.MakeToken(Comment)
}
