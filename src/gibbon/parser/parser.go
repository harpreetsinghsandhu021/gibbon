package parser

import (
	"fmt"
	"gibbon-lang/src/gibbon/ast"
	"gibbon-lang/src/gibbon/lexer"
	"gibbon-lang/src/gibbon/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

// Maps token types to their operator precedence levels.
// Higher values indicate higher precedence in the order of operations.
// The precedence levels are:
//   - PRODUCT (multiplication, division)
//   - SUM (addition, subtraction)
//   - LESSGREATER (comparison operators <, >)
//   - EQUALS (equality operators ==, !=)
var precendences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

// Implements a recursive descent parser for the language.
// A recursive descent parser is a top-down parser that uses a set of mutually
// recursive functions to process the grammar rules.
// This implementation uses the following techniques:
// - Lookahead: Maintains both current and peek tokens for parsing decisions
// - Predictive parsing: Uses the current token to determine which parsing function to call
// - No backtracking: Makes parsing decisions based on the current token context
type Parser struct {
	l         *lexer.Lexer
	currToken token.Token // Current token under examination
	peekToken token.Token // Next token in the stream(lookahead token)
	errors    []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parsedGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	// Read the next two tokens, so the currToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// Advances the parser's token window by one position.
func (p *Parser) nextToken() {
	p.currToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// Implements the top-level parsing function that processes the entire program.
// Following recursive descent principles, it:
// 1. Creates the root AST node
// 2. Repeatedly parses statements until EOF
// 3. Builds the program AST bottom-up
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// Parse statements until we reach the end of input
	// This forms the root of our AST
	for p.currToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

// A Dispatcher function in the recursive descent process.
// It examines the current token to determine which type of statement to parse.
// This implements the predictice parsing aspect of recursive descent.
func (p *Parser) parseStatement() ast.Statement {
	switch p.currToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// Parses let statements according to this grammar rule:
// let_statement -> 'let' identifier '=' expression ';'
// This is a recursive descent parsing function that:
// 1. Verifies the sequence of tokens matches the expected pattern
// 2. Constructs an AST node representing the let statement.
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.currToken}

	// Expect and consume an identifier token
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// Create the identifier node for the variable name
	stmt.Name = &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}

	// Expect and consume the assignment operaot
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	// Skip expression=s until we encounter a semicolon
	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.currToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// Handles expressions that appear as statements
// Examples:
// - function calls: someFunction();
// - assignments: x = 5;
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.currToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// The core of the Pratt parsing implementation. It uses precendence climbing to handle
// operator precendence.
// Parameters:
// - precedence: the current precedence level being parsed
// Returns: An AST expression node or nil if parsing fails
// Examples of expressions it can parse:
// - Simple: 5, x, true
// - Prefix: -5, !true
// - Infix: 5 + 3, x * y, a == b
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// Get the prefix parsing function for current type
	prefix := p.prefixParseFns[p.currToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.currToken.Type)
		return nil
	}
	// parse the prefix expression
	leftExp := prefix()

	// Continue parsing infix expressions while:
	// 1. The next token is'nt a semicolon
	// 2. The next operator has higher precedence than current
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

// Checks if the current token matches the expected type
func (p *Parser) currTokenIs(t token.TokenType) bool {
	return p.currToken.Type == t
}

// Provides lookahead capability by checking the next token
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// Implements predictive parsing by verifying the next token matches the expected grammar
// rule before proceeding.
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// Returns all parsing errors encountered during parsing
// These errors typically indicate syntax errors in the source code
func (p *Parser) Errors() []string {
	return p.errors
}

// Records an error when the next token does'nt match what's expected by the grammar.
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token ton be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
}

// Parses integer literals in the source code
// This function:
// 1. Creates an AST node for the integer literal
// 2. Converts the string representation to an int64
// 3. Reports an error if the conversion fails
func (p *Parser) parseIntegerLiteral() ast.Expression {
	// Create new AST node with current token
	lit := &ast.IntegerLiteral{Token: p.currToken}

	// Convert string to in64, using base 0 for automatic base detection
	// This allows parsing decimal, octal and hex literals
	val, err := strconv.ParseInt(p.currToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.currToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = val
	return lit
}

// Handles unary operator expressions
// Grammar: prefix_expression -> prefix_operator expression
// Examples:
// - Negation: -5, -foo
// - Logical NOT: !true, !isSomething
func (p *Parser) parsePrefixExpression() ast.Expression {
	// Create prefix expression node with current token
	expression := &ast.PrefixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Literal,
	}
	// Advance to the operand token
	p.nextToken()
	// Parse the operand with PREFIX precedence
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.currToken,
		Operator: p.currToken.Literal,
		Left:     left,
	}

	precedence := p.currPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.currToken, Value: p.currTokenIs(token.TRUE)}
}

// Handles parsing of function calls.
// Grammar: call_expression -> expression '(' argument_list? ')'
// The expression can be an identifier or a function literal.
// Examples:
// - Simple call: add(1, 2)
// - Nested calls: add(subtract(5, 3), multiply(2, 4))
// - Function literal call: fn(x,y){x+y}(1,2)
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	// Create new call expression node with current '(' token the function being called
	exp := &ast.CallExpression{Token: p.currToken, Function: function}
	// Parse the argument list b/w the parenthesis
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

// Handles parsing of function call arguments
// Grammar: argument_list -> expression (',' expression)*
// Parameters are evaluated left to right.
// Examples:
// - No arguments: foo()
// - Single argument: foo(5)
// - Multiple arguments: foo(1, x + y, bar())
// - Nested expressions: foo(1 + 2, bar(3))
func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	// Handle empty argument list: foo()
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	// Parse the first argument
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	// Parse additional arguments: foo(1, 2, 3)
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume the comma
		p.nextToken() // move to the next argument
		args = append(args, p.parseExpression(LOWEST))
	}

	// Expect and consume closing parenthesis
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

// Handles parsing of parenthesized expressions
// Grammar: grouped_expression -> '(' expression ')'
// This allows for explicit precedence control in expressions.
// Examples:
// - Simple grouping: (5 + 3)
// - Nested grouping: ((5 + 3) * 2)
// - Mixed operators: (a + b) * (c + d)
func (p *Parser) parsedGroupedExpression() ast.Expression {
	// Move past the opening parenthesis
	p.nextToken()

	// Parse the expression inside the parenthesis
	// Using LOWEST precedence to allow any expression type
	exp := p.parseExpression(LOWEST)

	// Ensure the next token is a closing parenthesis
	// Return nil if there's a syntax error (missing closing parenthesis)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// Handles parsing of conditional expressions
// Grammar: if_expression -> 'If' '(' expression ')' block_statement ('else' block_statement)?
// This implements conditionl branching in the language.
// / Examples:
// - Simple if: if (x > 5) { return true; }
// - If with else: if (x > 5) { return true; } else { return false; }
func (p *Parser) parseIfExpression() ast.Expression {
	// Create new if expression mode with current 'if' token
	expression := &ast.IfExpression{Token: p.currToken}

	// Expect and consume opening parenthesis for condition
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Move past the opening parenthesis and parse the condition
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	// Expect and consume closing parenthesis
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// Expect and consume opening brace for consequence block
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Parse the consequence (true branch) block
	expression.Consequence = p.parseBlockStatement()

	// Check if there is an else block along and consume it the same way as if block
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

// Handles parsing of function expressions
// Grammar: fn_literal -> 'fn' '(' parameter_list? ')' block_statement
// This implements function definitions in the language
// Examples:
// - Empty function: fn() { }
// - Single parameter: fn(x) { return x; }
// - Multiple parameters: fn(x, y) { return x + y; }
func (p *Parser) parseFunctionLiteral() ast.Expression {
	// Create new function literal node with current 'fn' token
	lit := &ast.FunctionLiteral{Token: p.currToken}

	// Expect and consume opening parenthesis for parameters
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Parse function parameters
	lit.Parameters = p.parseFunctionParameters()

	// Expect and consume opening brace for function body
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Parse the function body as a block statement
	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.currToken, Value: p.currToken.Literal}
}

// Handles parsing of array literal expression
// This implements array construction in the language
// Examples:
// - Empty array: []
// - Simple values: [1, 2, 3]
// - Mixed expressions: [1 + 2, foo(), x * y]
// - Nested arrays: [[], [1, 2], [3]]
// Key features:
// - Elements can be any valid expression
// - Supports comma-separated list of expressions
// - Handles nested array literals
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.currToken}

	array.Elements = p.parseExpressionList(token.RBRACKET)

	return array
}

// Handles parsing of comma-seperated expression lists
// Parameters:
// - end: The token type that terminates the list (e.g "]" for arrays)
// Returns:
// - Slice of parsed expressions
// - nil if parsing fails (missing terminator)
// Parsing process:
// 1. Handle empty list case
// 2. Parse first expression
// 3. Parse addtional comma-seperated expressions
// 4. Validate proper termination
func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	// Handle empty list case
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	// Parse first expression
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	// Parse additional expressions after commas
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume comma
		p.nextToken() // move to next expression
		list = append(list, p.parseExpression(LOWEST))
	}

	// Ensure proper termination
	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// Handles parsing of function parameter lists
// Grammar: parameter_list -> IDENT (',', IDENT)*
// Examples:
// - fn(): []
// - fn(x): [x]
// - fn(x, y, z): [x, y, z]
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	// Handle empty parameter list: fn()
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	// Parse firddt parameter
	p.nextToken()

	ident := &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
	identifiers = append(identifiers, ident)

	// Parse additional parameters: fn(x, y, z)
	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // consume the comma
		p.nextToken() // move to the next identifier
		ident := &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
		identifiers = append(identifiers, ident)
	}

	// Expect and consume closing parenthesis
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

// Handles parsing of block statements (code blocks)
// Grammar: block_statement -> '{' statement* '}'
// A block statement is a sequence of statements enclosed in curly braces used in:
// - Function bodies
// - If/else branches
// - Loop bodies
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.currToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.currTokenIs(token.RBRACE) && !p.currTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// Returns the precedence associated with token type of p.peekToken. If it
// does'nt find a precedence for p.peekToken, it default to LOWEST, the lowest
// possible precedence any operator can have.
func (p *Parser) peekPrecedence() int {
	if p, ok := precendences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// Returns the precedence associated with token type of p.currToken. If it
// does'nt find a precedence for p.currToken, it default to LOWEST, the lowest
// possible precedence any operator can have.
func (p *Parser) currPrecedence() int {
	if p, ok := precendences[p.currToken.Type]; ok {
		return p
	}

	return LOWEST
}

// Records an error when no prefix parse function exists
// for a given token type.
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
