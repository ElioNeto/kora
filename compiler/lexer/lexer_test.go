package lexer_test

import (
	"testing"

	"github.com/ElioNeto/kora/compiler/lexer"
)

func TestBasicTokens(t *testing.T) {
	src := `object Player { var hp: int = 5 }`
	l := lexer.New(src)
	tokens, err := l.Tokenise()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []lexer.TokenType{
		lexer.TokObject, lexer.TokIdent,
		lexer.TokLBrace,
		lexer.TokVar, lexer.TokIdent, lexer.TokColon, lexer.TokIdent, lexer.TokAssign, lexer.TokInt,
		lexer.TokRBrace,
		lexer.TokEOF,
	}
	for i, tt := range expected {
		if i >= len(tokens) {
			t.Fatalf("expected token %s at index %d, got nothing", tt, i)
		}
		if tokens[i].Type != tt {
			t.Errorf("token[%d]: expected %s, got %s (%q)", i, tt, tokens[i].Type, tokens[i].Literal)
		}
	}
}

func TestAsyncAwait(t *testing.T) {
	src := `async func create() { await wait(0.5) }`
	l := lexer.New(src)
	tokens, err := l.Tokenise()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, tok := range tokens {
		if tok.Type == lexer.TokEOF {
			break
		}
		t.Logf("%s", tok)
	}
}
