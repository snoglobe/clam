package main

import (
	"fmt"
	"reflect"
	"strconv"
)

type Interpreter struct {
	Program   []Statement
	Variables []map[string]interface{}
}

type ReturnValue struct {
	Value interface{}
}

func NewInterpreter(program []Statement) *Interpreter {
	var result = &Interpreter{Program: program}
	result.Variables = append(result.Variables, make(map[string]interface{}))
	return result
}

func (i *Interpreter) Run() {
	for _, stmt := range i.Program {
		i.exec(stmt)
	}
}

func (i *Interpreter) exec(stmt Statement) {
	switch stmt := stmt.(type) {
	case *MyStatement:
		if _, ok := i.Variables[len(i.Variables)-1][stmt.Name]; !ok {
			if stmt.Value != nil {
				i.Variables[len(i.Variables)-1][stmt.Name] = i.eval(*stmt.Value)
			} else {
				i.Variables[len(i.Variables)-1][stmt.Name] = nil
			}
		} else {
			panic("Variable already defined: " + stmt.Name)
		}
	case *SubStatement:
		if _, ok := i.Variables[len(i.Variables)-1][stmt.Name]; !ok {
			i.Variables[len(i.Variables)-1][stmt.Name] = func(args ...interface{}) (v interface{}) {
				var prev = i.Variables
				i.Variables = make([]map[string]interface{}, 2)
				i.Variables[0] = prev[0]
				i.Variables[1] = make(map[string]interface{})
				for j, arg := range stmt.Params {
					i.Variables[1][arg] = args[j]
				}
				defer func() {
					i.Variables = prev
				}()
				defer func() {
					if r := recover(); r != nil {
						if returnValue, ok := r.(ReturnValue); ok {
							v = returnValue.Value
						} else {
							panic(r)
						}
					}
				}()
				for _, s := range stmt.Body {
					i.exec(s)
				}
				return nil
			}
		} else {
			panic("Variable already defined: " + stmt.Name)
		}
	case *IfStatement:
		if truthy(i.eval(stmt.Conditions)) {
			for _, s := range stmt.Then {
				i.exec(s)
			}
		} else {
			for _, elseif := range stmt.ElseIfs {
				if truthy(i.eval(elseif.Condition)) {
					for _, s := range elseif.Then {
						i.exec(s)
					}
					return
				}
			}
			for _, s := range stmt.Else_ {
				i.exec(s)
			}
		}
	case *UnlessStatement:
		if !truthy(i.eval(stmt.Condition)) {
			for _, s := range stmt.Then {
				i.exec(s)
			}
		} else {
			for _, elseif := range stmt.ElseIfs {
				if truthy(i.eval(elseif.Condition)) {
					for _, s := range elseif.Then {
						i.exec(s)
					}
					return
				}
			}
			for _, s := range stmt.Else_ {
				i.exec(s)
			}
		}
	case *ReturnStatement:
		if stmt.Value != nil {
			panic(ReturnValue{i.eval(*stmt.Value)})
		} else {
			panic(ReturnValue{nil})
		}
	case *WhileStatement:
		for truthy(i.eval(stmt.Condition)) {
			for _, s := range stmt.Body {
				i.exec(s)
			}
		}
	case *UntilStatement:
		for !truthy(i.eval(stmt.Condition)) {
			for _, s := range stmt.Body {
				i.exec(s)
			}
		}
	case *ForStatement:
		var value = i.eval(stmt.Expression)
		if array, ok := value.([]interface{}); ok {
			for _, element := range array {
				i.Variables[len(i.Variables)-1][stmt.Name] = element
				for _, s := range stmt.Body {
					i.exec(s)
				}
			}
		} else if hash, ok := value.(map[interface{}]interface{}); ok {
			for key, element := range hash {
				i.Variables[len(i.Variables)-1][stmt.Name] = []interface{}{key, element}
				for _, s := range stmt.Body {
					i.exec(s)
				}
			}
		} else {
			panic("For loop not supported for type: " + fmt.Sprintf("%T", value))
		}
	case *DoWhileStatement:
		for {
			for _, s := range stmt.Body {
				i.exec(s)
			}
			if !truthy(i.eval(stmt.Condition)) {
				break
			}
		}
	case *DoUntilStatement:
		for {
			for _, s := range stmt.Body {
				i.exec(s)
			}
			if truthy(i.eval(stmt.Condition)) {
				break
			}
		}
	case *WhenStatement:
		for _, branch := range stmt.Cases {
			if truthy(i.eval(branch.Condition)) {
				for _, s := range branch.Then {
					i.exec(s)
				}
				return
			}
		}
		for _, s := range stmt.Else_ {
			i.exec(s)
		}
	case *WhenMatchStatement:
		var value = i.eval(stmt.Value)
		for _, branch := range stmt.Cases {
			if i.eval(branch.Condition) == value {
				for _, s := range branch.Then {
					i.exec(s)
				}
				return
			}
		}
		for _, s := range stmt.Else_ {
			i.exec(s)
		}
	case *CallStatement:
		var function = i.eval(stmt.Function)
		var args = make([]interface{}, len(stmt.Args))
		for j, arg := range stmt.Args {
			args[j] = i.eval(arg)
		}
		if anyFn, ok := function.(func(...interface{}) interface{}); ok {
			anyFn(args...)
		} else {
			// use go's reflection to call method
			reflectValue := reflect.ValueOf(function)
			if reflectValue.Kind() == reflect.Func {
				var reflectArgs = make([]reflect.Value, len(args))
				for j, arg := range args {
					reflectArgs[j] = reflect.ValueOf(arg)
				}
				reflectValue.Call(reflectArgs)
			} else {
				panic("Not a function: " + fmt.Sprintf("%T", function))
			}
		}
	case *AssignmentStatement:
		switch stmt.Left.(type) {
		case *Variable:
			var value = i.eval(stmt.Value)
			var scope = len(i.Variables) - 1
			for scope >= 0 {
				if _, ok := i.Variables[scope][stmt.Left.(*Variable).Name]; ok {
					i.Variables[scope][stmt.Left.(*Variable).Name] = value
					return
				}
				scope--
			}
			panic("Undefined variable: " + stmt.Left.(*Variable).Name)
		case *Index:
			var left = stmt.Left.(*Index)
			var value = i.eval(left.Left)
			var index = i.eval(left.Index)
			if array, ok := value.([]interface{}); ok {
				if idx, ok := index.(float64); ok {
					var idx = int(idx)
					if idx >= 0 && idx < len(array) {
						array[idx] = i.eval(stmt.Value)
						return
					}
					panic("Index out of range: " + strconv.Itoa(idx))
				}
				panic("Index must be an integer")
			}
			// use reflection to index
			reflectValue := reflect.ValueOf(value)
			if reflectValue.Kind() == reflect.Array || reflectValue.Kind() == reflect.Slice {
				if idx, ok := index.(float64); ok {
					var idx = int(idx)
					if idx >= 0 && idx < reflectValue.Len() {
						reflectValue.Index(idx).Set(reflect.ValueOf(i.eval(stmt.Value)))
						return
					}
					panic("Index out of range: " + strconv.Itoa(idx))
				}
				panic("Index must be an integer")
			} else if reflectValue.Kind() == reflect.Map {
				var reflectResult = reflectValue.MapIndex(reflect.ValueOf(index))
				if reflectResult.IsValid() {
					reflectResult.Set(reflect.ValueOf(i.eval(stmt.Value)))
					return
				}
				panic("Key not found: " + fmt.Sprintf("%v", index))
			}
			panic("Indexing not supported for type: " + fmt.Sprintf("%T", value))
		case *Member:
			var left = stmt.Left.(*Member)
			var value = i.eval(left.Left)
			if hash, ok := value.(map[interface{}]interface{}); ok {
				hash[left.Member] = i.eval(stmt.Value)
				return
			}
			panic("Member access not supported for type: " + fmt.Sprintf("%T", value))
		default:
			panic("Assignment not supported for type: " + fmt.Sprintf("%T", stmt.Left))
		}
	case *Increment:
		i.inc(stmt)
	case *Decrement:
		i.dec(stmt)
	default:
		panic("Statement not supported: " + fmt.Sprintf("%T", stmt))
	}
}

func (i *Interpreter) inc(node *Increment) {
	var byExpr = i.eval(node.By)
	if _, ok := byExpr.(float64); !ok {
		panic("By must be a number")
	}
	var by = byExpr.(float64)
	switch target := node.Left.(type) {
	case *Variable:
		var scope = len(i.Variables) - 1
		for scope >= 0 {
			if value, ok := i.Variables[scope][target.Name]; ok {
				if value, ok := value.(float64); ok {
					i.Variables[scope][target.Name] = value + by
					return
				}
				panic("Variable must be a number")
			}
			scope--
		}
		panic("Undefined variable: " + target.Name)
	case *Index:
		var value = i.eval(target.Left)
		var index = i.eval(target.Index)
		if array, ok := value.([]interface{}); ok {
			if idx, ok := index.(float64); ok {
				var idx = int(idx)
				if idx >= 0 && idx < len(array) {
					if value, ok := array[idx].(float64); ok {
						array[idx] = value + by
						return
					}
					panic("Value must be a number")
				}
				panic("Index out of range: " + strconv.Itoa(idx))
			}
			panic("Index must be an integer")
		} else if hash, ok := value.(map[interface{}]interface{}); ok {
			if _, ok := hash[index]; ok {
				panic("Cannot increment hash value")
			}
			panic("Key not found: " + index.(string))
		}
		// use reflection to index
		reflectValue := reflect.ValueOf(value)
		if reflectValue.Kind() == reflect.Array || reflectValue.Kind() == reflect.Slice {
			if idx, ok := index.(float64); ok {
				var idx = int(idx)
				if idx >= 0 && idx < reflectValue.Len() {
					reflectValue.Index(idx).Set(reflect.ValueOf(reflectValue.Index(idx).Interface().(float64) + by))
					return
				}
				panic("Index out of range: " + strconv.Itoa(idx))
			}
			panic("Index must be an integer")
		} else if reflectValue.Kind() == reflect.Map {
			var reflectResult = reflectValue.MapIndex(reflect.ValueOf(index))
			if reflectResult.IsValid() {
				reflectResult.Set(reflect.ValueOf(reflectResult.Interface().(float64) + by))
				return
			}
			panic("Key not found: " + fmt.Sprintf("%v", index))
		}
		panic("Indexing not supported for type: " + fmt.Sprintf("%T", value))
	case *Member:
		var value = i.eval(target.Left)
		if hash, ok := value.(map[interface{}]interface{}); ok {
			if value, ok := hash[target.Member]; ok {
				if value, ok := value.(float64); ok {
					hash[target.Member] = value + by
					return
				}
				panic("Value must be a number")
			}
			panic("Key not found: " + target.Member)
		}
		panic("Member access not supported for type: " + fmt.Sprintf("%T", value))
	default:
		panic("Increment not supported for type: " + fmt.Sprintf("%T", node.Left))
	}
}

func (i *Interpreter) dec(node *Decrement) {
	var byExpr = i.eval(node.By)
	if _, ok := byExpr.(float64); !ok {
		panic("By must be a number")
	}
	var by = byExpr.(float64)
	switch target := node.Left.(type) {
	case *Variable:
		var scope = len(i.Variables) - 1
		for scope >= 0 {
			if value, ok := i.Variables[scope][target.Name]; ok {
				if value, ok := value.(float64); ok {
					i.Variables[scope][target.Name] = value - by
					return
				}
				panic("Variable must be a number")
			}
			scope--
		}
		panic("Undefined variable: " + target.Name)
	case *Index:
		var value = i.eval(target.Left)
		var index = i.eval(target.Index)
		if array, ok := value.([]interface{}); ok {
			if idx, ok := index.(float64); ok {
				var idx = int(idx)
				if idx >= 0 && idx < len(array) {
					if value, ok := array[idx].(float64); ok {
						array[idx] = value - by
						return
					}
					panic("Value must be a number")
				}
				panic("Index out of range: " + strconv.Itoa(idx))
			}
			panic("Index must be an integer")
		} else if hash, ok := value.(map[interface{}]interface{}); ok {
			if _, ok := hash[index]; ok {
				panic("Cannot decrement hash value")
			}
			panic("Key not found: " + index.(string))
		}
		// use reflection to index
		reflectValue := reflect.ValueOf(value)
		if reflectValue.Kind() == reflect.Array || reflectValue.Kind() == reflect.Slice {
			if idx, ok := index.(float64); ok {
				var idx = int(idx)
				if idx >= 0 && idx < reflectValue.Len() {
					reflectValue.Index(idx).Set(reflect.ValueOf(reflectValue.Index(idx).Interface().(float64) - by))
					return
				}
				panic("Index out of range: " + strconv.Itoa(idx))
			}
			panic("Index must be an integer")
		} else if reflectValue.Kind() == reflect.Map {
			var reflectResult = reflectValue.MapIndex(reflect.ValueOf(index))
			if reflectResult.IsValid() {
				reflectResult.Set(reflect.ValueOf(reflectResult.Interface().(float64) - by))
				return
			}
			panic("Key not found: " + fmt.Sprintf("%v", index))
		}
		panic("Indexing not supported for type: " + fmt.Sprintf("%T", value))
	case *Member:
		var value = i.eval(target.Left)
		if hash, ok := value.(map[interface{}]interface{}); ok {
			if value, ok := hash[target.Member]; ok {
				if value, ok := value.(float64); ok {
					hash[target.Member] = value - by
					return
				}
				panic("Value must be a number")
			}
			panic("Key not found: " + target.Member)
		}
		panic("Member access not supported for type: " + fmt.Sprintf("%T", value))
	default:
		panic("Decrement not supported for type: " + fmt.Sprintf("%T", node.Left))
	}
}

func truthy(value interface{}) bool {
	if value == nil {
		return false
	}
	if value, ok := value.(bool); ok {
		return value
	}
	return true
}

func (i *Interpreter) eval(expr Expression) interface{} {
	switch expr := expr.(type) {
	case *NumberLiteral[int]:
		return expr.Value
	case *NumberLiteral[float64]:
		return expr.Value
	case *StringLiteral:
		return expr.Value
	case *BooleanLiteral:
		return expr.Value
	case *NilLiteral:
		return nil
	case *ArrayLiteral:
		var result = make([]interface{}, len(expr.Values))
		for j, value := range expr.Values {
			result[j] = i.eval(value)
		}
		return result
	case *HashLiteral:
		var result = make(map[interface{}]interface{})
		for key, value := range expr.Pairs {
			result[i.eval(key)] = i.eval(value)
		}
		return result
	case *Variable:
		var scope = len(i.Variables) - 1
		for scope >= 0 {
			if value, ok := i.Variables[scope][expr.Name]; ok {
				return value
			}
			scope--
		}
		panic("Undefined variable: " + expr.Name)
	case *Index:
		var value = i.eval(expr.Left)
		var index = i.eval(expr.Index)
		if array, ok := value.([]interface{}); ok {
			if idx, ok := index.(float64); ok {
				var idx = int(idx)
				if idx >= 0 && idx < len(array) {
					return array[idx]
				}
				panic("Index out of range: " + strconv.Itoa(idx))
			}
			panic("Index must be an integer")
		} else if hash, ok := value.(map[interface{}]interface{}); ok {
			if result, ok := hash[index]; ok {
				return result
			}
			panic("Key not found: " + index.(string))
		} else {
			// use reflection to index
			reflectValue := reflect.ValueOf(value)
			if reflectValue.Kind() == reflect.Array || reflectValue.Kind() == reflect.Slice {
				if idx, ok := index.(float64); ok {
					var idx = int(idx)
					if idx >= 0 && idx < reflectValue.Len() {
						return reflectValue.Index(idx).Interface()
					}
					panic("Index out of range: " + strconv.Itoa(idx))
				}
				panic("Index must be an integer")
			} else if reflectValue.Kind() == reflect.Map {
				var reflectResult = reflectValue.MapIndex(reflect.ValueOf(index))
				if reflectResult.IsValid() {
					return reflectResult.Interface()
				}
				panic("Key not found: " + fmt.Sprintf("%v", index))
			}
		}
	case *Call:
		var function = i.eval(expr.Function)
		var args = make([]interface{}, len(expr.Args))
		for j, arg := range expr.Args {
			args[j] = i.eval(arg)
		}
		if anyFn, ok := function.(func(...interface{}) interface{}); ok {
			return anyFn(args...)
		} else {
			// use go's reflection to call method
			reflectValue := reflect.ValueOf(function)
			if reflectValue.Kind() == reflect.Func {
				var reflectArgs = make([]reflect.Value, len(args))
				for j, arg := range args {
					reflectArgs[j] = reflect.ValueOf(arg)
				}
				var reflectResult = reflectValue.Call(reflectArgs)
				if len(reflectResult) == 1 {
					return reflectResult[0].Interface()
				}
				var result = make([]interface{}, len(reflectResult))
				for j, value := range reflectResult {
					result[j] = value.Interface()
				}
				return result
			}
			panic("Not a function: " + fmt.Sprintf("%T", function))
		}
	case *Member:
		var value = i.eval(expr.Left)
		if hash, ok := value.(map[interface{}]interface{}); ok {
			if result, ok := hash[expr.Member]; ok {
				return result
			}
			panic("Key not found: " + expr.Member)
		}
		panic("Member access not supported for type: " + fmt.Sprintf("%T", value))
	case *Unary:
		var value = i.eval(expr.Right)
		switch expr.Operator {
		case Not:
			return !truthy(value)
		case Minus:
			if value, ok := value.(float64); ok {
				return -value
			}
		}
		panic("Unary operator not supported for type: " + fmt.Sprintf("%T", value))
	case *Binary:
		var left = i.eval(expr.Left)
		switch expr.Operator {
		case Plus:
			var right = i.eval(expr.Right)
			if left, ok := left.(string); ok {
				if right, ok := right.(string); ok {
					return left + right
				}
				return left + fmt.Sprintf("%v", right)
			}
			if left, ok := left.(float64); ok {
				if right, ok := right.(float64); ok {
					return left + right
				}
				panic("Right operand must be a number")
			}
			panic("Left operand must be a string or a number")
		case Minus:
			var right = i.eval(expr.Right)
			if left, ok := left.(float64); ok {
				if right, ok := right.(float64); ok {
					return left - right
				}
				panic("Right operand must be a number")
			}
			panic("Left operand must be a number")
		case Multiply:
			var right = i.eval(expr.Right)
			if left, ok := left.(float64); ok {
				if right, ok := right.(float64); ok {
					return left * right
				}
				panic("Right operand must be a number")
			}
			panic("Left operand must be a number")
		case Divide:
			var right = i.eval(expr.Right)
			if left, ok := left.(float64); ok {
				if right, ok := right.(float64); ok {
					return left / right
				}
				panic("Right operand must be a number")
			}
			panic("Left operand must be a number")
		case Modulo:
			var right = i.eval(expr.Right)
			if left, ok := left.(float64); ok {
				if right, ok := right.(float64); ok {
					return float64(int(left) % int(right))
				}
				panic("Right operand must be a number")
			}
		case Equal:
			var right = i.eval(expr.Right)
			return left == right
		case NotEqual:
			var right = i.eval(expr.Right)
			return left != right
		case Less:
			var right = i.eval(expr.Right)
			if left, ok := left.(float64); ok {
				if right, ok := right.(float64); ok {
					return left < right
				}
				panic("Right operand must be a number")
			}
			panic("Left operand must be a number")
		case LessEqual:
			var right = i.eval(expr.Right)
			if left, ok := left.(float64); ok {
				if right, ok := right.(float64); ok {
					return left <= right
				}
				panic("Right operand must be a number")
			}
			panic("Left operand must be a number")
		case Greater:
			var right = i.eval(expr.Right)
			if left, ok := left.(float64); ok {
				if right, ok := right.(float64); ok {
					return left > right
				}
				panic("Right operand must be a number")
			}
			panic("Left operand must be a number")
		case GreaterEqual:
			var right = i.eval(expr.Right)
			if left, ok := left.(float64); ok {
				if right, ok := right.(float64); ok {
					return left >= right
				}
				panic("Right operand must be a number")
			}
			panic("Left operand must be a number")
		case And:
			if truthy(left) {
				return i.eval(expr.Right)
			}
			return left
		case Or:
			if truthy(left) {
				return left
			}
			return i.eval(expr.Right)
		}
	case *BlockExpression:
		return func(arg interface{}) interface{} {
			var prev = i.Variables
			i.Variables = make([]map[string]interface{}, 2)
			i.Variables[0] = prev[0]
			i.Variables[1] = make(map[string]interface{})
			i.Variables[1]["it"] = arg
			defer func() {
				i.Variables = prev
			}()
			return i.eval(expr.Body)
		}
	case *FunctionLiteral:
		return func(args ...interface{}) (v interface{}) {
			var prev = i.Variables
			i.Variables = make([]map[string]interface{}, 2)
			i.Variables[0] = prev[0]
			i.Variables[1] = make(map[string]interface{})
			for j, arg := range expr.Params {
				i.Variables[1][arg] = args[j]
			}
			defer func() {
				i.Variables = prev
			}()
			// execute block, catch return value using recover
			defer func() {
				if r := recover(); r != nil {
					if returnValue, ok := r.(ReturnValue); ok {
						v = returnValue.Value
					}
					panic(r)
				}
			}()
			for _, s := range expr.Body {
				i.exec(s)
			}
			return nil
		}
	case *Increment:
		i.inc(expr)
		return i.eval(expr.Left)
	case *Decrement:
		i.dec(expr)
		return i.eval(expr.Left)
	}
	return nil
}
