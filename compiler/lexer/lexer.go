// Package lexer tokenises KScript source code.
package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType identifies the kind of a lexed token.
type TokenType string

const (
	TokEOF     TokenType = "EOF"
	TokIdent   TokenType = "IDENT"
	TokInt     TokenType = "INT"
	TokFloat   TokenType = "FLOAT"
	TokString  TokenType = "STRING"
	TokBool    TokenType = "BOOL"

	// Punctuation
	TokLBrace  TokenType = "{"
	TokRBrace  TokenType = "}"
	TokLParen  TokenType = "("
	TokRParen  TokenType = ")"
	TokLBrack  TokenType = "["
	TokRBrack  TokenType = "]"
	TokComma   TokenType = ","
	TokDot     TokenType = "."
	TokColon   TokenType = ":"
	TokSemi    TokenType = ";"
	TokAssign  TokenType = "="
	TokPlus    TokenType = "+"
	TokMinus   TokenType = "-"
	TokStar    TokenType = "*"
	TokSlash   TokenType = "/"
	TokPercent TokenType = "%"
	TokBang    TokenType = "!"
	TokLt      TokenType = "<"
	TokGt      TokenType = ">"
	TokLtEq    TokenType = "<="
	TokGtEq    TokenType = ">="
	TokEqEq    TokenType = "=="
	TokBangEq  TokenType = "!="
	TokAmpAmp  TokenType = "&&"
	TokPipePipe TokenType = "||"
	TokPlusEq  TokenType = "+="
	TokMinusEq TokenType = "-="

	// Keywords
	TokObject   TokenType = "object"
	TokFunc     TokenType = "func"
	TokAsync    TokenType = "async"
	TokAwait    TokenType = "await"
	TokReturn   TokenType = "return"
	TokIf       TokenType = "if"
	TokElse     TokenType = "else"
	TokFor      TokenType = "for"
	TokWhile    TokenType = "while"
	TokBreak    TokenType = "break"
	TokContinue TokenType = "continue"
	TokConst    TokenType = "const"
	TokVar      TokenType = "var"
	TokTrue     TokenType = "true"
	TokFalse    TokenType = "false"
	TokNull     TokenType = "null"
	TokImport   TokenType = "import"
	TokFrom     TokenType = "from"
	TokThis     TokenType = "this"
	TokEmit     TokenType = "emit"
)

var keywords = map[string]TokenType{
	"object":   TokObject,
	"func":     TokFunc,
	"async":    TokAsync,
	"await":    TokAwait,
	"return":   TokReturn,
	"if":       TokIf,
	"else":     TokElse,
	"for":      TokFor,
	"while":    TokWhile,
	"break":    TokBreak,
	"continue": TokContinue,
	"const":    TokConst,
	"var":      TokVar,
	"true":     TokTrue,
	"false":    TokFalse,
	"null":     TokNull,
	"import":   TokImport,
	"from":     TokFrom,
	"this":     TokThis,
	"emit":     TokEmit,
}

// Token is a single lexical unit.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %q, %d:%d)", t.Type, t.Literal, t.Line, t.Col)
}

// Lexer tokenises a KScript source string.
type Lexer struct {
	src    []rune
	pos    int
	line   int
	col    int
	tokens []Token
}

// New creates a Lexer for the given source.
func New(src string) *Lexer {
	return &Lexer{src: []rune(src), line: 1, col: 1}
}

// Tokenise processes the entire source and returns all tokens.
func (l *Lexer) Tokenise() ([]Token, error) {
	for {
		tok, err := l.next()
		if err != nil {
			return nil, err
		}
		l.tokens = append(l.tokens, tok)
		if tok.Type == TokEOF {
			break
		}
	}
	return l.tokens, nil
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) advance() rune {
	ch := l.src[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) skipWhitespaceAndComments() {
	for l.pos < len(l.src) {
		ch := l.peek()
		if ch == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '/' {
			// Line comment.
			for l.pos < len(l.src) && l.peek() != '\n' {
				l.advance()
			}
		} else if unicode.IsSpace(ch) {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) next() (Token, error) {
	l.skipWhitespaceAndComments()
	if l.pos >= len(l.src) {
		return Token{Type: TokEOF, Line: l.line, Col: l.col}, nil
	}

	line, col := l.line, l.col
	ch := l.advance()

	switch {
	case unicode.IsLetter(ch) || ch == '_':
		return l.readIdent(ch, line, col), nil
	case unicode.IsDigit(ch):
		return l.readNumber(ch, line, col), nil
	case ch == '"':
		return l.readString(line, col)
	}

	// Two-character tokens.
	if l.pos < len(l.src) {
		next := l.peek()
		two := string([]rune{ch, next})
		switch two {
		case "<=":
			l.advance()
			return Token{Type: TokLtEq, Literal: two, Line: line, Col: col}, nil
		case ">=":
			l.advance()
			return Token{Type: TokGtEq, Literal: two, Line: line, Col: col}, nil
		case "==":
			l.advance()
			return Token{Type: TokEqEq, Literal: two, Line: line, Col: col}, nil
		case "!=":
			l.advance()
			return Token{Type: TokBangEq, Literal: two, Line: line, Col: col}, nil
		case "&&":
			l.advance()
			return Token{Type: TokAmpAmp, Literal: two, Line: line, Col: col}, nil
		case "||":
			l.advance()
			return Token{Type: TokPipePipe, Literal: two, Line: line, Col: col}, nil
		case "+=":
			l.advance()
			return Token{Type: TokPlusEq, Literal: two, Line: line, Col: col}, nil
		case "-=":
			l.advance()
			return Token{Type: TokMinusEq, Literal: two, Line: line, Col: col}, nil
		}
	}

	// Single-character tokens.
	singleMap := map[rune]TokenType{
		'{': TokLBrace, '}': TokRBrace,
		'(': TokLParen, ')': TokRParen,
		'[': TokLBrack, ']': TokRBrack,
		',': TokComma, '.': TokDot, ':': TokColon, ';': TokSemi,
		'=': TokAssign, '+': TokPlus, '-': TokMinus,
		'*': TokStar, '/': TokSlash, '%': TokPercent,
		'!': TokBang, '<': TokLt, '>': TokGt,
	}
	if tt, ok := singleMap[ch]; ok {
		return Token{Type: tt, Literal: string(ch), Line: line, Col: col}, nil
	}

	return Token{}, fmt.Errorf("unexpected character %q at %d:%d", ch, line, col)
}

func (l *Lexer) readIdent(first rune, line, col int) Token {
	var sb strings.Builder
	sb.WriteRune(first)
	for l.pos < len(l.src) && (unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_') {
		sb.WriteRune(l.advance())
	}
	lit := sb.String()
	tt, ok := keywords[lit]
	if !ok {
		tt = TokIdent
	}
	return Token{Type: tt, Literal: lit, Line: line, Col: col}
}

func (l *Lexer) readNumber(first rune, line, col int) Token {
	var sb strings.Builder
	sb.WriteRune(first)
	tt := TokInt
	for l.pos < len(l.src) && (unicode.IsDigit(l.peek()) || l.peek() == '.') {
		if l.peek() == '.' {
			tt = TokFloat
		}
		sb.WriteRune(l.advance())
	}
	return Token{Type: tt, Literal: sb.String(), Line: line, Col: col}
}

func (l *Lexer) readString(line, col int) (Token, error) {
	var sb strings.Builder
	for l.pos < len(l.src) {
		ch := l.advance()
		if ch == '"' {
			return Token{Type: TokString, Literal: sb.String(), Line: line, Col: col}, nil
		}
		if ch == '\n' {
			return Token{}, fmt.Errorf("unterminated string at %d:%d", line, col)
		}
		sb.WriteRune(ch)
	}
	return Token{}, fmt.Errorf("unterminated string at %d:%d", line, col)
}
