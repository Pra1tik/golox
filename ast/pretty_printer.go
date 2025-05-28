package ast

import (
	"strings"
)

type AstPrinter struct{}

func (p AstPrinter) Print(expr Expr) string {
	return expr.Accept(p).(string)
}

func (p AstPrinter) VisitBinaryExpr(expr BinaryExpr) interface{} {
	return p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (p AstPrinter) VisitGroupingExpr(expr GroupingExpr) interface{} {
	return p.parenthesize("group", expr.Expression)
}

func (p AstPrinter) VisitLiteralExpr(expr LiteralExpr) interface{} {
	if expr.Value == nil {
		return "nil"
	}
	return expr.Value
}

func (p AstPrinter) VisitUnaryExpr(expr UnaryExpr) interface{} {
	return p.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (p AstPrinter) parenthesize(name string, exprs ...Expr) string {
	var builder strings.Builder

	builder.WriteString("(")
	builder.WriteString(name)

	for _, expr := range exprs {
		builder.WriteString(" ")
		result := expr.Accept(p)
		strResult, ok := result.(string)
		if !ok {
			strResult = "<error>"
		}
		builder.WriteString(strResult)
	}

	builder.WriteString(")")

	return builder.String()
}
