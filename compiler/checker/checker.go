// Package checker performs semantic analysis and type checking on a KScript AST.
//
// Responsibilities:
//   - Resolve all identifier references.
//   - Infer and validate types for expressions and statements.
//   - Ensure `await` only appears inside `async` methods.
//   - Detect use of undeclared variables.
//   - Verify call argument counts.
//   - Annotate the AST with resolved types (stored in TypeInfo).
package checker

import (
	"fmt"

	"github.com/ElioNeto/kora/compiler/ast"
)

// ----------------------------------------------------------------------------
// Type constants
// ----------------------------------------------------------------------------

const (
	TyInt    = "int"
	TyFloat  = "float"
	TyBool   = "bool"
	TyString = "string"
	TyVoid   = "void"
	TyTask   = "Task"
	TyEntity = "Entity"
	TyVec2   = "Vec2"
	TyAny    = "any" // internal: unresolved / permissive
)

// builtinTypes is the set of valid KScript primitive type names.
var builtinTypes = map[string]bool{
	TyInt: true, TyFloat: true, TyBool: true, TyString: true,
	TyVoid: true, TyTask: true, TyEntity: true, TyVec2: true,
	"Sprite": true, "Sound": true, "Signal": true,
	"int[]": true, "float[]": true, "string[]": true, "Entity[]": true,
}

// ----------------------------------------------------------------------------
// Engine API signatures
// ----------------------------------------------------------------------------

type funcSig struct {
	params []string
	ret    string
}

// engineAPI describes the built-in KScript functions available to user code.
var engineAPI = map[string]funcSig{
	"wait":        {params: []string{TyFloat}, ret: TyTask},
	"waitFrames":  {params: []string{TyInt}, ret: TyTask},
	"waitSignal":  {params: []string{TyEntity, TyString}, ret: TyTask},
	"tween":       {params: []string{TyEntity, TyAny, TyFloat}, ret: TyTask},
	"race":        {params: nil, ret: TyTask},   // variadic
	"all":         {params: nil, ret: TyTask},   // variadic
	"cancel":      {params: []string{TyTask}, ret: TyVoid},
}

// engineModules lists valid import sources.
var engineModules = map[string]bool{
	"kora": true,
}

// ----------------------------------------------------------------------------
// Scope
// ----------------------------------------------------------------------------

// varInfo holds a declared variable's type and mutability.
type varInfo struct {
	typ     string
	mutable bool
}

// scope is a simple chained symbol table.
type scope struct {
	vars   map[string]varInfo
	parent *scope
}

func newScope(parent *scope) *scope {
	return &scope{vars: make(map[string]varInfo), parent: parent}
}

func (s *scope) declare(name, typ string, mutable bool) error {
	if _, exists := s.vars[name]; exists {
		return fmt.Errorf("variable %q already declared in this scope", name)
	}
	s.vars[name] = varInfo{typ: typ, mutable: mutable}
	return nil
}

func (s *scope) lookup(name string) (varInfo, bool) {
	if v, ok := s.vars[name]; ok {
		return v, true
	}
	if s.parent != nil {
		return s.parent.lookup(name)
	}
	return varInfo{}, false
}

// ----------------------------------------------------------------------------
// TypeInfo — annotation map
// ----------------------------------------------------------------------------

// TypeInfo stores the resolved type for each expression node.
type TypeInfo map[ast.Expr]string

// ----------------------------------------------------------------------------
// Checker
// ----------------------------------------------------------------------------

// Error represents a single type-checking or semantic error.
type Error struct {
	Message string
}

func (e *Error) Error() string { return e.Message }

// Checker walks a Program AST and collects semantic errors.
type Checker struct {
	program  *ast.Program
	objects  map[string]*ast.ObjectDecl // object name → decl
	Types    TypeInfo
	Errors   []*Error
}

// New creates a Checker for the given program.
func New(prog *ast.Program) *Checker {
	return &Checker{
		program: prog,
		objects: make(map[string]*ast.ObjectDecl),
		Types:   make(TypeInfo),
	}
}

// Check runs all semantic passes and returns any errors found.
func (c *Checker) Check() []*Error {
	c.collectObjects()
	c.checkImports()
	for _, obj := range c.program.Objects {
		c.checkObject(obj)
	}
	return c.Errors
}

// ----------------------------------------------------------------------------
// Pass 1: collect object names
// ----------------------------------------------------------------------------

func (c *Checker) collectObjects() {
	for _, obj := range c.program.Objects {
		if _, dup := c.objects[obj.Name]; dup {
			c.errorf("duplicate object declaration: %q", obj.Name)
		}
		c.objects[obj.Name] = obj
	}
}

// ----------------------------------------------------------------------------
// Pass 2: imports
// ----------------------------------------------------------------------------

func (c *Checker) checkImports() {
	for _, imp := range c.program.Imports {
		if !engineModules[imp.Module] {
			c.errorf("unknown module %q — only engine modules may be imported", imp.Module)
		}
	}
}

// ----------------------------------------------------------------------------
// Pass 3: objects
// ----------------------------------------------------------------------------

func (c *Checker) checkObject(obj *ast.ObjectDecl) {
	// Build an object-level scope with field names.
	objScope := newScope(nil)
	for _, f := range obj.Fields {
		if !c.isValidType(f.Type) {
			c.errorf("object %s: field %q has unknown type %q", obj.Name, f.Name, f.Type)
		}
		_ = objScope.declare(f.Name, f.Type, true)
	}

	// Check field defaults.
	for _, f := range obj.Fields {
		if f.Default != nil {
			c.checkExpr(f.Default, objScope)
		}
	}

	// Check methods.
	for _, m := range obj.Methods {
		c.checkMethod(obj, m, objScope)
	}
}

func (c *Checker) checkMethod(obj *ast.ObjectDecl, m *ast.MethodDecl, objScope *scope) {
	if m.Return != "" && !c.isValidType(m.Return) {
		c.errorf("object %s.%s: unknown return type %q", obj.Name, m.Name, m.Return)
	}

	// Method scope inherits object scope.
	methodScope := newScope(objScope)
	for _, p := range m.Params {
		if !c.isValidType(p.Type) {
			c.errorf("object %s.%s: param %q has unknown type %q", obj.Name, m.Name, p.Name, p.Type)
		}
		_ = methodScope.declare(p.Name, p.Type, false)
	}

	c.checkBlock(m.Body, methodScope, m.Async, obj)
}

// ----------------------------------------------------------------------------
// Statement checking
// ----------------------------------------------------------------------------

func (c *Checker) checkBlock(stmts []ast.Stmt, s *scope, inAsync bool, obj *ast.ObjectDecl) {
	blockScope := newScope(s)
	for _, stmt := range stmts {
		c.checkStmt(stmt, blockScope, inAsync, obj)
	}
}

func (c *Checker) checkStmt(stmt ast.Stmt, s *scope, inAsync bool, obj *ast.ObjectDecl) {
	switch st := stmt.(type) {
	case *ast.VarStmt:
		var typ string
		if st.Value != nil {
			typ = c.checkExpr(st.Value, s)
		}
		if st.Type != "" {
			if !c.isValidType(st.Type) {
				c.errorf("unknown type %q", st.Type)
			}
			typ = st.Type
		}
		if typ == "" {
			typ = TyAny
		}
		if err := s.declare(st.Name, typ, !st.Const); err != nil {
			c.errorf("%s", err.Error())
		}

	case *ast.AssignStmt:
		// Ensure LHS is assignable.
		c.checkAssignTarget(st.Target, s)
		c.checkExpr(st.Value, s)

	case *ast.ReturnStmt:
		if st.Value != nil {
			c.checkExpr(st.Value, s)
		}

	case *ast.ExprStmt:
		c.checkExpr(st.X, s)

	case *ast.IfStmt:
		condType := c.checkExpr(st.Cond, s)
		if condType != TyBool && condType != TyAny {
			c.errorf("if condition must be bool, got %s", condType)
		}
		c.checkBlock(st.Then, s, inAsync, obj)
		if len(st.Else) > 0 {
			c.checkBlock(st.Else, s, inAsync, obj)
		}

	case *ast.WhileStmt:
		condType := c.checkExpr(st.Cond, s)
		if condType != TyBool && condType != TyAny {
			c.errorf("while condition must be bool, got %s", condType)
		}
		c.checkBlock(st.Body, s, inAsync, obj)

	case *ast.AwaitStmt:
		if !inAsync {
			c.errorf("`await` used outside of an async method")
		}
		taskType := c.checkExpr(st.Task, s)
		if taskType != TyTask && taskType != TyAny {
			c.errorf("`await` expression must return Task, got %s", taskType)
		}

	case *ast.EmitStmt:
		if st.Signal == "" {
			c.errorf("`emit` requires a non-empty signal name")
		}

	case *ast.BreakStmt, *ast.ContinueStmt:
		// Always valid inside a loop; loop nesting check omitted for v0.1.
	}
}

func (c *Checker) checkAssignTarget(expr ast.Expr, s *scope) {
	switch e := expr.(type) {
	case *ast.Ident:
		v, ok := s.lookup(e.Name)
		if !ok {
			c.errorf("undeclared variable %q", e.Name)
			return
		}
		if !v.mutable {
			c.errorf("cannot assign to const %q", e.Name)
		}
	case *ast.MemberExpr:
		// e.g. this.x = ... — always allowed on Entity members.
	case *ast.IndexExpr:
		// array[i] = ...
	default:
		c.errorf("invalid assignment target")
	}
}

// ----------------------------------------------------------------------------
// Expression checking — returns resolved type string
// ----------------------------------------------------------------------------

func (c *Checker) checkExpr(expr ast.Expr, s *scope) string {
	var typ string
	switch e := expr.(type) {
	case *ast.IntLit:
		typ = TyInt
	case *ast.FloatLit:
		typ = TyFloat
	case *ast.BoolLit:
		typ = TyBool
	case *ast.StringLit:
		typ = TyString
	case *ast.NullLit:
		typ = TyAny
	case *ast.ThisExpr:
		typ = TyEntity
	case *ast.Ident:
		typ = c.resolveIdent(e, s)
	case *ast.BinaryExpr:
		typ = c.checkBinary(e, s)
	case *ast.UnaryExpr:
		typ = c.checkExpr(e.Operand, s)
		if e.Op == "!" {
			typ = TyBool
		}
	case *ast.CallExpr:
		typ = c.checkCall(e, s)
	case *ast.MemberExpr:
		typ = c.checkMember(e, s)
	case *ast.IndexExpr:
		c.checkExpr(e.Object, s)
		c.checkExpr(e.Index, s)
		typ = TyAny // array element type resolution in v0.2
	default:
		typ = TyAny
	}
	c.Types[expr] = typ
	return typ
}

func (c *Checker) resolveIdent(e *ast.Ident, s *scope) string {
	// Built-in engine namespaces.
	switch e.Name {
	case "Input", "Audio", "Scene", "Math":
		return TyAny // namespace — member resolution handled in checkMember
	}
	// Local / field scope.
	if v, ok := s.lookup(e.Name); ok {
		return v.typ
	}
	// Object reference.
	if _, ok := c.objects[e.Name]; ok {
		return e.Name
	}
	c.errorf("undeclared identifier %q", e.Name)
	return TyAny
}

func (c *Checker) checkBinary(e *ast.BinaryExpr, s *scope) string {
	lt := c.checkExpr(e.Left, s)
	rt := c.checkExpr(e.Right, s)
	switch e.Op {
	case "==", "!=", "<", ">", "<=", ">=":
		return TyBool
	case "&&", "||":
		if lt != TyBool && lt != TyAny {
			c.errorf("logical operator requires bool, got %s", lt)
		}
		if rt != TyBool && rt != TyAny {
			c.errorf("logical operator requires bool, got %s", rt)
		}
		return TyBool
	case "+":
		if lt == TyString || rt == TyString {
			return TyString
		}
		return c.numericResult(lt, rt)
	case "-", "*", "/", "%":
		return c.numericResult(lt, rt)
	}
	return TyAny
}

func (c *Checker) numericResult(lt, rt string) string {
	if lt == TyAny || rt == TyAny {
		return TyAny
	}
	if lt == TyFloat || rt == TyFloat {
		return TyFloat
	}
	return TyInt
}

func (c *Checker) checkCall(e *ast.CallExpr, s *scope) string {
	// Direct function call: wait(...), race(...), etc.
	if ident, ok := e.Callee.(*ast.Ident); ok {
		if sig, found := engineAPI[ident.Name]; found {
			if sig.params != nil && len(e.Args) != len(sig.params) {
				c.errorf("%s() expects %d argument(s), got %d", ident.Name, len(sig.params), len(e.Args))
			}
			for _, arg := range e.Args {
				c.checkExpr(arg, s)
			}
			return sig.ret
		}
	}
	// Method calls: callee is MemberExpr, e.g. this.destroy(), Input.pressed()
	for _, arg := range e.Args {
		c.checkExpr(arg, s)
	}
	c.checkExpr(e.Callee, s)
	return TyAny
}

func (c *Checker) checkMember(e *ast.MemberExpr, s *scope) string {
	c.checkExpr(e.Object, s)
	// Known namespace members with return types.
	if ident, ok := e.Object.(*ast.Ident); ok {
		switch ident.Name {
		case "Input":
			switch e.Prop {
			case "axisX", "axisY":
				return TyFloat
			case "pressed", "held", "released":
				return TyBool
			case "touchPos":
				return TyVec2
			}
		case "Audio":
			return TyVoid
		case "Scene":
			switch e.Prop {
			case "spawn", "find":
				return TyEntity
			}
			return TyVoid
		case "Math":
			return TyFloat
		}
	}
	return TyAny
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

func (c *Checker) isValidType(t string) bool {
	if t == "" {
		return true // untyped is allowed (inferred)
	}
	if builtinTypes[t] {
		return true
	}
	// User-defined object type.
	if _, ok := c.objects[t]; ok {
		return true
	}
	return false
}

func (c *Checker) errorf(format string, args ...any) {
	c.Errors = append(c.Errors, &Error{Message: fmt.Sprintf(format, args...)})
}
