package evaluator

import (
	"gibbon-lang/src/gibbon/ast"
	"gibbon-lang/src/gibbon/object"
)

// Package level boolean/null singletons for performance optimization
var TRUE = &object.Boolean{Value: true}
var FALSE = &object.Boolean{Value: false}
var NULL = &object.Null{}

// Main entry point for the evaluator
// It recursively walks the AST and evaluates each node to produce a value
// Parameters:
// - node: The AST node to evaluate
// Returns: An Object representing the evaluated result
// Examples:
// - IntegerLiteral(5) -> Integer{5}
// - ExpressionStmt(1 + 2) -> Integer{3}
// - Program([let x = 5]) -> Integer{5}
func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node)
	case *ast.IfExpression:
		return evalIfExpression(node)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue)
		return &object.ReturnValue{Value: val}
	}

	return nil
}

// Evaluates a sequence of statements
// It evaluates each statement in order and returns the result of the last one
// This implements the behavior where a sequence of statements returns the value of it last expression
// Parameters:
// - stmts: Slice of statements to evaluate
// Returns: The result of the last statement, or nil for empty sequences
// Examples:
// - [5] -> Integer{5}
// - [let x = 5, x + 3] -> Integer{8}
func evalStatements(stmts []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement)

		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
	}
	return result
}

// Converts Go's native boolean values to Gibbon boolean objects
// This function implements the singleton pattern for boolean values, which:
// 1. Reduces memory allocation by reusing the same true/false objects
// 2. Enables pointer comparison instead instead of value comparison
// 3. Improves garbage collection performance
// Performance benefits:
// - Memory: Only two boolean objects are ever created
// - CPU: Pointer comparison is faster than value comparison
// - GC: Less pressure on garbage collector due to object reuse
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}

	return FALSE
}

// Evaluates unary operator expressions
// Currently supports:
// - Logical NOT (!): Inverts boolean values
// - Negation (-): Negates numeric values
// Examples:
// - !true -> FALSE
// - !false -> TRUE
// - !null -> TRUE
// - -5 -> Integer{-5} (TODO)
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return NULL
	}
}

// Implements logical NOT (!) operator. This implements Javascript-style truthiness where:
// - true -> false
// - false -> true
// - null -> true
// - everything else -> false
func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return NULL
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	default:
		return NULL
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return NULL
	}
}

func evalIfExpression(ie *ast.IfExpression) object.Object {
	condition := Eval(ie.Condition)

	if isTruthy(condition) {
		return Eval(ie.Consequence)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative)
	} else {
		return NULL
	}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalProgram(program *ast.Program) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement)

		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement)

		if result != nil && result.Type() == object.RETURN_VALUE_OBJ {
			return result
		}
	}

	return result
}
