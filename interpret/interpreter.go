package interpret

import (
	"fmt"
	"io"

	"github.com/Pra1tik/golox/ast"
	env "github.com/Pra1tik/golox/environment"
)

type Interpreter struct {
	environment *env.Environment
	globals     *env.Environment
	stdOut      io.Writer
	stdErr      io.Writer
	locals      map[ast.Expr]int
}

type runtimeError struct {
	token   ast.Token
	message string
}

type Return struct {
	Value interface{}
}

func (r runtimeError) Error() string {
	return fmt.Sprintf("%s\n[line %d]", r.message, r.token.Line)
}

func CreateInterpreter(stdOut io.Writer, stdErr io.Writer) *Interpreter {
	globals := env.CreateEnvironment(nil)
	globals.Define("clock", clock{})

	return &Interpreter{globals: globals, environment: globals, stdOut: stdOut, stdErr: stdErr, locals: make(map[ast.Expr]int)}
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

func (interp *Interpreter) VisitWhileStmt(stmt ast.WhileStmt) interface{} {
	for interp.isTruthy(interp.evaluate(stmt.Condition)) {
		interp.execute(stmt.Body)
	}
	return nil
}

func (interp *Interpreter) VisitFunctionStmt(stmt ast.FunctionStmt) interface{} {
	function := function{declaration: stmt, closure: interp.environment, isInitializer: false}
	interp.environment.Define(stmt.Name.Lexeme, function)
	return nil
}

func (interp *Interpreter) VisitClassStmt(stmt ast.ClassStmt) interface{} {
	var superclass *class
	if stmt.Superclass != nil {
		superclassVal, ok := interp.evaluate(stmt.Superclass).(class)
		if !ok {
			interp.error(stmt.Superclass.Name, "Superclass must be a class.")
		}
		superclass = &superclassVal
	}

	interp.environment.Define(stmt.Name.Lexeme, nil)

	if stmt.Superclass != nil {
		interp.environment = env.CreateEnvironment(interp.environment)
		interp.environment.Define("super", superclass)
	}

	methods := make(map[string]function, len(stmt.Methods))
	for _, method := range stmt.Methods {
		fn := function{
			declaration:   method,
			closure:       interp.environment,
			isInitializer: method.Name.Lexeme == "init",
		}
		methods[method.Name.Lexeme] = fn
	}

	class := class{
		name:       stmt.Name.Lexeme,
		methods:    methods,
		superclass: superclass,
	}

	if superclass != nil {
		interp.environment = interp.environment.Enclosing
	}

	interp.environment.Assign(stmt.Name.Lexeme, class)
	return nil
}

func (interp *Interpreter) VisitReturnStmt(stmt ast.ReturnStmt) interface{} {
	var value interface{}
	if stmt.Value != nil {
		value = interp.evaluate(stmt.Value)
	}
	panic(Return{Value: value})
}

func (interp *Interpreter) VisitAssignExpr(expr ast.AssignExpr) interface{} {
	value := interp.evaluate(expr.Value)
	// interp.environment.Assign(expr.Name.Lexeme, value)
	if distance, ok := interp.locals[expr]; ok {
		interp.environment.AssignAt(distance, expr.Name.Lexeme, value)
	} else {
		if err := interp.globals.Assign(expr.Name.Lexeme, value); err != nil {
			panic(err)
		}
	}
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

func (interp *Interpreter) VisitGetExpr(expr ast.GetExpr) interface{} {
	object := interp.evaluate(expr.Object)
	if instance, ok := object.(*instance); ok {
		val, err := instance.Get(interp, expr.Name)
		if err != nil {
			panic(err)
		}
		return val
	}
	interp.error(expr.Name, "Only instances have properties.")
	return nil
}

func (interp *Interpreter) VisitSetExpr(expr ast.SetExpr) interface{} {
	object := interp.evaluate(expr.Object)

	instance, ok := object.(*instance) //doesnt work if not pointer?
	if !ok {
		interp.error(expr.Name, "Only instances have fields")
	}

	value := interp.evaluate(expr.Value)
	instance.set(expr.Name, value)
	return nil
}

func (interp *Interpreter) VisitThisExpr(expr ast.ThisExpr) interface{} {
	val, err := interp.lookupVariable(expr.Keyword, expr)
	if err != nil {
		panic(err)
	}
	return val
}

func (interp *Interpreter) VisitSuperExpr(expr ast.SuperExpr) interface{} {
	distance := interp.locals[expr]
	superclass := interp.environment.GetAt(distance, "super").(*class)
	object := interp.environment.GetAt(distance-1, "this").(*instance)
	method := superclass.findMethod(expr.Method.Lexeme)
	if method == nil {
		interp.error(expr.Method, fmt.Sprintf("Undefined property '%s'.", expr.Method.Lexeme))
	}
	return method.bind(object)
}

func (interp *Interpreter) VisitVariableExpr(expr ast.VariableExpr) interface{} {
	// val, err := interp.environment.Get(expr.Name.Lexeme)
	val, err := interp.lookupVariable(expr.Name, expr)
	if err != nil {
		panic(err)
	}
	return val
}

func (interp *Interpreter) lookupVariable(name ast.Token, expr ast.Expr) (interface{}, error) {
	if distance, ok := interp.locals[expr]; ok {
		return interp.environment.GetAt(distance, name.Lexeme), nil
	}
	return interp.globals.Get(name.Lexeme)
}

func (interp *Interpreter) VisitCallExpr(expr ast.CallExpr) interface{} {
	callee := interp.evaluate(expr.Callee)

	args := make([]interface{}, len(expr.Arguments))
	for i, arg := range expr.Arguments {
		args[i] = interp.evaluate(arg)
	}

	fn, ok := callee.(callable)
	if !ok {
		interp.error(expr.Paren, "Can only call function and classes.")
	}

	if len(args) != fn.arity() {
		interp.error(expr.Paren, fmt.Sprintf("Expected %d arguments but got %d.", fn.arity(), len(args)))
	}

	return fn.call(interp, args)
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

func (interp *Interpreter) VisitLogicalExpr(expr ast.LogicalExpr) interface{} {
	left := interp.evaluate(expr.Left)

	if expr.Operator.TokenType == ast.TokenOr {
		if interp.isTruthy(left) {
			return left
		}
	} else {
		if !interp.isTruthy(left) {
			return left
		}
	}
	return interp.evaluate(expr.Right)
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

func (interp *Interpreter) Resolve(expr ast.Expr, depth int) {
	interp.locals[expr] = depth
}

func (interp *Interpreter) error(token ast.Token, message string) {
	panic(runtimeError{token: token, message: message})
}
