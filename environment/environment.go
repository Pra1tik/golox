package env

import "errors"

var ErrUndefined = errors.New("undefined variable")

type Environment struct {
	Enclosing *Environment
	values    map[string]interface{}
}

func CreateEnvironment(enclosing *Environment) *Environment {
	return &Environment{Enclosing: enclosing, values: make(map[string]interface{})}
}

func (e *Environment) Define(name string, value interface{}) {
	e.values[name] = value
}

func (e *Environment) Get(name string) (interface{}, error) {
	if val, ok := e.values[name]; ok {
		return val, nil
	}
	if e.Enclosing != nil {
		return e.Enclosing.Get(name)
	}
	return nil, ErrUndefined
}

func (e *Environment) Assign(name string, value interface{}) error {
	if _, ok := e.values[name]; ok {
		e.Define(name, value)
		return nil
	}
	if e.Enclosing != nil {
		return e.Enclosing.Assign(name, value)
	}
	return ErrUndefined
}
