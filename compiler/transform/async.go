// Package transform rewrites async methods into state-machine structs
// that implement core/async.Task.
//
// Each `await` point becomes a new state transition. Local variables
// that are live across await points are promoted to struct fields so
// they survive between ticks.
//
// Example input (KScript AST):
//
//	async create() {
//	  await wait(0.5)
//	  this.hp = 5
//	  await tween(this, 0.3)
//	}
//
// Output representation (stored as AsyncMethod on the MethodDecl):
//
//	States[0]: { stmts: [],           await: wait(0.5)    }
//	States[1]: { stmts: [this.hp=5],  await: tween(...)   }
//	States[2]: { stmts: [],           await: nil (done)   }
package transform

import "github.com/ElioNeto/kora/compiler/ast"

// State is one step in the compiled async state machine.
type State struct {
	// Stmts are the synchronous statements that run when this state is entered.
	Stmts []ast.Stmt
	// Await is the Task expression to wait on before advancing to the next state.
	// Nil means this is the terminal state.
	Await ast.Expr
}

// AsyncMethod holds the state-machine representation of an async method.
type AsyncMethod struct {
	Name   string
	Params []*ast.Param
	States []*State
	// LiveVars are variables that must survive across await boundaries.
	LiveVars []LiveVar
}

// LiveVar is a variable that must be promoted to a struct field.
type LiveVar struct {
	Name string
	Type string
}

// TransformProgram rewrites all async methods in a program and attaches
// the AsyncMethod metadata to each MethodDecl via the returned map.
func TransformProgram(prog *ast.Program) map[*ast.MethodDecl]*AsyncMethod {
	result := make(map[*ast.MethodDecl]*AsyncMethod)
	for _, obj := range prog.Objects {
		for _, m := range obj.Methods {
			if m.Async {
				result[m] = transformMethod(m)
			}
		}
	}
	return result
}

func transformMethod(m *ast.MethodDecl) *AsyncMethod {
	am := &AsyncMethod{
		Name:   m.Name,
		Params: m.Params,
	}

	// Split the body at every AwaitStmt.
	current := &State{}
	for _, stmt := range m.Body {
		if aw, ok := stmt.(*ast.AwaitStmt); ok {
			current.Await = aw.Task
			am.States = append(am.States, current)
			current = &State{}
		} else {
			current.Stmts = append(current.Stmts, stmt)
		}
	}
	// Terminal state: remaining statements, no await.
	am.States = append(am.States, current)

	// Collect live vars (variables declared before an await that are used after).
	am.LiveVars = collectLiveVars(m.Body)

	return am
}

// collectLiveVars finds variables declared before an await that may be
// referenced after it. For v0.1 we promote ALL var declarations that
// appear before at least one await — conservative but correct.
func collectLiveVars(body []ast.Stmt) []LiveVar {
	var live []LiveVar
	seenAwait := false
	declaredBeforeAwait := map[string]string{} // name -> type

	for _, stmt := range body {
		switch s := stmt.(type) {
		case *ast.AwaitStmt:
			seenAwait = true
		case *ast.VarStmt:
			if !seenAwait {
				declaredBeforeAwait[s.Name] = s.Type
			}
		}
	}

	if seenAwait {
		for name, typ := range declaredBeforeAwait {
			if typ == "" {
				typ = "any"
			}
			live = append(live, LiveVar{Name: name, Type: typ})
		}
	}
	return live
}
