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
	declaration ast.FunctionStmt
}

func (f function) arity() int {
	return len(f.declaration.Params)
}

func (f function) call(interp *Interpreter, args []interface{}) (returnVal interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(Return); ok {
				returnVal = v.Value
				return
			}
			panic(err)
		}
	}()

	environment := env.CreateEnvironment(interp.environment)
	for index, arg := range f.declaration.Params {
		environment.Define(arg.Lexeme, args[index])
	}

	interp.executeBlock(f.declaration.Body, environment)
	return nil
}

func (f function) String() string {
	return "<fn " + f.declaration.Name.Lexeme + ">"
}
