package evaluator

import (
	"gibbon-lang/src/gibbon/ast"
	"gibbon-lang/src/gibbon/object"
)

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
		return evalStatements(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
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
	}
	return result
}
