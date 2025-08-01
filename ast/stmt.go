package ast

type Stmt interface {
	Accept(visitor StmtVisitor) interface{}
}

type ExpressionStmt struct {
	Expr Expr
}

func (b ExpressionStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitExpressionStmt(b)
}

type PrintStmt struct {
	Expr Expr
}

func (b PrintStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitPrintStmt(b)
}

type VarStmt struct {
	Name        Token
	Initializer Expr
}

func (b VarStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitVarStmt(b)
}

type BlockStmt struct {
	Statements []Stmt
}

func (b BlockStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitBlockStmt(b)
}

type IfStmt struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (b IfStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitIfStmt(b)
}

type WhileStmt struct {
	Condition Expr
	Body      Stmt
}

func (b WhileStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitWhileStmt(b)
}

type FunctionStmt struct {
	Name   Token
	Params []Token
	Body   []Stmt
}

func (b FunctionStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitFunctionStmt(b)
}

type ReturnStmt struct {
	Keyword Token
	Value   Expr
}

func (b ReturnStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitReturnStmt(b)
}

type ClassStmt struct {
	Name       Token
	Methods    []FunctionStmt
	Superclass *VariableExpr
}

func (b ClassStmt) Accept(visitor StmtVisitor) interface{} {
	return visitor.VisitClassStmt(b)
}

type StmtVisitor interface {
	VisitExpressionStmt(stmt ExpressionStmt) interface{}
	VisitPrintStmt(stmt PrintStmt) interface{}
	VisitVarStmt(stmt VarStmt) interface{}
	VisitBlockStmt(stmt BlockStmt) interface{}
	VisitIfStmt(stmt IfStmt) interface{}
	VisitWhileStmt(stmt WhileStmt) interface{}
	VisitFunctionStmt(stmt FunctionStmt) interface{}
	VisitReturnStmt(stmt ReturnStmt) interface{}
	VisitClassStmt(stmt ClassStmt) interface{}
}
