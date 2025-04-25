package vm

import (
	"gibbon-lang/src/gibbon/code"
	"gibbon-lang/src/gibbon/object"
)

// Represents a function call frame in the VM
// Used for:
// - Managing function execution state
// - Supporting nested function calls
// - Tracking return positions
type Frame struct {
	fn *object.CompiledFunction //  Reference to compiled function
	ip int                      // Instruction pointer for this frame
}

func NewFrame(fn *object.CompiledFunction) *Frame {
	return &Frame{fn: fn, ip: -1}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
