package main

import "golang.org/x/exp/constraints"

type Pos struct {
	Line   int
	Column int
}

func (p Pos) Pos() (line, col int) {
	return p.Line, p.Column
}

func NewPos(line, column int) Pos {
	return Pos{Line: line, Column: column}
}

func TokenPos(t Token) Pos {
	return NewPos(t.Line, t.Column)
}

type Statement interface {
	PosFrom() Pos
}

type Expression interface {
	PosFrom() Pos
}

type MyStatement struct {
	Name  string
	Value *Expression
	Pos
}

func (s MyStatement) PosFrom() Pos {
	return s.Pos
}

type SubStatement struct {
	Name   string
	Params []string
	Body   []Statement
	Pos
}

func (s SubStatement) PosFrom() Pos {
	return s.Pos
}

type ElseIf struct {
	Condition Expression
	Then      []Statement
}

type IfStatement struct {
	Conditions Expression
	Then       []Statement
	ElseIfs    []ElseIf
	Else_      []Statement
	Pos
}

func (s IfStatement) PosFrom() Pos {
	return s.Pos
}

type UnlessStatement struct {
	Condition Expression
	Then      []Statement
	ElseIfs   []ElseIf
	Else_     []Statement
	Pos
}

func (s UnlessStatement) PosFrom() Pos {
	return s.Pos
}

type WhileStatement struct {
	Condition Expression
	Body      []Statement
	Pos
}

func (s WhileStatement) PosFrom() Pos {
	return s.Pos
}

type DoWhileStatement struct {
	Body      []Statement
	Condition Expression
	Pos
}

func (s DoWhileStatement) PosFrom() Pos {
	return s.Pos
}

type UntilStatement struct {
	Condition Expression
	Body      []Statement
	Pos
}

func (s UntilStatement) PosFrom() Pos {
	return s.Pos
}

type DoUntilStatement struct {
	Body      []Statement
	Condition Expression
	Pos
}

func (s DoUntilStatement) PosFrom() Pos {
	return s.Pos
}

type ForStatement struct {
	Name       string
	Expression Expression
	Body       []Statement
	Pos
}

func (s ForStatement) PosFrom() Pos {
	return s.Pos
}

type Branch struct {
	Condition Expression
	Then      []Statement
}

type WhenStatement struct {
	Cases []Branch
	Else_ []Statement
	Pos
}

func (s WhenStatement) PosFrom() Pos {
	return s.Pos
}

type WhenMatchStatement struct {
	Value Expression
	Cases []Branch
	Else_ []Statement
	Pos
}

func (s WhenMatchStatement) PosFrom() Pos {
	return s.Pos
}

type CallStatement struct {
	Function Expression
	Args     []Expression
	Pos
}

func (s CallStatement) PosFrom() Pos {
	return s.Pos
}

type ReturnStatement struct {
	Value *Expression
	Pos
}

func (s ReturnStatement) PosFrom() Pos {
	return s.Pos
}

type AssignmentStatement struct {
	Left  Expression
	Value Expression
	Pos
}

func (s AssignmentStatement) PosFrom() Pos {
	return s.Pos
}

type BreakStatement struct {
	Pos
}

func (s BreakStatement) PosFrom() Pos {
	return s.Pos
}

type NextStatement struct {
	Pos
}

func (s NextStatement) PosFrom() Pos {
	return s.Pos
}

type Numeric interface {
	constraints.Float | constraints.Integer
}

type NumberLiteral[T Numeric] struct {
	Value T
	Pos
}

func (s NumberLiteral[T]) PosFrom() Pos {
	return s.Pos
}

type StringLiteral struct {
	Value string
	Pos
}

func (s StringLiteral) PosFrom() Pos {
	return s.Pos
}

type BooleanLiteral struct {
	Value bool
	Pos
}

func (s BooleanLiteral) PosFrom() Pos {
	return s.Pos
}

type NilLiteral struct {
	Pos
}

func (s NilLiteral) PosFrom() Pos {
	return s.Pos
}

type ArrayLiteral struct {
	Values []Expression
	Pos
}

func (s ArrayLiteral) PosFrom() Pos {
	return s.Pos
}

type HashLiteral struct {
	Pairs map[Expression]Expression
	Pos
}

func (s HashLiteral) PosFrom() Pos {
	return s.Pos
}

type Variable struct {
	Name string
	Pos
}

func (s Variable) PosFrom() Pos {
	return s.Pos
}

type Index struct {
	Left  Expression
	Index Expression
	Pos
}

func (s Index) PosFrom() Pos {
	return s.Pos
}

type Member struct {
	Left   Expression
	Member string
	Pos
}

func (s Member) PosFrom() Pos {
	return s.Pos
}

type Call struct {
	Function Expression
	Args     []Expression
	Pos
}

func (s Call) PosFrom() Pos {
	return s.Pos
}

type Unary struct {
	Operator TokenType
	Right    Expression
	Pos
}

func (s Unary) PosFrom() Pos {
	return s.Pos
}

type Binary struct {
	Left     Expression
	Operator TokenType
	Right    Expression
	Pos
}

func (s Binary) PosFrom() Pos {
	return s.Pos
}

type BlockExpression struct {
	Body Expression
	Pos
}

func (s BlockExpression) PosFrom() Pos {
	return s.Pos
}

type FunctionLiteral struct {
	Params []string
	Body   []Statement
	Pos
}

func (s FunctionLiteral) PosFrom() Pos {
	return s.Pos
}

type Increment struct {
	Left Expression
	By   Expression
}

func (s Increment) PosFrom() Pos {
	return s.Left.PosFrom()
}

type Decrement struct {
	Left Expression
	By   Expression
}

func (s Decrement) PosFrom() Pos {
	return s.Left.PosFrom()
}
