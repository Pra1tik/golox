package ast

// "Binary   : Expr left, Token operator, Expr right",
// "Grouping : Expr expression",
// "Literal  : Object value",
// "Unary    : Token operator, Expr right"
// "Logical  : Expr left, Token operator, Expr right"

type Expr interface {
	Accept(visitor ExprVisitor) interface{}
}

type BinaryExpr struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (b BinaryExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitBinaryExpr(b)
}

type GroupingExpr struct {
	Expression Expr
}

func (b GroupingExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitGroupingExpr(b)
}

type LiteralExpr struct {
	Value interface{}
}

func (b LiteralExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitLiteralExpr(b)
}

type UnaryExpr struct {
	Operator Token
	Right    Expr
}

func (b UnaryExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitUnaryExpr(b)
}

type LogicalExpr struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (b LogicalExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitLogicalExpr(b)
}

type VariableExpr struct {
	Name Token
}

func (b VariableExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitVariableExpr(b)
}

type AssignExpr struct {
	Name  Token
	Value Expr
}

func (b AssignExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitAssignExpr(b)
}

type CallExpr struct {
	Callee    Expr
	Paren     Token
	Arguments []Expr
}

func (b CallExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitCallExpr(b)
}

type GetExpr struct {
	Object Expr
	Name   Token
}

func (b GetExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitGetExpr(b)
}

type SetExpr struct {
	Object Expr
	Name   Token
	Value  Expr
}

func (b SetExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitSetExpr(b)
}

type ExprVisitor interface {
	VisitBinaryExpr(expr BinaryExpr) interface{}
	VisitGroupingExpr(expr GroupingExpr) interface{}
	VisitLiteralExpr(expr LiteralExpr) interface{}
	VisitUnaryExpr(expr UnaryExpr) interface{}
	VisitVariableExpr(expr VariableExpr) interface{}
	VisitAssignExpr(expr AssignExpr) interface{}
	VisitLogicalExpr(expr LogicalExpr) interface{}
	VisitCallExpr(expr CallExpr) interface{}
	VisitGetExpr(expr GetExpr) interface{}
	VisitSetExpr(expr SetExpr) interface{}
}
