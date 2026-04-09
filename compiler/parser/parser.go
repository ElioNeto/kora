// Package parser converts a KScript token stream into an AST.
package parser

import (
	"fmt"
	"strconv"

	"github.com/ElioNeto/kora/compiler/ast"
	"github.com/ElioNeto/kora/compiler/lexer"
)

// Parser holds the token stream and current position.
type Parser struct {
	tokens []lexer.Token
	pos    int
	errors []string
}

// New creates a Parser from a token slice (as returned by lexer.Tokenise).
func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

// Errors returns all parse errors collected during parsing.
func (p *Parser) Errors() []string { return p.errors }

// Parse processes the full token stream and returns the Program AST node.
func (p *Parser) Parse() (*ast.Program, error) {
	prog := &ast.Program{}
	for !p.isEOF() {
		switch p.peek().Type {
		case lexer.TokImport:
			imp, err := p.parseImport()
			if err != nil {
				return nil, err
			}
			prog.Imports = append(prog.Imports, imp)
		case lexer.TokObject:
			obj, err := p.parseObject()
			if err != nil {
				return nil, err
			}
			prog.Objects = append(prog.Objects, obj)
		default:
			return nil, p.errorf("unexpected token %s at top level", p.peek().Type)
		}
	}
	return prog, nil
}

// ---------------------------------------------------------------------------
// Top-level
// ---------------------------------------------------------------------------

func (p *Parser) parseImport() (*ast.ImportDecl, error) {
	if _, err := p.expect(lexer.TokImport); err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TokLBrace); err != nil {
		return nil, err
	}
	var symbols []string
	for !p.check(lexer.TokRBrace) && !p.isEOF() {
		tok, err := p.expect(lexer.TokIdent)
		if err != nil {
			return nil, err
		}
		symbols = append(symbols, tok.Literal)
		if p.check(lexer.TokComma) {
			p.advance()
		}
	}
	if _, err := p.expect(lexer.TokRBrace); err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TokFrom); err != nil {
		return nil, err
	}
	mod, err := p.expect(lexer.TokString)
	if err != nil {
		return nil, err
	}
	return &ast.ImportDecl{Module: mod.Literal, Symbols: symbols}, nil
}

func (p *Parser) parseObject() (*ast.ObjectDecl, error) {
	if _, err := p.expect(lexer.TokObject); err != nil {
		return nil, err
	}
	name, err := p.expect(lexer.TokIdent)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TokLBrace); err != nil {
		return nil, err
	}

	obj := &ast.ObjectDecl{Name: name.Literal}

	for !p.check(lexer.TokRBrace) && !p.isEOF() {
		switch {
		case p.check(lexer.TokVar) || p.check(lexer.TokConst):
			f, err := p.parseField()
			if err != nil {
				return nil, err
			}
			obj.Fields = append(obj.Fields, f)
		case p.check(lexer.TokAsync) || p.check(lexer.TokFunc) || p.checkIdent():
			m, err := p.parseMethod()
			if err != nil {
				return nil, err
			}
			obj.Methods = append(obj.Methods, m)
		default:
			return nil, p.errorf("unexpected token %s inside object %s", p.peek().Type, name.Literal)
		}
	}

	if _, err := p.expect(lexer.TokRBrace); err != nil {
		return nil, err
	}
	return obj, nil
}

// ---------------------------------------------------------------------------
// Fields and methods
// ---------------------------------------------------------------------------

func (p *Parser) parseField() (*ast.FieldDecl, error) {
	p.advance() // consume var / const
	name, err := p.expect(lexer.TokIdent)
	if err != nil {
		return nil, err
	}
	var typeName string
	if p.check(lexer.TokColon) {
		p.advance()
		t, err := p.expectType()
		if err != nil {
			return nil, err
		}
		typeName = t
	}
	var def ast.Expr
	if p.check(lexer.TokAssign) {
		p.advance()
		def, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}
	p.consumeSemi()
	return &ast.FieldDecl{Name: name.Literal, Type: typeName, Default: def}, nil
}

func (p *Parser) parseMethod() (*ast.MethodDecl, error) {
	isAsync := false
	if p.check(lexer.TokAsync) {
		isAsync = true
		p.advance()
	}
	// optional `func` keyword
	if p.check(lexer.TokFunc) {
		p.advance()
	}
	name, err := p.expect(lexer.TokIdent)
	if err != nil {
		return nil, err
	}
	params, err := p.parseParams()
	if err != nil {
		return nil, err
	}
	// optional return type
	retType := ""
	if p.check(lexer.TokColon) {
		p.advance()
		retType, err = p.expectType()
		if err != nil {
			return nil, err
		}
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.MethodDecl{
		Name:   name.Literal,
		Async:  isAsync,
		Params: params,
		Return: retType,
		Body:   body,
	}, nil
}

func (p *Parser) parseParams() ([]*ast.Param, error) {
	if _, err := p.expect(lexer.TokLParen); err != nil {
		return nil, err
	}
	var params []*ast.Param
	for !p.check(lexer.TokRParen) && !p.isEOF() {
		name, err := p.expect(lexer.TokIdent)
		if err != nil {
			return nil, err
		}
		var typeName string
		if p.check(lexer.TokColon) {
			p.advance()
			typeName, err = p.expectType()
			if err != nil {
				return nil, err
			}
		}
		params = append(params, &ast.Param{Name: name.Literal, Type: typeName})
		if p.check(lexer.TokComma) {
			p.advance()
		}
	}
	if _, err := p.expect(lexer.TokRParen); err != nil {
		return nil, err
	}
	return params, nil
}

// ---------------------------------------------------------------------------
// Block and statements
// ---------------------------------------------------------------------------

func (p *Parser) parseBlock() ([]ast.Stmt, error) {
	if _, err := p.expect(lexer.TokLBrace); err != nil {
		return nil, err
	}
	var stmts []ast.Stmt
	for !p.check(lexer.TokRBrace) && !p.isEOF() {
		s, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if s != nil {
			stmts = append(stmts, s)
		}
	}
	if _, err := p.expect(lexer.TokRBrace); err != nil {
		return nil, err
	}
	return stmts, nil
}

func (p *Parser) parseStmt() (ast.Stmt, error) {
	switch p.peek().Type {
	case lexer.TokVar, lexer.TokConst:
		return p.parseVarStmt()
	case lexer.TokReturn:
		return p.parseReturnStmt()
	case lexer.TokIf:
		return p.parseIfStmt()
	case lexer.TokWhile:
		return p.parseWhileStmt()
	case lexer.TokBreak:
		p.advance()
		p.consumeSemi()
		return &ast.BreakStmt{}, nil
	case lexer.TokContinue:
		p.advance()
		p.consumeSemi()
		return &ast.ContinueStmt{}, nil
	case lexer.TokAwait:
		return p.parseAwaitStmt()
	case lexer.TokEmit:
		return p.parseEmitStmt()
	default:
		return p.parseExprOrAssignStmt()
	}
}

func (p *Parser) parseVarStmt() (ast.Stmt, error) {
	isConst := p.peek().Type == lexer.TokConst
	p.advance()
	name, err := p.expect(lexer.TokIdent)
	if err != nil {
		return nil, err
	}
	var typeName string
	if p.check(lexer.TokColon) {
		p.advance()
		typeName, err = p.expectType()
		if err != nil {
			return nil, err
		}
	}
	var val ast.Expr
	if p.check(lexer.TokAssign) {
		p.advance()
		val, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}
	p.consumeSemi()
	return &ast.VarStmt{Const: isConst, Name: name.Literal, Type: typeName, Value: val}, nil
}

func (p *Parser) parseReturnStmt() (ast.Stmt, error) {
	p.advance()
	var val ast.Expr
	if !p.check(lexer.TokRBrace) && !p.check(lexer.TokSemi) && !p.isEOF() {
		var err error
		val, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}
	p.consumeSemi()
	return &ast.ReturnStmt{Value: val}, nil
}

func (p *Parser) parseIfStmt() (ast.Stmt, error) {
	p.advance()
	if _, err := p.expect(lexer.TokLParen); err != nil {
		return nil, err
	}
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TokRParen); err != nil {
		return nil, err
	}
	then, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	var els []ast.Stmt
	if p.check(lexer.TokElse) {
		p.advance()
		if p.check(lexer.TokIf) {
			nested, err := p.parseIfStmt()
			if err != nil {
				return nil, err
			}
			els = []ast.Stmt{nested}
		} else {
			els, err = p.parseBlock()
			if err != nil {
				return nil, err
			}
		}
	}
	return &ast.IfStmt{Cond: cond, Then: then, Else: els}, nil
}

func (p *Parser) parseWhileStmt() (ast.Stmt, error) {
	p.advance()
	if _, err := p.expect(lexer.TokLParen); err != nil {
		return nil, err
	}
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TokRParen); err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.WhileStmt{Cond: cond, Body: body}, nil
}

func (p *Parser) parseAwaitStmt() (ast.Stmt, error) {
	p.advance()
	task, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	p.consumeSemi()
	return &ast.AwaitStmt{Task: task}, nil
}

func (p *Parser) parseEmitStmt() (ast.Stmt, error) {
	p.advance()
	sig, err := p.expect(lexer.TokString)
	if err != nil {
		return nil, err
	}
	p.consumeSemi()
	return &ast.EmitStmt{Signal: sig.Literal}, nil
}

func (p *Parser) parseExprOrAssignStmt() (ast.Stmt, error) {
	lhs, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.checkAssignOp() {
		op := p.advance().Literal
		rhs, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		p.consumeSemi()
		return &ast.AssignStmt{Target: lhs, Op: op, Value: rhs}, nil
	}
	p.consumeSemi()
	return &ast.ExprStmt{X: lhs}, nil
}

// ---------------------------------------------------------------------------
// Expressions (Pratt / precedence climbing)
// ---------------------------------------------------------------------------

func (p *Parser) parseExpr() (ast.Expr, error) {
	return p.parseBinary(0)
}

var precedence = map[lexer.TokenType]int{
	lexer.TokPipePipe: 1,
	lexer.TokAmpAmp:  2,
	lexer.TokEqEq:    3,
	lexer.TokBangEq:  3,
	lexer.TokLt:      4,
	lexer.TokGt:      4,
	lexer.TokLtEq:    4,
	lexer.TokGtEq:    4,
	lexer.TokPlus:    5,
	lexer.TokMinus:   5,
	lexer.TokStar:    6,
	lexer.TokSlash:   6,
	lexer.TokPercent: 6,
}

func (p *Parser) parseBinary(minPrec int) (ast.Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		prec, ok := precedence[p.peek().Type]
		if !ok || prec <= minPrec {
			break
		}
		op := p.advance().Literal
		right, err := p.parseBinary(prec)
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseUnary() (ast.Expr, error) {
	if p.check(lexer.TokBang) || p.check(lexer.TokMinus) {
		op := p.advance().Literal
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Op: op, Operand: operand}, nil
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() (ast.Expr, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for {
		switch {
		case p.check(lexer.TokDot):
			p.advance()
			prop, err := p.expect(lexer.TokIdent)
			if err != nil {
				return nil, err
			}
			expr = &ast.MemberExpr{Object: expr, Prop: prop.Literal}
		case p.check(lexer.TokLParen):
			args, err := p.parseArgs()
			if err != nil {
				return nil, err
			}
			expr = &ast.CallExpr{Callee: expr, Args: args}
		case p.check(lexer.TokLBrack):
			p.advance()
			idx, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(lexer.TokRBrack); err != nil {
				return nil, err
			}
			expr = &ast.IndexExpr{Object: expr, Index: idx}
		default:
			return expr, nil
		}
	}
}

func (p *Parser) parsePrimary() (ast.Expr, error) {
	tok := p.peek()
	switch tok.Type {
	case lexer.TokInt:
		p.advance()
		v, _ := strconv.ParseInt(tok.Literal, 10, 64)
		return &ast.IntLit{Value: v}, nil
	case lexer.TokFloat:
		p.advance()
		v, _ := strconv.ParseFloat(tok.Literal, 64)
		return &ast.FloatLit{Value: v}, nil
	case lexer.TokString:
		p.advance()
		return &ast.StringLit{Value: tok.Literal}, nil
	case lexer.TokTrue:
		p.advance()
		return &ast.BoolLit{Value: true}, nil
	case lexer.TokFalse:
		p.advance()
		return &ast.BoolLit{Value: false}, nil
	case lexer.TokNull:
		p.advance()
		return &ast.NullLit{}, nil
	case lexer.TokThis:
		p.advance()
		return &ast.ThisExpr{}, nil
	case lexer.TokIdent:
		p.advance()
		return &ast.Ident{Name: tok.Literal}, nil
	case lexer.TokLParen:
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.TokRParen); err != nil {
			return nil, err
		}
		return expr, nil
	}
	return nil, p.errorf("unexpected token %s (%q) in expression", tok.Type, tok.Literal)
}

func (p *Parser) parseArgs() ([]ast.Expr, error) {
	if _, err := p.expect(lexer.TokLParen); err != nil {
		return nil, err
	}
	var args []ast.Expr
	for !p.check(lexer.TokRParen) && !p.isEOF() {
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		if p.check(lexer.TokComma) {
			p.advance()
		}
	}
	if _, err := p.expect(lexer.TokRParen); err != nil {
		return nil, err
	}
	return args, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (p *Parser) peek() lexer.Token {
	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		if tok.Type != lexer.TokEOF {
			return tok
		}
		break
	}
	return lexer.Token{Type: lexer.TokEOF}
}

func (p *Parser) advance() lexer.Token {
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

func (p *Parser) check(tt lexer.TokenType) bool {
	return p.peek().Type == tt
}

func (p *Parser) checkIdent() bool {
	return p.peek().Type == lexer.TokIdent
}

func (p *Parser) checkAssignOp() bool {
	tt := p.peek().Type
	return tt == lexer.TokAssign || tt == lexer.TokPlusEq || tt == lexer.TokMinusEq
}

func (p *Parser) isEOF() bool {
	return p.pos >= len(p.tokens) || p.tokens[p.pos].Type == lexer.TokEOF
}

func (p *Parser) expect(tt lexer.TokenType) (lexer.Token, error) {
	if p.peek().Type != tt {
		return lexer.Token{}, p.errorf("expected %s, got %s (%q)", tt, p.peek().Type, p.peek().Literal)
	}
	return p.advance(), nil
}

func (p *Parser) expectType() (string, error) {
	tok, err := p.expect(lexer.TokIdent)
	if err != nil {
		return "", err
	}
	name := tok.Literal
	// Handle array types: int[], string[], etc.
	if p.check(lexer.TokLBrack) {
		p.advance()
		if _, err := p.expect(lexer.TokRBrack); err != nil {
			return "", err
		}
		name += "[]"
	}
	return name, nil
}

func (p *Parser) consumeSemi() {
	if p.check(lexer.TokSemi) {
		p.advance()
	}
}

func (p *Parser) errorf(format string, args ...any) error {
	tok := p.peek()
	msg := fmt.Sprintf("[%d:%d] "+format, append([]any{tok.Line, tok.Col}, args...)...)
	p.errors = append(p.errors, msg)
	return fmt.Errorf("%s", msg)
}
