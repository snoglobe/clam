package main

import (
	_ "embed"
)

// embed file as string

//go:embed example.clm
var code string

func main() {
	var lexer = NewLexer(code)
	var parser = NewParser(lexer)
	var program []Statement
	for {
		program = append(program, parser.stmt())
		if parser.next() == Eof {
			break
		}
	}
	var interpreter = NewInterpreter(program)
	for k, v := range library {
		interpreter.Variables[0][k] = v
	}
	for _, doc := range docs {
		println(doc)
	}
	interpreter.Run()
}
