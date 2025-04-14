package parser

import (
	"fmt"
	"gibbon-lang/src/gibbon/ast"
	"gibbon-lang/src/gibbon/lexer"
	"gibbon-lang/src/gibbon/token"
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

	// Skip expression=s until we encounter a semicolon
	for !p.currTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.currToken}

	p.nextToken()

	for !p.currTokenIs(token.SEMICOLON) {
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
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// Get the prefix parsing function for current type
	prefix := p.prefixParseFns[p.currToken.Type]
	if prefix == nil {
		return nil
	}
	// parse the prefix expression
	leftExp := prefix()
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

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
