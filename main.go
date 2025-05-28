package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/Pra1tik/golox/ast"
	"github.com/Pra1tik/golox/lexer"
)

var (
	hadError bool
	stdErr   io.Writer
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
}

func runPrompt() {
	inputScanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !inputScanner.Scan() {
			break
		}

		line := inputScanner.Text()
		run(line)
		hadError = false // mistake shouldn't kill the entire session
	}
}

func run(source string) {
	stdErr = os.Stderr
	lexer := lexer.CreateScanner(source, stdErr)
	tokens := lexer.ScanTokens()

	// print tokens
	for _, token := range tokens {
		fmt.Println(token.Lexeme)
	}

	// pretty printer
	expression := ast.BinaryExpr{
		Left: ast.LiteralExpr{Value: "123"},
		Operator: ast.Token{
			TokenType: ast.TokenMinus,
			Lexeme:    "-",
			Literal:   nil,
			Line:      1,
			Start:     0,
		},
		Right: ast.LiteralExpr{Value: "456"},
	}

	printer := ast.AstPrinter{}
	fmt.Println("Print: ", printer.Print(expression))
}

func errorFunc(line int, message string) {
	report(line, "", message)
}

func report(line int, where string, message string) {
	fmt.Errorf("[line %d] Error %s : %s", line, where, message)
}
