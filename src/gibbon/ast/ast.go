package ast

import "gibbon-lang/src/gibbon/token"

// Base interface for all AST nodes.
// Every node in the AST must be able to return its literal token value
type Node interface {
	TokenLiteral() string
}

// Represents a statement node in the AST.
// Statements are language constructs that perform actions but dont't return values.
// Examples: variable declarations, return statements, etc.
type Statement interface {
	Node
	statementNode()
}

// Represents an expression node in the AST.
// Expression are language constructs that can be evaluated to produce a value
// Examples: arithmetic operations, function calls, etc.
type Expression interface {
	Node
	expressionNode()
}

// Root node of every AST.
// Contains a slice of all statements in the program.
type Program struct {
	Statements []Statement
}

// Returns the literal value of the first token in the program.
// If the program has no statements, it returns an empty string.
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

// STATEMENTS

// Represents a variable declaration statement.
// Example: let x = 5
type LetStatement struct {
	Token token.Token // Holds the 'let' token
	Name  *Identifier // The identifier being declared
	Value Expression  // The expression being assigned to the identifier
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

// Represents a 'return' statement in the AST.
// It consists of the 'return' token and an optional return value expression.
// For example: 'return 5' or 'return x + y'
type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// EXPRESSIONS

// Represents a named entity in the program.
// Examples: variable names, function names.
type Identifier struct {
	Token token.Token // Identifier token
	Value string      // Identifier name as string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
