package evaluator

import (
	"fmt"
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
func Eval(node ast.Node, env *object.Enviroment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
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
func evalStatements(stmts []ast.Statement, env *object.Enviroment) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement, env)

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
		return newError("unknown operator: %s%s", operator, right.Type())
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

// Handles numeric negation (unary minus)
// Only works with integers, returns an error other types
// Parameters:
// - right: The operadn to negate (must be an integer)
// Returns: Negated integer or error object
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

// Handles binary operator expressions
// Supports:
// - Arithmetic operations bw integers
// - Equality comparisons bw any types
// - Comparisons bw integers
// Parameters:
// - operator: The binary operator ("+", "-", "*", "/", "==", "!=", "<", ">")
// - left, right: The operands
// Returns: Result object or error for invalid operations
func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// Handles binary operations bw integers
// Supports:
// - Arithmetic: +, -, *, /
// - Comparison: <, >, ==, !=
// Parameters;
// - operator: The binary operator
// - left, right: The integer operands
// Returns: Result or error for unknown operatos
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// Handles conditional expressions
// Evaluates the condition and returns either:
// - The consequence if condition is truthy
// - The alternative if condition is falsy
// - NULL if condition is falsy and no alternative exists
func evalIfExpression(ie *ast.IfExpression, env *object.Enviroment) object.Object {
	condition := Eval(ie.Condition, env)

	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

// Implements Javascript-style truthiness
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

// Evaluates a complete program AST
// Handles special cases:
// - Return statements: Unwraps return value
// - Errors: Stops execution and returns error
// Returns the last evaluated expression
func evalProgram(program *ast.Program, env *object.Enviroment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

// Evaluates a sequence of statements in a block
func evalBlockStatement(block *ast.BlockStatement, env *object.Enviroment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()

			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

// Evaluates identifier references
// Looks up the identifier in the current enviroment
// Returns the bound value if found, error if identifier is not defined
func evalIdentifier(node *ast.Identifier, env *object.Enviroment) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: " + node.Value)
	}
	return val
}

// Evaluates a list of expressions
func evalExpressions(exps []ast.Expression, env *object.Enviroment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

// Implements function application (calling)
// It handles:
// 1. Type checking the function object
// 2. Creating a new enviroment for the function execution
// 3. Evaluating the function body
// 4. Unwrapping any return values
// Returns:
// - The result of evaluating the function body
// - Error if fn is not a callable object
func applyFunction(fn object.Object, args []object.Object) object.Object {
	// Type check - ensure we have a callable function
	function, ok := fn.(*object.Function)
	if !ok {
		return newError("not a function: %s", fn.Type())
	}

	// Create new scope and evaluate function bodt
	extendedEnv := extendFunctionEnv(function, args)
	evaluated := Eval(function.Body, extendedEnv)

	// Handle return values
	return unwrapReturnValue(evaluated)
}

// Creates a new enclosed enviroment for function execution
// This implements lexical scoping by:
// 1. Creating a new enviroment that extends the function's closure
// 2. Binding parameters to argument values in the new scope
// Parameters:
// - fn: The function whose enviroment to extend
// - args: The argument values to bind to parameters
// Returns:
// A new enviroment containing parameter bindings
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Enviroment {
	env := object.NewEnclosedEnviroment(fn.Env)

	// Bind each parameter to its corresponding argument
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}

// Handles return value processing
// If the evaluated expression was a return value:
// - Extracts the actual value from the return value:
// Otherwise:
// - Returns the original value inchanged
func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

// Creates a new error object with formatted message
// Used throughout the evaluator to report runtime errors
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// Checks if an object is an error
// Used to short-circuit evaluation when errors occur
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}

	return false
}
