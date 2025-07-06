package interpret

import (
	"fmt"
	"io"

	"github.com/Pra1tik/golox/ast"
	env "github.com/Pra1tik/golox/environment"
)

type Interpreter struct {
	environment *env.Environment
	stdOut      io.Writer
	stdErr      io.Writer
}

type runtimeError struct {
	token   ast.Token
	message string
}

func (r runtimeError) Error() string {
	return fmt.Sprintf("%s\n[line %d]", r.message, r.token.Line)
}

func CreateInterpreter(stdOut io.Writer, stdErr io.Writer) *Interpreter {
	globals := env.CreateEnvironment(nil)

	return &Interpreter{environment: globals, stdOut: stdOut, stdErr: stdErr}
}

func (interp *Interpreter) Interpret(stmts []ast.Stmt) (result interface{}, hadRuntimeError bool) {
	defer func() {
		if err := recover(); err != nil {
			if e, ok := err.(runtimeError); ok {
				_, _ = interp.stdErr.Write([]byte(e.Error() + "\n"))
				hadRuntimeError = true
			} else {
				fmt.Printf("Error: %s\n", err)
			}
		}
	}()

	for _, statement := range stmts {
		result = interp.execute(statement)
	}
	return
}

func (interp *Interpreter) execute(stmt ast.Stmt) interface{} {
	return stmt.Accept(interp)
}

func (interp *Interpreter) evaluate(expr ast.Expr) interface{} {
	return expr.Accept(interp)
}

func (interp *Interpreter) VisitVarStmt(stmt ast.VarStmt) interface{} {
	var val interface{}
	if stmt.Initializer != nil {
		val = interp.evaluate(stmt.Initializer)
	}
	interp.environment.Define(stmt.Name.Lexeme, val)
	return nil
}

func (interp *Interpreter) VisitPrintStmt(stmt ast.PrintStmt) interface{} {
	value := interp.evaluate(stmt.Expr)
	_, _ = interp.stdOut.Write([]byte(interp.stringify(value) + "\n"))
	return nil
}

func (interp *Interpreter) VisitBlockStmt(stmt ast.BlockStmt) interface{} {
	interp.executeBlock(stmt.Statements, env.CreateEnvironment(interp.environment))
	return nil
}

func (interp *Interpreter) VisitIfStmt(stmt ast.IfStmt) interface{} {
	if interp.isTruthy(interp.evaluate(stmt.Condition)) {
		interp.execute(stmt.ThenBranch)
	} else if stmt.ElseBranch != nil {
		interp.execute(stmt.ElseBranch)
	}
	return nil
}

func (interp *Interpreter) VisitAssignExpr(expr ast.AssignExpr) interface{} {
	value := interp.evaluate(expr.Value)
	interp.environment.Assign(expr.Name.Lexeme, value)
	return value
}

func (interp *Interpreter) VisitExpressionStmt(stmt ast.ExpressionStmt) interface{} {
	return interp.evaluate(stmt.Expr)
}

func (interp *Interpreter) VisitLiteralExpr(expr ast.LiteralExpr) interface{} {
	return expr.Value
}

func (interp *Interpreter) VisitGroupingExpr(expr ast.GroupingExpr) interface{} {
	return interp.evaluate(expr.Expression)
}

func (interp *Interpreter) VisitUnaryExpr(expr ast.UnaryExpr) interface{} {
	right := interp.evaluate(expr.Right)

	switch expr.Operator.TokenType {
	case ast.TokenMinus:
		interp.checkOperands(expr.Operator, right)
		return -right.(float64)
	case ast.TokenBang:
		return !interp.isTruthy(right)
	}
	return nil
}

func (interp *Interpreter) VisitVariableExpr(expr ast.VariableExpr) interface{} {
	val, err := interp.environment.Get(expr.Name.Lexeme)
	if err != nil {
		panic(err)
	}
	return val
}

func (interp *Interpreter) VisitBinaryExpr(expr ast.BinaryExpr) interface{} {
	left := interp.evaluate(expr.Left)
	right := interp.evaluate(expr.Right)

	switch expr.Operator.TokenType {
	// arithmetic
	case ast.TokenPlus:
		_, isLeftFloat := left.(float64)
		_, isRightFloat := right.(float64)
		if isLeftFloat && isRightFloat {
			return left.(float64) + right.(float64)
		}
		_, isLeftString := left.(string)
		_, isRightString := right.(string)
		if isLeftString && isRightString {
			return left.(string) + right.(string)
		}
		interp.error(expr.Operator, "Operands must be numbers or strings")
	case ast.TokenMinus:
		interp.checkOperands(expr.Operator, left, right)
		return left.(float64) - right.(float64)
	case ast.TokenStar:
		interp.checkOperands(expr.Operator, left, right)
		return left.(float64) * right.(float64)
	case ast.TokenSlash:
		interp.checkOperands(expr.Operator, left, right)
		return left.(float64) / right.(float64)

	// logical
	case ast.TokenGreater:
		interp.checkOperands(expr.Operator, left, right)
		return left.(float64) > right.(float64)
	case ast.TokenGreaterEqual:
		interp.checkOperands(expr.Operator, left, right)
		return left.(float64) >= right.(float64)
	case ast.TokenLess:
		interp.checkOperands(expr.Operator, left, right)
		return left.(float64) < right.(float64)
	case ast.TokenLessEqual:
		interp.checkOperands(expr.Operator, left, right)
		return left.(float64) <= right.(float64)
	case ast.TokenEqualEqual:
		return left == right
	case ast.TokenBangEqual:
		return left != right
	}

	return nil
}

func (interp *Interpreter) executeBlock(statements []ast.Stmt, env *env.Environment) {
	previous := interp.environment
	defer func() {
		interp.environment = previous
	}()

	interp.environment = env
	for _, statement := range statements {
		interp.execute(statement)
	}
}

func (interp *Interpreter) checkOperands(operator ast.Token, operands ...interface{}) {
	for _, operand := range operands {
		if _, ok := operand.(float64); !ok {
			panic(runtimeError{token: operator, message: "Operand must be number"})
		}
	}
}

func (interp *Interpreter) isTruthy(val interface{}) bool {
	if val == nil {
		return false
	}
	if v, ok := val.(bool); ok {
		return v
	}
	return true
}

func (interp *Interpreter) stringify(value interface{}) string {
	if value == nil {
		return "nil"
	}
	return fmt.Sprint(value)
}

func (interp *Interpreter) error(token ast.Token, message string) {
	panic(runtimeError{token: token, message: message})
}
