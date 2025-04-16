package object

// Implements a hierarchical symbol table for variable scoping
// It provides:
// - Variable storage and lookup
// - Lexical scoping through linked enviroments
// - Support for closure in functions
type Enviroment struct {
	store map[string]Object
	outer *Enviroment
}

// Creates a new top-level enviroment
// This is used for:
// - Global scope
// - Initial REPL enviroment
// Returns: An empty enviroment with no outer scope
func NewEnviroment() *Enviroment {
	s := make(map[string]Object)
	return &Enviroment{store: s, outer: nil}
}

// Creates a new enviroment that extends an outer one.
// This implements lexical scoping by:
// - Creating a new enviroment for local scope
// - Linking it to its enclosing scope
// Used for:
// - Function bodies
// - Block statements
// Parameters:
// - outer: The enclosing scope's enviroment
func NewEnclosedEnviroment(outer *Enviroment) *Enviroment {
	env := NewEnviroment()
	env.outer = outer

	return env
}

// Looks up and identifier in the enviroment
// It implements lexical scoping by:
// 1. Looking in the current scope first
// 2. If not found, recursively checking outer scopes
// Parameters:
// - name: The identifier to look up
// Returns:
// - The bound value and true if found
// - nil and false it not found in any scope
func (e *Enviroment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Binds a value to an identifier in the current scope
// This implements:
// - Variable declaration
// - Variable assignment
// Returns: The bound value
func (e *Enviroment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
