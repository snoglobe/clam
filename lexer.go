package main

import (
	"strconv"
	"strings"
)

type TokenType uint8

//go:generate stringer -type=TokenType
const (
	Id TokenType = iota
	Number
	String
	True
	False
	Nil

	// Keywords
	My
	Sub
	When
	Case
	If
	Unless
	Else
	While
	For
	In
	Until
	Do
	Return
	Inc
	Dec
	By

	// Operators
	Plus
	Minus
	Multiply
	Divide
	Modulo
	And
	Or
	Not
	Equal
	NotEqual
	Less
	LessEqual
	Greater
	GreaterEqual
	Assign

	// Punctuation
	Comma
	Colon
	Semicolon
	LeftParen
	RightParen
	LeftBrace
	RightBrace
	LeftBracket
	RightBracket
	Dot

	// Special
	Eof
)

var keywords = map[string]TokenType{
	"my":     My,
	"sub":    Sub,
	"when":   When,
	"case":   Case,
	"if":     If,
	"unless": Unless,
	"else":   Else,
	"while":  While,
	"for":    For,
	"in":     In,
	"until":  Until,
	"do":     Do,
	"return": Return,
	"true":   True,
	"false":  False,
	"nil":    Nil,
	"inc":    Inc,
	"dec":    Dec,
	"by":     By,
}

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

func (t Token) String() string {
	return t.Type.String() + " '" + t.Literal + "' (" + strconv.Itoa(t.Line) + ":" + strconv.Itoa(t.Column) + ")"
}

type Lexer struct {
	input    string
	position int

	line   int
	column int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

func (l *Lexer) peekChar() byte {
	if l.position >= len(l.input) {
		return 0
	}
	return l.input[l.position]
}

func (l *Lexer) readChar() byte {
	var ch = l.peekChar()
	l.position++
	l.column++
	return ch
}

func (l *Lexer) matchChar(ch byte) bool {
	if l.peekChar() == ch {
		l.readChar()
		return true
	}
	return false
}

func (l *Lexer) NextToken() Token {
	var line = l.line
	var column = l.column
	var ch = l.readChar()
	switch ch {
	case '#':
		for l.peekChar() != '\n' {
			notAtEnd(l.readChar())
		}
		l.line++
		l.column = 0
		return l.NextToken()
	case '+':
		return Token{Type: Plus, Literal: "+", Line: line, Column: column}
	case '-':
		return Token{Type: Minus, Literal: "-", Line: line, Column: column}
	case '*':
		return Token{Type: Multiply, Literal: "*", Line: line, Column: column}
	case '/':
		return Token{Type: Divide, Literal: "/", Line: line, Column: column}
	case '%':
		return Token{Type: Modulo, Literal: "%", Line: line, Column: column}
	case '(':
		return Token{Type: LeftParen, Literal: "(", Line: line, Column: column}
	case ')':
		return Token{Type: RightParen, Literal: ")", Line: line, Column: column}
	case '{':
		return Token{Type: LeftBrace, Literal: "{", Line: line, Column: column}
	case '}':
		return Token{Type: RightBrace, Literal: "}", Line: line, Column: column}
	case '[':
		return Token{Type: LeftBracket, Literal: "[", Line: line, Column: column}
	case ']':
		return Token{Type: RightBracket, Literal: "]", Line: line, Column: column}
	case ',':
		return Token{Type: Comma, Literal: ",", Line: line, Column: column}
	case ':':
		return Token{Type: Colon, Literal: ":", Line: line, Column: column}
	case ';':
		return Token{Type: Semicolon, Literal: ";", Line: line, Column: column}
	case '.':
		return Token{Type: Dot, Literal: ".", Line: line, Column: column}
	case '=':
		if l.matchChar('=') {
			return Token{Type: Equal, Literal: "==", Line: line, Column: column}
		}
		return Token{Type: Assign, Literal: "=", Line: line, Column: column}
	case '<':
		if l.matchChar('=') {
			return Token{Type: LessEqual, Literal: "<=", Line: line, Column: column}
		}
		return Token{Type: Less, Literal: "<", Line: line, Column: column}
	case '>':
		if l.matchChar('=') {
			return Token{Type: GreaterEqual, Literal: ">=", Line: line, Column: column}
		}
		return Token{Type: Greater, Literal: ">", Line: line, Column: column}
	case '&':
		return Token{Type: And, Literal: "&", Line: line, Column: column}
	case '|':
		return Token{Type: Or, Literal: "|", Line: line, Column: column}
	case '!':
		if l.matchChar('=') {
			return Token{Type: NotEqual, Literal: "!=", Line: line, Column: column}
		}
		return Token{Type: Not, Literal: "!", Line: line, Column: column}
	case '"':
		fallthrough
	case '\'':
		return Token{Type: String, Literal: l.readString(ch), Line: line, Column: column}
	case ' ', '\t', '\r':
		return l.NextToken()
	case '\n':
		l.line++
		l.column = 0
		return l.NextToken()
	case 0:
		return Token{Type: Eof, Literal: "", Line: line, Column: column}
	default:
		if isLetter(ch) {
			l.position--
			l.column--
			return l.readIdentifier()
		} else if isDigit(ch) {
			l.position--
			l.column--
			return l.readNumber()
		} else {
			panic("Unknown token type: " + string(ch) + " (" + strconv.Itoa(l.line) + ":" + strconv.Itoa(l.column) + ")")
		}
	}
	return Token{}
}

func notAtEnd(ch byte) byte {
	if ch == 0 {
		panic("Unexpected end of file")
	}
	return ch
}

func (l *Lexer) readIdentifier() Token {
	var position = l.position
	var line = l.line
	var column = l.column
	for isLetter(l.peekChar()) || isDigit(l.peekChar()) {
		notAtEnd(l.readChar())
	}
	if tokenType, ok := keywords[l.input[position:l.position]]; ok {
		return Token{Type: tokenType, Literal: l.input[position:l.position], Line: line, Column: column}
	}
	return Token{Type: Id, Literal: l.input[position:l.position], Line: line, Column: column}
}

func (l *Lexer) readNumber() Token {
	var position = l.position
	var line = l.line
	var column = l.column
	for isDigit(l.peekChar()) {
		notAtEnd(l.readChar())
	}
	if l.peekChar() == '.' {
		notAtEnd(l.readChar())
		for isDigit(l.peekChar()) {
			notAtEnd(l.readChar())
		}
	}
	return Token{Type: Number, Literal: l.input[position:l.position], Line: line, Column: column}
}

func (l *Lexer) readString(ch byte) string {
	var position = l.position
	for l.peekChar() != ch {
		var ch = notAtEnd(l.readChar())
		if ch == '\n' {
			l.line++
		}
		if ch == '\\' {
			if l.peekChar() == ch {
				l.readChar()
			}
		}
	}
	notAtEnd(l.readChar())
	var literal = l.input[position : l.position-1]
	literal = strings.ReplaceAll(literal, "\\"+string(ch), string(ch))
	literal = strings.ReplaceAll(literal, "\\n", "\n")
	literal = strings.ReplaceAll(literal, "\\r", "\r")
	literal = strings.ReplaceAll(literal, "\\t", "\t")
	return literal
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '$'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
