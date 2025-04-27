package compiler

// Represents the scope level of a symbol
// Currently supports:
// - Global scope (top-level variables)
type SymbolScope string

// Available scope levels
const (
	GlobalScope   SymbolScope = "GLOBAL"
	LocalScope    SymbolScope = "LOCAL"
	BuiltinScope  SymbolScope = "BUILTIN"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTIOn"
)

// Represents a variable or identifier in the source code
type Symbol struct {
	Name  string      // The identifier string
	Scope SymbolScope // The scope level (global, local, etc)
	Index int         // Unique index for bytecode generation
}

// Implements a symbol lookup and tracking system
type SymbolTable struct {
	Outer          *SymbolTable
	store          map[string]Symbol // Maps names to their Symbol definitions
	numDefinitions int               // Counter for assigning unique indices
	FreeSymbols    []Symbol
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}
	return &SymbolTable{store: s, FreeSymbols: free}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

// Adds a new symbol to the table
// Effects:
// - Creates new Symbol with next available index
// - Adds to symbol store
// - Increments definition counter
func (s *SymbolTable) Define(name string) Symbol {
	symbol := Symbol{Name: name, Index: s.numDefinitions}

	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	s.store[name] = symbol
	return symbol
}

// Creates a new free variable symbol based on an original symbol from an outer scope.
// This is used for handling closures, where inner functions need to access variables from
// enclosing scopes.
func (s *SymbolTable) defineFree(original Symbol) Symbol {
	// Add the original symbol to the list of free symbols
	s.FreeSymbols = append(s.FreeSymbols, original)

	// Create new symbol for the free variable
	// Index corresponds to position in freeSymbols
	symbol := Symbol{Name: original.Name, Index: len(s.FreeSymbols) - 1}
	symbol.Scope = FreeScope

	s.store[original.Name] = symbol

	return symbol
}

// Creates a new symbol for a function definition in the current scope.
func (s *SymbolTable) DefineFunctionName(name string) Symbol {
	symbol := Symbol{Name: name, Index: 0, Scope: FunctionScope}
	s.store[name] = symbol
	return symbol
}

// Looks up for a symbol in the current and outer symbol tables.
// This implements lexical scoping by searching in the current scope first, then checking
// the outer scopes if the symbol is'nt found.
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	// Try to find symbol in current scope's store
	obj, ok := s.store[name]

	// If not found in current scope and we have an outer scope
	if !ok && s.Outer != nil {
		// Recursively check outer scope
		obj, ok := s.Outer.Resolve(name)
		if !ok {
			return obj, ok
		}

		if obj.Scope == GlobalScope || obj.Scope == BuiltinScope {
			return obj, ok
		}

		// If found, registers it as free variable
		free := s.defineFree(obj)
		return free, true
	}

	return obj, ok
}
