package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/Pra1tik/golox/ast"
	"github.com/Pra1tik/golox/interpret"
	"github.com/Pra1tik/golox/lexer"
	"github.com/Pra1tik/golox/parser"
)

var (
	hadError        bool
	hadRuntimeError bool
	stdErr          io.Writer
	stdOut          io.Writer
)

func main() {
	args := os.Args
	if len(args) > 2 {
		fmt.Println("Usage: golox [script]")
	} else if len(args) == 2 {
		fmt.Println("Run script from file")
		runFile(args[1])
	} else {
		fmt.Println("Interactive mode")
		runPrompt()
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func runFile(path string) {
	source, err := os.ReadFile(path)
	checkError(err)

	run(string(source))
	if hadError {
		os.Exit(65)
	}
	if hadRuntimeError {
		os.Exit(70)
	}
}

func runPrompt() {
	inputScanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !inputScanner.Scan() {
			break
		}

		line := inputScanner.Text()
		fmt.Println(run(line))
		hadError = false // mistake shouldn't kill the entire session
	}
}

func run(source string) interface{} {
	stdErr = os.Stderr
	stdOut = os.Stdout
	lexer := lexer.CreateScanner(source, stdErr)
	tokens := lexer.ScanTokens()

	// print tokens
	// for _, token := range tokens {
	// 	fmt.Println(token.Lexeme)
	// }

	parser := parser.CreateParser(tokens, stdErr)
	var statements []ast.Stmt
	statements, hadError = parser.Parse()

	if hadError {
		return nil
	}

	interpreter := interpret.CreateInterpreter(stdOut, stdErr)
	var result interface{}
	result, hadRuntimeError = interpreter.Interpret(statements)
	return result
}

func errorFunc(line int, message string) {
	report(line, "", message)
}

func report(line int, where string, message string) {
	fmt.Errorf("[line %d] Error %s : %s", line, where, message)
}
