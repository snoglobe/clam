package main

import (
	"strconv"
)

type Parser struct {
	lexer *Lexer
	token Token
}

func NewParser(lexer *Lexer) *Parser {
	var parser = &Parser{lexer: lexer}
	parser.token = lexer.NextToken()
	return parser
}

func (p *Parser) peek(type_ TokenType) bool {
	return p.token.Type == type_
}

func (p *Parser) next() TokenType {
	return p.token.Type
}

func (p *Parser) match(t TokenType) bool {
	if p.next() == t {
		p.token = p.lexer.NextToken()
		return true
	}
	return false
}

func (p *Parser) eat(t TokenType) Token {
	var token = p.token
	if !p.match(t) {
		panic("unexpected token: " + p.token.String())
	}
	return token
}

func (p *Parser) expr() Expression {
	return p.or()
}

func (p *Parser) or() Expression {
	var expr = p.and()
	for p.peek(Or) {
		var pos = TokenPos(p.token)
		p.eat(Or)
		expr = &Binary{Left: expr, Operator: Or, Right: p.and(), Pos: pos}
	}
	return expr
}

func (p *Parser) and() Expression {
	var expr = p.equality()
	for p.peek(And) {
		var pos = TokenPos(p.token)
		p.eat(And)
		expr = &Binary{Left: expr, Operator: And, Right: p.equality(), Pos: pos}
	}
	return expr
}

func (p *Parser) equality() Expression {
	var expr = p.comparison()
	for op := p.next(); op == Equal || op == NotEqual; op = p.next() {
		var pos = TokenPos(p.token)
		p.eat(op)
		expr = &Binary{Left: expr, Operator: op, Right: p.comparison(), Pos: pos}
	}
	return expr
}

func (p *Parser) comparison() Expression {
	var expr = p.addition()
	for op := p.next(); op == Less || op == LessEqual || op == Greater || op == GreaterEqual; op = p.next() {
		var pos = TokenPos(p.token)
		p.eat(op)
		expr = &Binary{Left: expr, Operator: op, Right: p.addition(), Pos: pos}
	}
	return expr
}

func (p *Parser) addition() Expression {
	var expr = p.multiplication()
	for op := p.next(); op == Plus || op == Minus; op = p.next() {
		var pos = TokenPos(p.token)
		p.eat(op)
		expr = &Binary{Left: expr, Operator: op, Right: p.multiplication(), Pos: pos}
	}
	return expr
}

func (p *Parser) multiplication() Expression {
	var expr = p.unary()
	for op := p.next(); op == Multiply || op == Divide || op == Modulo; op = p.next() {
		var pos = TokenPos(p.token)
		p.eat(op)
		expr = &Binary{Left: expr, Operator: op, Right: p.unary(), Pos: pos}
	}
	return expr
}

func (p *Parser) unary() Expression {
	var pos = TokenPos(p.token)
	if p.match(Not) {
		return &Unary{Operator: Not, Right: p.unary(), Pos: pos}
	}
	if p.match(Minus) {
		return &Unary{Operator: Minus, Right: p.unary(), Pos: pos}
	}
	return p.call(false)
}

func (p *Parser) call(dotOnly bool) Expression {
	var expr = p.primary()
	var pos = TokenPos(p.token)
	for {
		if p.match(LeftParen) && !dotOnly {
			expr = p.finishCall(expr, pos)
		} else if p.match(Dot) {
			expr = p.finishMember(expr, pos)
		} else if p.match(LeftBracket) && !dotOnly {
			expr = p.finishIndex(expr, pos)
		} else {
			break
		}
	}
	return expr
}

func (p *Parser) finishCall(expr Expression, pos Pos) Expression {
	var args []Expression
	for !p.match(RightParen) {
		args = append(args, p.expr())
		if p.next() != RightParen {
			p.eat(Comma)
		}
	}
	return &Call{Function: expr, Args: args, Pos: pos}
}

func (p *Parser) finishMember(expr Expression, pos Pos) Expression {
	return &Member{Left: expr, Member: p.eat(Id).Literal, Pos: pos}
}

func (p *Parser) finishIndex(expr Expression, pos Pos) Expression {
	var index = p.expr()
	p.eat(RightBracket)
	return &Index{Left: expr, Index: index, Pos: pos}
}

func (p *Parser) primary() Expression {
	var pos = TokenPos(p.token)
	switch p.next() {
	case Id:
		return &Variable{Name: p.eat(Id).Literal, Pos: pos}
	case String:
		return &StringLiteral{Value: p.eat(String).Literal, Pos: pos}
	case Number:
		var number = p.eat(Number)
		if i, err := strconv.Atoi(number.Literal); err == nil {
			return &NumberLiteral[int]{Value: i, Pos: pos}
		}
		var f, _ = strconv.ParseFloat(number.Literal, 64)
		return &NumberLiteral[float64]{Value: f, Pos: pos}
	case True:
		p.eat(True)
		return &BooleanLiteral{Value: true, Pos: pos}
	case False:
		p.eat(False)
		return &BooleanLiteral{Value: false, Pos: pos}
	case Nil:
		p.eat(Nil)
		return &NilLiteral{Pos: pos}
	case LeftParen:
		p.eat(LeftParen)
		var expr = p.expr()
		p.eat(RightParen)
		return expr
	case LeftBracket:
		return p.array()
	case LeftBrace:
		return p.exprBlock()
	case Inc:
		p.eat(Inc)
		var value = p.expr()
		if p.match(By) {
			return &Increment{Left: value, By: p.expr()}
		}
		return &Increment{Left: value, By: &NumberLiteral[int]{Value: 1, Pos: pos}}
	case Dec:
		p.eat(Dec)
		var value = p.expr()
		if p.match(By) {
			return &Decrement{Left: value, By: p.expr()}
		}
		return &Decrement{Left: value, By: &NumberLiteral[int]{Value: 1, Pos: pos}}
	default:
		panic("unexpected token: " + p.token.String())
	}
}

func (p *Parser) array() Expression {
	var pos = TokenPos(p.token)
	p.eat(LeftBracket)
	if p.match(Colon) {
		p.eat(RightBracket)
		return &HashLiteral{Pairs: map[Expression]Expression{}, Pos: pos}
	} else if p.match(RightBracket) {
		return &ArrayLiteral{Values: []Expression{}, Pos: pos}
	} else {
		var first = p.expr()
		if p.match(Colon) {
			var values = map[Expression]Expression{}
			values[first] = p.expr()
			for p.match(Comma) {
				var key = p.expr()
				p.eat(Colon)
				values[key] = p.expr()
			}
			p.eat(RightBracket)
			return &HashLiteral{Pairs: values, Pos: pos}
		} else {
			var values = []Expression{first}
			for p.match(Comma) {
				values = append(values, p.expr())
			}
			p.eat(RightBracket)
			return &ArrayLiteral{Values: values, Pos: pos}
		}
	}
}

func (p *Parser) exprBlock() Expression {
	var pos = TokenPos(p.token)
	p.eat(LeftBrace)
	var expr = p.expr()
	p.eat(RightBrace)
	return &BlockExpression{Body: expr, Pos: pos}
}

func (p *Parser) stmt() Statement {
	var stmt Statement
	switch p.next() {
	case If:
		stmt = p.ifStmt()
		return stmt
	case Unless:
		stmt = p.unlessStmt()
		return stmt
	case While:
		stmt = p.whileStmt()
		return stmt
	case Until:
		stmt = p.untilStmt()
		return stmt
	case For:
		stmt = p.forStmt()
		return stmt
	case Do:
		stmt = p.doStmt()
		p.eat(Semicolon)
		return stmt
	case Sub:
		stmt = p.subStmt()
		return stmt
	case My:
		stmt = p.myStmt()
		p.eat(Semicolon)
		return stmt
	case When:
		stmt = p.whenStmt()
		return stmt
	case Return:
		stmt = p.returnStmt()
	case Inc:
		var pos = TokenPos(p.token)
		p.eat(Inc)
		var value = p.expr()
		var by Expression = &NumberLiteral[int]{Value: 1, Pos: pos}
		if p.match(By) {
			by = p.expr()
		}
		stmt = &Increment{Left: value, By: by}
	case Dec:
		var pos = TokenPos(p.token)
		p.eat(Dec)
		var value = p.expr()
		var by Expression = &NumberLiteral[int]{Value: 1, Pos: pos}
		if p.match(By) {
			by = p.expr()
		}
		stmt = &Decrement{Left: value, By: by}
	default:
		var pos = TokenPos(p.token)
		var left = p.call(true)
		if p.match(Assign) {
			var right = p.expr()
			stmt = &AssignmentStatement{Left: left, Value: right, Pos: pos}
		} else {
			var args []Expression
			for !p.peek(Semicolon) &&
				!p.peek(If) &&
				!p.peek(Unless) &&
				!p.peek(While) &&
				!p.peek(Until) {
				args = append(args, p.expr())
			}
			stmt = &CallStatement{Function: left, Args: args, Pos: pos}
		}
	}
loop:
	for {
		switch p.next() {
		case If:
			p.eat(If)
			var condition = p.expr()
			stmt = &IfStatement{Conditions: condition, Then: []Statement{stmt}, ElseIfs: nil, Else_: nil, Pos: stmt.PosFrom()}
		case Unless:
			p.eat(Unless)
			var condition = p.expr()
			stmt = &UnlessStatement{Condition: condition, Then: []Statement{stmt}, ElseIfs: nil, Else_: nil, Pos: stmt.PosFrom()}
		case While:
			p.eat(While)
			var condition = p.expr()
			p.eat(Semicolon)
			stmt = &WhileStatement{Condition: condition, Body: []Statement{stmt}, Pos: stmt.PosFrom()}
		case Until:
			p.eat(Until)
			var condition = p.expr()
			stmt = &UntilStatement{Condition: condition, Body: []Statement{stmt}, Pos: stmt.PosFrom()}
		default:
			break loop
		}
	}
	p.eat(Semicolon)
	return stmt
}

func (p *Parser) ifStmt() Statement {
	var pos = TokenPos(p.token)
	p.eat(If)
	var condition = p.expr()
	var then []Statement
	p.eat(LeftBrace)
	for !p.match(RightBrace) {
		then = append(then, p.stmt())
	}
	var elseIfs []ElseIf
	for p.match(Else) {
		if p.match(If) {
			var condition = p.expr()
			var elseIfThen []Statement
			p.eat(LeftBrace)
			for !p.match(RightBrace) {
				elseIfThen = append(elseIfThen, p.stmt())
			}
			elseIfs = append(elseIfs, ElseIf{Condition: condition, Then: elseIfThen})
		} else if p.match(Unless) {
			var condition = p.expr()
			var unlessThen []Statement
			p.eat(LeftBrace)
			for !p.match(RightBrace) {
				unlessThen = append(unlessThen, p.stmt())
			}
			elseIfs = append(elseIfs, ElseIf{Condition: &Unary{Operator: Not, Right: condition, Pos: condition.PosFrom()}, Then: unlessThen})
		} else {
			var elseThen []Statement
			p.eat(LeftBrace)
			for !p.match(RightBrace) {
				elseThen = append(elseThen, p.stmt())
			}
			return &IfStatement{Conditions: condition, Then: then, ElseIfs: elseIfs, Else_: elseThen, Pos: pos}
		}
	}
	return &IfStatement{Conditions: condition, Then: then, ElseIfs: elseIfs, Else_: nil, Pos: pos}
}

func (p *Parser) unlessStmt() Statement {
	var pos = TokenPos(p.token)
	p.eat(Unless)
	var condition = p.expr()
	var then []Statement
	p.eat(LeftBrace)
	for !p.match(RightBrace) {
		then = append(then, p.stmt())
	}
	var elseIfs []ElseIf
	for p.match(Else) {
		if p.match(If) {
			var condition = p.expr()
			var elseIfThen []Statement
			p.eat(LeftBrace)
			for !p.match(RightBrace) {
				elseIfThen = append(elseIfThen, p.stmt())
			}
			elseIfs = append(elseIfs, ElseIf{Condition: condition, Then: elseIfThen})
		} else if p.match(Unless) {
			var condition = p.expr()
			var unlessThen []Statement
			p.eat(LeftBrace)
			for !p.match(RightBrace) {
				unlessThen = append(unlessThen, p.stmt())
			}
			elseIfs = append(elseIfs, ElseIf{Condition: &Unary{Operator: Not, Right: condition, Pos: condition.PosFrom()}, Then: unlessThen})
		} else {
			var elseThen []Statement
			p.eat(LeftBrace)
			for !p.match(RightBrace) {
				elseThen = append(elseThen, p.stmt())
			}
			return &UnlessStatement{Condition: condition, Then: then, ElseIfs: elseIfs, Else_: elseThen, Pos: pos}
		}
	}
	return &UnlessStatement{Condition: condition, Then: then, ElseIfs: nil, Else_: nil, Pos: pos}
}

func (p *Parser) whileStmt() Statement {
	var pos = TokenPos(p.token)
	p.eat(While)
	var condition = p.expr()
	var body []Statement
	p.eat(LeftBrace)
	for !p.match(RightBrace) {
		body = append(body, p.stmt())
	}
	return &WhileStatement{Condition: condition, Body: body, Pos: pos}
}

func (p *Parser) untilStmt() Statement {
	var pos = TokenPos(p.token)
	p.eat(Until)
	var condition = p.expr()
	var body []Statement
	p.eat(LeftBrace)
	for !p.match(RightBrace) {
		body = append(body, p.stmt())
	}
	return &UntilStatement{Condition: condition, Body: body, Pos: pos}
}

func (p *Parser) forStmt() Statement {
	var pos = TokenPos(p.token)
	p.eat(For)
	var left = p.expr()
	if p.match(In) {
		var right = p.expr()
		var body []Statement
		p.eat(LeftBrace)
		for !p.match(RightBrace) {
			body = append(body, p.stmt())
		}
		return &ForStatement{Name: left.(Variable).Name, Expression: right, Body: body, Pos: pos}
	} else {
		var body []Statement
		p.eat(LeftBrace)
		for !p.match(RightBrace) {
			body = append(body, p.stmt())
		}
		return &ForStatement{Name: "it", Expression: left, Body: body, Pos: pos}
	}
}

func (p *Parser) doStmt() Statement {
	var pos = TokenPos(p.token)
	p.eat(Do)
	var body []Statement
	p.eat(LeftBrace)
	for !p.match(RightBrace) {
		body = append(body, p.stmt())
	}
	if p.match(While) {
		var condition = p.expr()
		return &DoWhileStatement{Body: body, Condition: condition, Pos: pos}
	} else {
		p.eat(Until)
		var condition = p.expr()
		return &DoUntilStatement{Body: body, Condition: condition, Pos: pos}
	}
}

func (p *Parser) subStmt() Statement {
	var pos = TokenPos(p.token)
	p.eat(Sub)
	var name = p.eat(Id).Literal
	var args []string
	if p.match(LeftParen) {
		for !p.match(RightParen) {
			args = append(args, p.eat(Id).Literal)
			if p.next() != RightParen {
				p.eat(Comma)
			}
		}
	}
	var body []Statement
	p.eat(LeftBrace)
	for !p.match(RightBrace) {
		body = append(body, p.stmt())
	}
	return &SubStatement{Name: name, Params: args, Body: body, Pos: pos}
}

func (p *Parser) myStmt() Statement {
	var pos = TokenPos(p.token)
	p.eat(My)
	var name = p.eat(Id).Literal
	if p.match(Assign) {
		var value = p.expr()
		return &MyStatement{Name: name, Value: &value, Pos: pos}
	} else {
		return &MyStatement{Name: name, Value: nil, Pos: pos}
	}
}

func (p *Parser) whenStmt() Statement {
	var pos = TokenPos(p.token)
	p.eat(When)
	if p.match(LeftBrace) {
		var branches []Branch
		for !p.match(RightBrace) {
			if p.match(Case) {
				var condition = p.expr()
				var body []Statement
				p.eat(LeftBrace)
				for !p.match(RightBrace) {
					body = append(body, p.stmt())
				}
				branches = append(branches, Branch{Condition: condition, Then: body})
			} else if p.match(Else) {
				var body []Statement
				p.eat(LeftBrace)
				for !p.match(RightBrace) {
					body = append(body, p.stmt())
				}
				p.eat(RightBrace)
				return &WhenStatement{Cases: branches, Else_: body, Pos: pos}
			}
		}
		return &WhenStatement{Cases: branches, Else_: nil, Pos: pos}
	}
	var condition = p.expr()
	p.eat(LeftBrace)
	var branches []Branch
	for !p.match(RightBrace) {
		if p.match(Case) {
			var condition = p.expr()
			var body []Statement
			p.eat(LeftBrace)
			for !p.match(RightBrace) {
				body = append(body, p.stmt())
			}
			branches = append(branches, Branch{Condition: condition, Then: body})
		} else if p.match(Else) {
			var body []Statement
			p.eat(LeftBrace)
			for !p.match(RightBrace) {
				body = append(body, p.stmt())
			}
			p.eat(RightBrace)
			return &WhenMatchStatement{Value: condition, Cases: branches, Else_: body, Pos: pos}
		}
	}
	return &WhenMatchStatement{Value: condition, Cases: branches, Else_: nil, Pos: pos}
}

func (p *Parser) returnStmt() Statement {
	var pos = TokenPos(p.token)
	p.eat(Return)
	if p.match(Semicolon) {
		return &ReturnStatement{Value: nil, Pos: pos}
	}
	var value = p.expr()
	return &ReturnStatement{Value: &value, Pos: pos}
}
