package parser

import (
	"fmt"
	"io"

	"github.com/Pra1tik/golox/ast"
)

// program → declaration* EOF ;
// declaration → varDecl | statement;
// statement → exprStmt | printStmt | block ;
// block → "{" declaration* "}" ;
// varDecl → "var" IDENTIFIER ( "=" expression )? ";" ;
// exprStmt → expression ";" ;
// printStmt → "print" expression ";" ;
// expression → assignment ;
// assignment -> IDENTIFIER "=" assignment | equality ;
// equality → comparison ( ( "!=" | "==" ) comparison )* ;
// comparison → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
// term → factor ( ( "-" | "+" ) factor )* ;
// factor → unary ( ( "/" | "*" ) unary )* ;
// unary → ( "!" | "-" ) unary | primary ;
// primary → NUMBER | STRING | "true" | "false" | "nil"
// 		|  "(" expression ")" | IDENTIFIER;

type Parser struct {
	tokens   []ast.Token
	current  int
	stdErr   io.Writer
	hadError bool
}

func CreateParser(tokens []ast.Token, stdErr io.Writer) *Parser {
	return &Parser{tokens: tokens, stdErr: stdErr}
}

func (p *Parser) Parse() ([]ast.Stmt, bool) {
	var statements []ast.Stmt
	for !p.isAtEnd() {
		stmt := p.declaration()
		statements = append(statements, stmt)
	}
	return statements, p.hadError
}

func (p *Parser) declaration() ast.Stmt {

	if p.match(ast.TokenVar) {
		return p.varDeclaration()
	}
	return p.statement()
}

func (p *Parser) varDeclaration() ast.Stmt {
	var_name := p.consume(ast.TokenIdentifier, "Expected variable name")

	var initializer ast.Expr
	if p.match(ast.TokenEqual) {
		initializer = p.expression()
	}
	p.consume(ast.TokenSemicolon, "Expected token ';' after value")
	return ast.VarStmt{Name: var_name, Initializer: initializer}
}

func (p *Parser) statement() ast.Stmt {
	if p.match(ast.TokenPrint) {
		return p.printStatement()
	}
	if p.match(ast.TokenLeftBrace) {
		stmt := p.block()
		return ast.BlockStmt{Statements: stmt}
	}
	return p.expressionStatement()
}

func (p *Parser) printStatement() ast.Stmt {
	expr := p.expression()
	p.consume(ast.TokenSemicolon, "Expected token ';' after value")
	return ast.PrintStmt{Expr: expr}
}

func (p *Parser) expressionStatement() ast.Stmt {
	expr := p.expression()
	p.consume(ast.TokenSemicolon, "Expected token ';' after value")
	return ast.ExpressionStmt{Expr: expr}
}

func (p *Parser) expression() ast.Expr {
	return p.assignment()
}

func (p *Parser) block() []ast.Stmt {
	var statements []ast.Stmt
	for !p.check(ast.TokenRightBrace) && !p.isAtEnd() {
		stmt := p.declaration()
		statements = append(statements, stmt)
	}
	p.consume(ast.TokenRightBrace, "Expected '}' after block")
	return statements
}

func (p *Parser) assignment() ast.Expr {
	expr := p.equality()

	if p.match(ast.TokenEqual) {
		equals := p.previous()
		value := p.assignment()

		if varExpr, ok := expr.(ast.VariableExpr); ok {
			return ast.AssignExpr{Name: varExpr.Name, Value: value}
		}

		p.error(equals, "Invalid assignment target.")
	}

	return expr
}

func (p *Parser) equality() ast.Expr {
	expr := p.comparison()

	for p.match(ast.TokenBangEqual, ast.TokenEqualEqual) {
		operator := p.previous()
		right := p.comparison()
		expr = ast.BinaryExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) comparison() ast.Expr {
	expr := p.term()

	for p.match(ast.TokenGreater, ast.TokenGreaterEqual, ast.TokenLess, ast.TokenLessEqual) {
		operator := p.previous()
		right := p.term()
		expr = ast.BinaryExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) term() ast.Expr {
	expr := p.factor()

	for p.match(ast.TokenMinus, ast.TokenPlus) {
		operator := p.previous()
		right := p.factor()
		expr = ast.BinaryExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) factor() ast.Expr {
	expr := p.unary()

	for p.match(ast.TokenSlash, ast.TokenStar) {
		operator := p.previous()
		right := p.unary()
		expr = ast.BinaryExpr{Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func (p *Parser) unary() ast.Expr {
	if p.match(ast.TokenBang, ast.TokenMinus) {
		operator := p.previous()
		right := p.unary()
		return ast.UnaryExpr{Operator: operator, Right: right}
	}

	return p.primary()
}

func (p *Parser) primary() ast.Expr {
	switch {
	case p.match(ast.TokenFalse):
		return ast.LiteralExpr{Value: false}
	case p.match(ast.TokenTrue):
		return ast.LiteralExpr{Value: true}
	case p.match(ast.TokenNil):
		return ast.LiteralExpr{Value: nil}
	case p.match(ast.TokenNumber, ast.TokenString):
		return ast.LiteralExpr{Value: p.previous().Literal}
	case p.match(ast.TokenIdentifier):
		return ast.VariableExpr{Name: p.previous()}
	case p.match(ast.TokenLeftParen):
		expr := p.expression()
		p.consume(ast.TokenRightParen, "Expected ) after expression.")
		return ast.GroupingExpr{Expression: expr}
	}

	p.error(p.peek(), "Expected expression.")
	return nil
}

func (p *Parser) consume(tokenType ast.TokenType, message string) ast.Token {
	if p.check(tokenType) {
		return p.advance()
	}

	p.error(p.peek(), message)
	return ast.Token{}
}

func (p *Parser) error(token ast.Token, message string) {
	var where string
	if token.TokenType == ast.TokenEof {
		where = " at end"
	} else {
		where = " at '" + token.Lexeme + "'"
	}

	err := fmt.Sprintf("[line %d] Error%s: %s\n", token.Line+1, where, message)
	_, _ = p.stdErr.Write([]byte(err))
	panic(err)
}

func (p *Parser) match(types ...ast.TokenType) bool {
	for _, tokenType := range types {
		if p.check(tokenType) {
			p.advance()
			return true
		}
	}

	return false
}

func (p *Parser) check(tokenType ast.TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().TokenType == tokenType
}

func (p *Parser) advance() ast.Token {
	if !p.isAtEnd() {
		p.current++
	}

	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().TokenType == ast.TokenEof
}

func (p *Parser) peek() ast.Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() ast.Token {
	return p.tokens[p.current-1]
}

func (p *Parser) peekNext() ast.Token {
	return p.tokens[p.current+1]
}
