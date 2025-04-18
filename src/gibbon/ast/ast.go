package ast

import (
	"bytes"
	"gibbon-lang/src/gibbon/token"
	"strings"
)

// Base interface for all AST nodes.
// Every node in the AST must be able to return its literal token value
type Node interface {
	TokenLiteral() string
	String() string
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

// Creates a buffer and writes the return value of each statements String() method to it.
// And returns the buffer as a string.
func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

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

// Produces a string representation of the let statement
// Format: "let <identifier> = <expression>;"
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ") // Write "let "
	out.WriteString(ls.Name.String())        // Write identifier
	out.WriteString(" = ")                   // Write " = "

	if ls.Value != nil {
		out.WriteString(ls.Value.String()) // Write expression
	}
	out.WriteString(";")

	return out.String()
}

// Represents a 'return' statement in the AST.
// It consists of the 'return' token and an optional return value expression.
// For example: 'return 5' or 'return x + y'
type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// Produces a string representation of the return statement
// Format: "return <expression>;"
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ") // Write "return"

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String()) // Write expression if present
	}

	out.WriteString(";")

	return out.String()
}

// Represents a statement that consists of a single expression.
// It is used when an expression appears in a statement context, such as a function
// call appearing on its own line. The Token field holds the first token of the
// expression and Expression holds the actual expression being used as a statement.
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// Returns the string representation of the contained expression
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}

	return ""
}

type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}

	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// EXPRESSIONS

// Represents a named entity in the program.
// Examples: variable names, function names.
type Identifier struct {
	Token token.Token // Identifier token
	Value string      // Identifier name as string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// Returns the identifier's name
func (i *Identifier) String() string {
	return i.Value
}

// Represents an integer literal node in the AST.
// It consists of the token containing location and literal text information,
// and the actual parsed int64 value.
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// Represents a prefix operator expression in the AST.
// It consists of an operator token (e.g., "!", "-"), the operator as a string,
// and the expression on which the operator acts (Right).
// For example, in the expression "!true" or "-5":
// - Token would store the token information
// - Operator would be "!" or "-"
// - Right would be the expression being operated on
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

// Returns a string representation of the prefix expression in the format "(<operator><right_expr>)".
// For example, if the operator is "-" and right expression is "5", it returns "(-5)".
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")

	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

// Represents a binary operator expression with a left and right operand.
// For example: a + b, x * y, etc.
// The Token field contains the first token of the expression.
// Left holds the expression on the left side of the operator.
// Operator is the string value of the operator (e.g., "+", "-", "*", "/").
// Right holds the expression on the right side of the operator.
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}

// Represents a boolean literal in the AST.
// It holds both the token information and the actual boolean value.
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

// Represents an if expression node in the AST.
// It consists of a condition expression that evaluates to a boolean value,
// a consequence block that executes if the condition is true,
// and an optional alternative block that executes if the condition is false.
// The Token field stores the lexical token for the 'if' keyword.
type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

// Represents a block of statements enclosed in curly braces.
// It consists of a '{' token and a slice of statements that are executed
// sequentially within the block scope.
type BlockStatement struct {
	Token      token.Token // The '{' token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// Represents a function expression in the source code.
// It consists of:
// - Token: The 'fn' token that starts the function literal
// - Parameters: A slice of identifiers representing the function parameters
// - Body: A block statement containing the function's implementation
type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

// Represents a function call in the AST.
// It consists of a function identifier or literal followed by arguments in parentheses.
// For example: add(1, 2) or fn(x,y){x+y}(1,2)
type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}

	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type HashLiteral struct {
	Token token.Token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
