package compiler

// Represents the scope level of a symbol
// Currently supports:
// - Global scope (top-level variables)
type SymbolScope string

// Available scope levels
const (
	GlobalScope SymbolScope = "GLOBAL"
)

// Represents a variable or identifier in the source code
type Symbol struct {
	Name  string      // The identifier string
	Scope SymbolScope // The scope level (global, local, etc)
	Index int         // Unique index for bytecode generation
}

// Implements a symbol lookup and tracking system
type SymbolTable struct {
	store          map[string]Symbol // Maps names to their Symbol definitions
	numDefinitions int               // Counter for assigning unique indices
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

// Adds a new symbol to the table
// Effects:
// - Creates new Symbol with next available index
// - Adds to symbol store
// - Increments definition counter
func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: s.numDefinitions, Scope: GlobalScope}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

// Looks up a symbol by name
// Parameters:
// - name: The identifier to look up
// Returns:
// - Symbol: The found symbol
// - bool: true if found, false if not defined
// Used by compiler to:
// - Generate correct bytecode for variables
// - Check for undefined variables
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	return obj, ok
}
