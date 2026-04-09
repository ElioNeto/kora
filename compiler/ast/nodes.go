// Package ast defines the Abstract Syntax Tree nodes for KScript.
package ast

// Node is the base interface for all AST nodes.
type Node interface {
	nodeType() string
}

// --- Top-level ---

// Program is the root of a KScript file.
type Program struct {
	Objects []*ObjectDecl
	Imports []*ImportDecl
}

func (p *Program) nodeType() string { return "Program" }

// ImportDecl represents: import { X } from "module"
type ImportDecl struct {
	Module  string
	Symbols []string
}

func (i *ImportDecl) nodeType() string { return "ImportDecl" }

// --- Object ---

// ObjectDecl represents: object Name { ... }
type ObjectDecl struct {
	Name    string
	Fields  []*FieldDecl
	Methods []*MethodDecl
}

func (o *ObjectDecl) nodeType() string { return "ObjectDecl" }

// FieldDecl represents a typed field with optional default: name: Type = expr
type FieldDecl struct {
	Name    string
	Type    string
	Default Expr
}

func (f *FieldDecl) nodeType() string { return "FieldDecl" }

// MethodDecl represents a (possibly async) method.
type MethodDecl struct {
	Name   string
	Async  bool
	Params []*Param
	Return string
	Body   []Stmt
}

func (m *MethodDecl) nodeType() string { return "MethodDecl" }

// Param is a method parameter: name: Type
type Param struct {
	Name string
	Type string
}

// --- Statements ---

// Stmt is implemented by all statement nodes.
type Stmt interface {
	Node
	stmtNode()
}

type VarStmt struct {
	Const bool
	Name  string
	Type  string // optional explicit type
	Value Expr
}
func (v *VarStmt) nodeType() string { return "VarStmt" }
func (v *VarStmt) stmtNode()        {}

type AssignStmt struct {
	Target Expr
	Op     string // "=", "+=", "-="
	Value  Expr
}
func (a *AssignStmt) nodeType() string { return "AssignStmt" }
func (a *AssignStmt) stmtNode()        {}

type ReturnStmt struct{ Value Expr }
func (r *ReturnStmt) nodeType() string { return "ReturnStmt" }
func (r *ReturnStmt) stmtNode()        {}

type ExprStmt struct{ X Expr }
func (e *ExprStmt) nodeType() string { return "ExprStmt" }
func (e *ExprStmt) stmtNode()        {}

type IfStmt struct {
	Cond Expr
	Then []Stmt
	Else []Stmt
}
func (i *IfStmt) nodeType() string { return "IfStmt" }
func (i *IfStmt) stmtNode()        {}

type WhileStmt struct {
	Cond Expr
	Body []Stmt
}
func (w *WhileStmt) nodeType() string { return "WhileStmt" }
func (w *WhileStmt) stmtNode()        {}

type BreakStmt struct{}
func (b *BreakStmt) nodeType() string { return "BreakStmt" }
func (b *BreakStmt) stmtNode()        {}

type ContinueStmt struct{}
func (c *ContinueStmt) nodeType() string { return "ContinueStmt" }
func (c *ContinueStmt) stmtNode()        {}

// AwaitStmt represents: await <expr>
type AwaitStmt struct{ Task Expr }
func (a *AwaitStmt) nodeType() string { return "AwaitStmt" }
func (a *AwaitStmt) stmtNode()        {}

type EmitStmt struct{ Signal string }
func (e *EmitStmt) nodeType() string { return "EmitStmt" }
func (e *EmitStmt) stmtNode()        {}

// --- Expressions ---

// Expr is implemented by all expression nodes.
type Expr interface {
	Node
	exprNode()
}

type IntLit struct{ Value int64 }
func (i *IntLit) nodeType() string { return "IntLit" }
func (i *IntLit) exprNode()        {}

type FloatLit struct{ Value float64 }
func (f *FloatLit) nodeType() string { return "FloatLit" }
func (f *FloatLit) exprNode()        {}

type StringLit struct{ Value string }
func (s *StringLit) nodeType() string { return "StringLit" }
func (s *StringLit) exprNode()        {}

type BoolLit struct{ Value bool }
func (b *BoolLit) nodeType() string { return "BoolLit" }
func (b *BoolLit) exprNode()        {}

type NullLit struct{}
func (n *NullLit) nodeType() string { return "NullLit" }
func (n *NullLit) exprNode()        {}

type Ident struct{ Name string }
func (id *Ident) nodeType() string { return "Ident" }
func (id *Ident) exprNode()        {}

type ThisExpr struct{}
func (t *ThisExpr) nodeType() string { return "ThisExpr" }
func (t *ThisExpr) exprNode()        {}

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}
func (b *BinaryExpr) nodeType() string { return "BinaryExpr" }
func (b *BinaryExpr) exprNode()        {}

type UnaryExpr struct {
	Op      string
	Operand Expr
}
func (u *UnaryExpr) nodeType() string { return "UnaryExpr" }
func (u *UnaryExpr) exprNode()        {}

type CallExpr struct {
	Callee Expr
	Args   []Expr
}
func (c *CallExpr) nodeType() string { return "CallExpr" }
func (c *CallExpr) exprNode()        {}

type MemberExpr struct {
	Object Expr
	Prop   string
}
func (m *MemberExpr) nodeType() string { return "MemberExpr" }
func (m *MemberExpr) exprNode()        {}

type IndexExpr struct {
	Object Expr
	Index  Expr
}
func (i *IndexExpr) nodeType() string { return "IndexExpr" }
func (i *IndexExpr) exprNode()        {}
