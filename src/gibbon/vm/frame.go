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
	cl          *object.Closure //  Reference to compiled function
	ip          int             // Instruction pointer for this frame
	basePointer int             // Points to the bottom of the stack of the current call frame
}

func NewFrame(cl *object.Closure, basePointer int) *Frame {
	return &Frame{cl: cl, ip: -1, basePointer: basePointer}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
