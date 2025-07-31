package resolve

import (
	"fmt"
	"io"

	"github.com/Pra1tik/golox/ast"
	"github.com/Pra1tik/golox/interpret"
)

type scope map[string]bool

func (s scope) declare(name string, token ast.Token) {
	s[name] = false
}

func (s scope) define(name string) {
	s[name] = true
}

func (s scope) has(name string) (declared bool, defined bool) {
	v, ok := s[name]
	if !ok {
		return false, false
	}
	return true, v
}

func (s scope) set(name string) {
	s[name] = true
}

type scopes []scope

func (s *scopes) push(scope scope) {
	*s = append(*s, scope)
}

func (s *scopes) pop() {
	*s = (*s)[:len(*s)-1]
}

func (s *scopes) peek() scope {
	return (*s)[len(*s)-1]
}

type functionType int

const (
	functionTypeNone functionType = iota
	functionTypeFunction
	functionTypeMethod
	functionTypeInitializer
)

type classType int

const (
	classTypeNone classType = iota
	classTypeClass
	classTypeSubClass
)

type Resolver struct {
	interpreter *interpret.Interpreter

	scopes          scopes
	currentFunction functionType
	currentClass    classType

	stdErr   io.Writer
	hadError bool
}

func CreateResolver(interpreter *interpret.Interpreter, stdErr io.Writer) *Resolver {
	return &Resolver{interpreter: interpreter, stdErr: stdErr}
}

func (r *Resolver) VisitBlockStmt(stmt ast.BlockStmt) interface{} {
	r.beginScope()
	r.ResolveStmts(stmt.Statements)
	r.endScope()
	return nil
}

func (r *Resolver) ResolveStmts(statements []ast.Stmt) (hadError bool) {
	for _, statement := range statements {
		r.resolveStmt(statement)
	}
	return r.hadError
}

func (r *Resolver) resolveStmt(stmt ast.Stmt) {
	stmt.Accept(r)
}

func (r *Resolver) resolveExpr(expr ast.Expr) {
	expr.Accept(r)
}

func (r *Resolver) VisitVarStmt(stmt ast.VarStmt) interface{} {
	r.declare(stmt.Name)
	if stmt.Initializer != nil {
		r.resolveExpr(stmt.Initializer)
	}
	r.define(stmt.Name)
	return nil
}

func (r *Resolver) VisitFunctionStmt(stmt ast.FunctionStmt) interface{} {
	r.declare(stmt.Name)
	r.define(stmt.Name)

	r.resolveFunction(stmt, functionTypeFunction)
	return nil
}

func (r *Resolver) VisitClassStmt(stmt ast.ClassStmt) interface{} {
	enclosingClass := r.currentClass
	defer func() { r.currentClass = enclosingClass }()
	r.currentClass = classTypeClass

	r.declare(stmt.Name)
	r.define(stmt.Name)

	if stmt.Superclass != nil && stmt.Name.Lexeme == stmt.Superclass.Name.Lexeme {
		r.error(stmt.Superclass.Name, "A class can't inherit from itself.")
	}
	if stmt.Superclass != nil {
		r.currentClass = classTypeSubClass
		r.resolveExpr(stmt.Superclass)
	}

	if stmt.Superclass != nil {
		r.beginScope()
		defer func() { r.endScope() }()
		r.scopes.peek().set("super")
	}

	r.beginScope()
	r.scopes.peek().set("this")

	for _, method := range stmt.Methods {
		declaration := functionTypeMethod
		if method.Name.Lexeme == "init" {
			declaration = functionTypeInitializer
		}
		r.resolveFunction(method, declaration)
	}

	r.endScope()

	return nil
}

func (r *Resolver) VisitExpressionStmt(stmt ast.ExpressionStmt) interface{} {
	r.resolveExpr(stmt.Expr)
	return nil
}

func (r *Resolver) VisitIfStmt(stmt ast.IfStmt) interface{} {
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.ThenBranch)
	if stmt.ElseBranch != nil {
		r.resolveStmt((stmt.ElseBranch))
	}
	return nil
}

func (r *Resolver) VisitPrintStmt(stmt ast.PrintStmt) interface{} {
	r.resolveExpr(stmt.Expr)
	return nil
}

func (r *Resolver) VisitReturnStmt(stmt ast.ReturnStmt) interface{} {
	if r.currentFunction == functionTypeNone {
		r.error(stmt.Keyword, "Can't return from top-level code.")
	}

	if stmt.Value != nil {
		if r.currentFunction == functionTypeInitializer {
			r.error(stmt.Keyword, "Can't return value from initializer.")
		}
		r.resolveExpr(stmt.Value)
	}
	return nil
}

func (r *Resolver) VisitWhileStmt(stmt ast.WhileStmt) interface{} {
	r.resolveExpr(stmt.Condition)
	r.resolveStmt(stmt.Body)
	return nil
}

func (r *Resolver) VisitVariableExpr(expr ast.VariableExpr) interface{} {
	if len(r.scopes) > 0 {
		if declared, defined := r.scopes.peek().has(expr.Name.Lexeme); declared && !defined {
			r.error(expr.Name, "Can't read local variable in its own initializer.")
		}
	}

	r.resolveLocal(expr, expr.Name)
	return nil
}

func (r *Resolver) VisitAssignExpr(expr ast.AssignExpr) interface{} {
	r.resolveExpr(expr.Value)
	r.resolveLocal(expr, expr.Name)
	return nil
}

func (r *Resolver) VisitBinaryExpr(expr ast.BinaryExpr) interface{} {
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)
	return nil
}

func (r *Resolver) VisitCallExpr(expr ast.CallExpr) interface{} {
	r.resolveExpr(expr.Callee)
	for _, argument := range expr.Arguments {
		r.resolveExpr(argument)
	}
	return nil
}

func (r *Resolver) VisitGroupingExpr(expr ast.GroupingExpr) interface{} {
	r.resolveExpr(expr.Expression)
	return nil
}

func (r *Resolver) VisitLiteralExpr(expr ast.LiteralExpr) interface{} {
	return nil
}

func (r *Resolver) VisitThisExpr(expr ast.ThisExpr) interface{} {
	if r.currentClass == classTypeNone {
		r.error(expr.Keyword, "Can't use 'this' outside of a class.")
	}

	r.resolveLocal(expr, expr.Keyword)
	return nil
}

func (r *Resolver) VisitSuperExpr(expr ast.SuperExpr) interface{} {
	if r.currentClass == classTypeNone {
		r.error(expr.Keyword, "Can't use 'super' outside of a class.")
	} else if r.currentClass != classTypeSubClass {
		r.error(expr.Keyword, "Can't use 'super' in a class with no superclass")
	}

	r.resolveLocal(expr, expr.Keyword)
	return nil
}

func (r *Resolver) VisitLogicalExpr(expr ast.LogicalExpr) interface{} {
	r.resolveExpr(expr.Left)
	r.resolveExpr(expr.Right)
	return nil
}

func (r *Resolver) VisitUnaryExpr(expr ast.UnaryExpr) interface{} {
	r.resolveExpr(expr.Right)
	return nil
}

func (r *Resolver) VisitGetExpr(expr ast.GetExpr) interface{} {
	r.resolveExpr(expr.Object) // property itself is dynamically evaluated so no need to resolve expr.Name
	return nil
}

func (r *Resolver) VisitSetExpr(expr ast.SetExpr) interface{} {
	r.resolveExpr(expr.Value)
	r.resolveExpr(expr.Object)
	return nil
}

func (r *Resolver) resolveLocal(expr ast.Expr, name ast.Token) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		s := r.scopes[i]
		if _, defined := s.has(name.Lexeme); defined {
			depth := len(r.scopes) - 1 - i
			r.interpreter.Resolve(expr, depth)
			return
		}
	}
}

func (r *Resolver) resolveFunction(function ast.FunctionStmt, fnType functionType) {
	enclosingFunction := r.currentFunction
	r.currentFunction = fnType
	defer func() { r.currentFunction = enclosingFunction }()

	r.beginScope()
	for _, param := range function.Params {
		r.declare(param)
		r.define(param)
	}
	r.ResolveStmts(function.Body)
	r.endScope()
}

func (r *Resolver) declare(name ast.Token) {
	if len(r.scopes) == 0 {
		return
	}

	scope := r.scopes.peek()
	if _, defined := scope.has(name.Lexeme); defined {
		r.error(name, "Already variable with this name in this scope")
	}

	scope.declare(name.Lexeme, name)
}

func (r *Resolver) define(name ast.Token) {
	if len(r.scopes) == 0 {
		return
	}

	r.scopes.peek().define(name.Lexeme)
}

func (r *Resolver) beginScope() {
	r.scopes.push(make(scope))
}

func (r *Resolver) endScope() {
	r.scopes.pop()
}

func (r *Resolver) error(token ast.Token, message string) {
	var where string
	if token.TokenType == ast.TokenEof {
		where = " at end"
	} else {
		where = " at '" + token.Lexeme + "'"
	}

	_, _ = r.stdErr.Write([]byte(fmt.Sprintf("[line %d] Error%s: %s\n", token.Line, where, message)))
	r.hadError = true
}
