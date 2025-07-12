package parser

import (
	"fmt"
	"io"

	"github.com/Pra1tik/golox/ast"
)

// program → declaration* EOF ;
// declaration → varDecl | statement | funDecl ;
// funDecl → "fun" function ;
// function → IDENTIFIER "(" parameters? ")" block ;
// parameters → IDENTIFIER ( "," IDENTIFIER )* ;
// statement → exprStmt | printStmt | block | ifStmt
// 			 | whileStmt | forStmt | returnStmt ;
// block → "{" declaration* "}" ;
// varDecl → "var" IDENTIFIER ( "=" expression )? ";" ;
// exprStmt → expression ";" ;
// printStmt → "print" expression ";" ;
// whileStmt → "while" "(" expression ")" statement ;
// forStmt → "for" "(" ( varDecl | exprStmt | ";" )
//			expression? ";"
//			expression? ")" statement ;
// ifStmt → "if" "(" expression ")" statement
//          ( "else" statement )? ;
// returnStmt → "return" expression? ";" ;
// expression → assignment ;
// assignment → IDENTIFIER "=" assignment | logic_or ;
// logic_or → logic_and ( "or" logic_and )* ;
// logic_and → equality ( "and" equality )* ;
// equality → comparison ( ( "!=" | "==" ) comparison )* ;
// comparison → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
// term → factor ( ( "-" | "+" ) factor )* ;
// factor → unary ( ( "/" | "*" ) unary )* ;
// unary → ( "!" | "-" ) unary | call ;
// call → primary ( "(" arguments? ")" )* ;
// arguments → expression ( "," expression )* ;
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
	if p.match(ast.TokenFun) {
		return p.function("function")
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

func (p *Parser) function(kind string) ast.FunctionStmt {
	name := p.consume(ast.TokenIdentifier, "Expect "+kind+" name.")

	p.consume(ast.TokenLeftParen, "Expect '(' after "+kind+" name.")
	var parameters []ast.Token
	if !p.check(ast.TokenRightParen) {
		for {
			if len(parameters) >= 255 {
				p.error(p.peek(), "Can't have more than 255 paramters.")
			}

			arg := p.consume(ast.TokenIdentifier, "Expect parameter name.")
			parameters = append(parameters, arg)
			if !p.match(ast.TokenComma) {
				break
			}
		}
	}
	p.consume(ast.TokenRightParen, "Expect ')' after parameters.")

	p.consume(ast.TokenLeftBrace, "Expect '{' before "+kind+" body.")
	body := p.block()

	return ast.FunctionStmt{Name: name, Params: parameters, Body: body}
}

func (p *Parser) statement() ast.Stmt {
	if p.match(ast.TokenPrint) {
		return p.printStatement()
	}
	if p.match(ast.TokenLeftBrace) {
		stmt := p.block()
		return ast.BlockStmt{Statements: stmt}
	}
	if p.match(ast.TokenIf) {
		return p.ifStatement()
	}
	if p.match(ast.TokenWhile) {
		return p.whileStatement()
	}
	if p.match(ast.TokenFor) {
		return p.forStatement()
	}
	if p.match(ast.TokenReturn) {
		return p.returnStatement()
	}
	return p.expressionStatement()
}

func (p *Parser) printStatement() ast.Stmt {
	expr := p.expression()
	p.consume(ast.TokenSemicolon, "Expected token ';' after value")
	return ast.PrintStmt{Expr: expr}
}

func (p *Parser) ifStatement() ast.Stmt {
	p.consume(ast.TokenLeftParen, "Expect '(' after 'if'.")
	condition := p.expression()
	p.consume(ast.TokenRightParen, "Expect ')' after if condition.")

	thenBranch := p.statement()
	var elseBranch ast.Stmt
	if p.match(ast.TokenElse) {
		elseBranch = p.statement()
	}

	return ast.IfStmt{Condition: condition, ThenBranch: thenBranch, ElseBranch: elseBranch}
}

func (p *Parser) whileStatement() ast.Stmt {
	p.consume(ast.TokenLeftParen, "Expect '(' after 'while'.")
	condition := p.expression()
	p.consume(ast.TokenRightParen, "Expect ')' after condition.")
	body := p.statement()

	return ast.WhileStmt{Condition: condition, Body: body}
}

func (p *Parser) forStatement() ast.Stmt {
	p.consume(ast.TokenLeftParen, "Expect '(' after 'for'.")

	var initializer ast.Stmt
	if p.match(ast.TokenSemicolon) {
		initializer = nil
	} else if p.match(ast.TokenVar) {
		initializer = p.varDeclaration()
	} else {
		initializer = p.expressionStatement()
	}

	var condition ast.Expr
	if !p.check(ast.TokenSemicolon) {
		condition = p.expression()
	}
	p.consume(ast.TokenSemicolon, "Expect ';' after loop condition.")

	var increment ast.Expr
	if !p.check(ast.TokenRightParen) {
		increment = p.expression()
	}
	p.consume(ast.TokenRightParen, "Expect ')' after for clauses.")
	body := p.statement()

	if increment != nil {
		body = ast.BlockStmt{Statements: []ast.Stmt{body, ast.ExpressionStmt{Expr: increment}}}
	}

	if condition == nil {
		condition = ast.LiteralExpr{Value: true}
	}
	body = ast.WhileStmt{Body: body, Condition: condition}

	if initializer != nil {
		body = ast.BlockStmt{Statements: []ast.Stmt{initializer, body}}
	}

	return body
}

func (p *Parser) returnStatement() ast.Stmt {
	keyword := p.previous()
	var value ast.Expr
	if !p.check(ast.TokenSemicolon) {
		value = p.expression()
	}

	p.consume(ast.TokenSemicolon, "Expect ';' after return value.")
	return ast.ReturnStmt{Keyword: keyword, Value: value}
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
	expr := p.or()

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

func (p *Parser) or() ast.Expr {
	expr := p.and()

	for p.match(ast.TokenOr) {
		operator := p.previous()
		right := p.and()
		expr = ast.LogicalExpr{Left: expr, Operator: operator, Right: right}
	}
	return expr
}

func (p *Parser) and() ast.Expr {
	expr := p.equality()

	for p.match(ast.TokenAnd) {
		operator := p.previous()
		right := p.and()
		expr = ast.LogicalExpr{Left: expr, Operator: operator, Right: right}
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

	return p.call()
}

func (p *Parser) call() ast.Expr {
	expr := p.primary()

	for {
		if p.match(ast.TokenLeftParen) {
			expr = p.finishCall(expr)
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) finishCall(callee ast.Expr) ast.Expr {
	args := make([]ast.Expr, 0)
	if !p.check(ast.TokenRightParen) {
		for {
			if len(args) >= 255 {
				p.error(p.peek(), "Can't have more than 255 arguments.")
			}
			expr := p.expression()
			args = append(args, expr)
			if !p.match(ast.TokenComma) {
				break
			}
		}
	}
	paren := p.consume(ast.TokenRightParen, "Expect ')' after arguments.")
	return ast.CallExpr{Callee: callee, Paren: paren, Arguments: args}
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
