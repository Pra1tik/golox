package interpret

import (
	"fmt"

	"github.com/Pra1tik/golox/ast"
)

type class struct {
	name       string
	methods    map[string]function
	superclass *class
}

func (c class) arity() int {
	initializer := c.findMethod("init")
	if initializer == nil {
		return 0
	}
	return initializer.arity()
}

func (c class) call(interpreter *Interpreter, arguments []interface{}) interface{} {
	in := &instance{class: c}
	initializer := c.findMethod("init")

	if initializer != nil {
		initializer.bind(in).call(interpreter, arguments)
	}

	return in
}

func (c class) findMethod(name string) *function {
	if method, ok := c.methods[name]; ok {
		return &method
	}

	if c.superclass != nil {
		return c.superclass.findMethod(name)
	}

	return nil
}

func (c class) String() string {
	return c.name
}

type instance struct {
	class  class
	fields map[string]interface{}
}

func (i *instance) Get(interpreter *Interpreter, name ast.Token) (interface{}, error) {
	if val, ok := i.fields[name.Lexeme]; ok { // field take precendence over method
		return val, nil
	}

	method := i.class.findMethod(name.Lexeme)
	if method != nil {
		return method.bind(i), nil
	}

	return nil, runtimeError{token: name, message: fmt.Sprintf("Undefined property '%s'.'", name.Lexeme)}
}

func (i *instance) set(name ast.Token, value interface{}) {
	if i.fields == nil {
		i.fields = make(map[string]interface{})
	}
	i.fields[name.Lexeme] = value
}

func (i instance) String() string {
	return i.class.name + " instance"
}
