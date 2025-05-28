package lexer

import (
	"fmt"
	"io"
	"strconv"

	"github.com/Pra1tik/golox/ast"
)

type Scanner struct {
	start   int
	current int
	line    int
	source  string
	tokens  []ast.Token
	stdErr  io.Writer
}

func CreateScanner(source string, stdErr io.Writer) *Scanner {
	return &Scanner{source: source, stdErr: stdErr}
}

func (s *Scanner) ScanTokens() []ast.Token {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, ast.Token{TokenType: ast.TokenEof, Line: s.line})
	return s.tokens
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) scanToken() {
	c := s.advance()
	switch c {
	case '(':
		s.addToken(ast.TokenLeftParen)
	case ')':
		s.addToken(ast.TokenRightParen)
	case '{':
		s.addToken(ast.TokenLeftBrace)
	case '}':
		s.addToken(ast.TokenRightBrace)
	case ',':
		s.addToken(ast.TokenComma)
	case '.':
		s.addToken(ast.TokenDot)
	case '-':
		s.addToken(ast.TokenMinus)
	case '+':
		s.addToken(ast.TokenPlus)
	case ';':
		s.addToken(ast.TokenSemicolon)
	case '*':
		s.addToken(ast.TokenStar)

	case '!':
		var tokenType ast.TokenType
		if s.match('=') {
			tokenType = ast.TokenBangEqual
		} else {
			tokenType = ast.TokenBang
		}
		s.addToken(tokenType)

	case '=':
		var tokenType ast.TokenType
		if s.match('=') {
			tokenType = ast.TokenEqualEqual
		} else {
			tokenType = ast.TokenEqual
		}
		s.addToken(tokenType)

	case '<':
		var tokenType ast.TokenType
		if s.match('=') {
			tokenType = ast.TokenLessEqual
		} else {
			tokenType = ast.TokenLess
		}
		s.addToken(tokenType)

	case '>':
		var tokenType ast.TokenType
		if s.match('=') {
			tokenType = ast.TokenGreaterEqual
		} else {
			tokenType = ast.TokenGreater
		}
		s.addToken(tokenType)

	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else if s.match('*') {
			for !s.isAtEnd() && (s.peek() != '*' && s.peekNext() != '/') {
				s.advance()
			}

			if s.peek() != '*' || s.peekNext() != '/' {
				s.error("Multiline comment not terminated")
				break
			}

			s.advance()
			s.advance()
		} else {
			s.addToken(ast.TokenSlash)
		}

	case ' ':
	case '\r':
	case '\t':
		break
	case '\n':
		s.line++

	case '"':
		s.string()

	default:
		if isDigit(c) {
			s.number()
		} else if isAlpha(c) {
			s.identifier()
		} else {
			s.error("Unexpected character.")
		}
	}
}

func (s *Scanner) advance() rune {
	ch := rune(s.source[s.current])
	s.current++
	return ch
}

func (s *Scanner) addToken(tokenType ast.TokenType) {
	s.addTokenWithLiteral(tokenType, nil)
}

func (s *Scanner) addTokenWithLiteral(tokenType ast.TokenType, literal interface{}) {
	text := s.source[s.start:s.current]
	token := ast.Token{
		TokenType: tokenType,
		Lexeme:    text,
		Literal:   literal,
		Line:      s.line,
		Start:     s.start,
	}

	s.tokens = append(s.tokens, token)
}

func (s *Scanner) string() {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		s.error("Unterminated string.")
		return
	}

	s.advance()
	value := s.source[s.start+1 : s.current-1]
	s.addTokenWithLiteral(ast.TokenString, value)
}

func (s *Scanner) number() {
	for isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && isDigit((s.peekNext())) {
		s.advance()
		for isDigit(s.peek()) {
			s.advance()
		}
	}

	value, _ := strconv.ParseFloat(s.source[s.start:s.current], 64)
	s.addTokenWithLiteral(ast.TokenNumber, value)
}

func (s *Scanner) identifier() {
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}

	text := s.source[s.start:s.current]
	tokenType, exist := keywords[text]
	if !exist {
		tokenType = ast.TokenIdentifier
	}

	s.addToken(tokenType)
}

func (s *Scanner) match(expected rune) bool {
	if s.isAtEnd() {
		return false
	}

	if rune(s.source[s.current]) != expected {
		return false
	}

	s.current++
	return true
}

func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return '\000'
	}

	return rune(s.source[s.current])
}

func (s *Scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return '\000'
	}

	return rune(s.source[s.current+1])
}

func isDigit(char rune) bool {
	return char >= '0' && char <= '9'
}

func isAlpha(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		char == '_'
}

func isAlphaNumeric(char rune) bool {
	return isAlpha(char) || isDigit(char)
}

var keywords = map[string]ast.TokenType{
	"and":    ast.TokenAnd,
	"class":  ast.TokenClass,
	"else":   ast.TokenElse,
	"false":  ast.TokenFalse,
	"for":    ast.TokenFor,
	"fun":    ast.TokenFun,
	"if":     ast.TokenIf,
	"nil":    ast.TokenNil,
	"or":     ast.TokenOr,
	"print":  ast.TokenPrint,
	"return": ast.TokenReturn,
	"super":  ast.TokenSuper,
	"this":   ast.TokenThis,
	"true":   ast.TokenTrue,
	"var":    ast.TokenVar,
	"while":  ast.TokenWhile,

	// "break":    ast.TokenBreak,
	// "continue": ast.TokenContinue,
	// "type":     ast.TokenTypeType,
}

func (s *Scanner) error(msg string) {
	_, _ = s.stdErr.Write([]byte(fmt.Sprintf("[line %d] Error: %s\n", s.line, msg)))
}
