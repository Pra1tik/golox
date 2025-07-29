package interpret

import (
	"github.com/Pra1tik/golox/ast"
	env "github.com/Pra1tik/golox/environment"
)

type callable interface {
	arity() int
	call(interp *Interpreter, args []interface{}) interface{}
}

type function struct {
	declaration   ast.FunctionStmt
	closure       *env.Environment
	isInitializer bool
}

func (f function) arity() int {
	return len(f.declaration.Params)
}

func (f function) call(interp *Interpreter, args []interface{}) (returnVal interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(Return); ok {
				if f.isInitializer {
					returnVal = f.closure.GetAt(0, "this")
					return
				}

				returnVal = v.Value
				return
			}
			panic(err)
		}
	}()

	environment := env.CreateEnvironment(f.closure)
	for index, arg := range f.declaration.Params {
		environment.Define(arg.Lexeme, args[index])
	}

	interp.executeBlock(f.declaration.Body, environment)

	if f.isInitializer {
		return f.closure.GetAt(0, "this")
	}

	return nil
}

func (f function) bind(i *instance) function {
	environment := env.CreateEnvironment(f.closure)
	environment.Define("this", i)
	return function{
		declaration:   f.declaration,
		closure:       environment,
		isInitializer: f.isInitializer,
	}
}

func (f function) String() string {
	return "<fn " + f.declaration.Name.Lexeme + ">"
}
