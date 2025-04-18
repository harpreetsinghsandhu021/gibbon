package code

import (
	"encoding/binary"
	"fmt"
)

/*
Bytecode Implementation for the Gibbon Programming language
----------------------------------------------------------

This file implements the core bytecode infrastructure that powers Gibbon's virtual machine.
Bytecode is a low-level representation of program instructions that can be efficiently executed by the virtual machine.

Key Concepts:
------------
1. Instructions: A sequence of bytes where each instruction consists of:
   - An opcode (1 byte)
   - Zero or more operands (variable bytes)

2. Opcodes: Single-byte values that identify the type of operation
   - Each opcode has a unique value
   - Opcodes are the first byte in any instruction

3. Operands: Additional data that opcodes may need
   - Number and size of operands varies by opcode
   - Operands follow immeditely after their opcode

4. Constant pool:
   - Storage for constant values (numbers, strings, etc.)
   - Constants are referenced by index in instructions
   - Reduces bytecode size by avoiding direct value embedding
*/

// Represents a sequence of bytecode instructions.
// It's implemented as a byte slice for direct memory access and manipulation
type Instructions []byte

// Represents a single byte operation code
// Each opcode identifies a specific operation for the virtual machine to perform
type Opcode byte

// The operation code for loading constants.
// When executed, it loads a constant from the constant pool onto the stack.
// The operand (2 bytes) specifies the index of the constant in the pool.
const OpConstant Opcode = iota // Uses iota for auto-incrementing opcode values

// Provides metadata about an opcode
// This structure is crucial for debugging and correctly parsing instructions.
type Definition struct {
	Name          string // Human-readable name of the opcode (e.g, "OpConstant")
	OperandWidths []int  //Specifies how many bytes each operand uses, for e.g: []int{2} means one operand that is 2 bytes wide
}

// Maps opcodes to their metadata
// This allows us to look up information about an opcode at runtime
var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}}, // OpConstant has one 2-byte operand
}

// Retrieves the definition for a given opcode.
// This is used during instruction parsing and debugging.
// Returns an error if the opcode is not recognized
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

/*
	   Creates a new bytecode instruction by combining an opcode with its operands.
	   It handles the binary encoding of operands and ensures proper byte alignment.
	   Parameters:
	   - op: The opcode for the instruction (e.g., OpConstant)
	   - operands: Variable number of integer operands that the instruction requires

	   Detailed Operation:
	   ------------------
	   1. Validation Phase:
	      - Verifies the opcode exists in the definitions table
		  - Returns empty slice for invalid opcodes to prevent undefined behavior

	   2. Size calculation Phase:
	      - Allocates spaces for opcode (1 byte)
		  - Adds space for operand based on its width
		  - Ensures proper memory alignment

	   3. Encoding Phase:
	      - First byte: stores the opcode
		  - Subsequent bytes: stores operands in big-endian format
		  - Maintains offset pointer for proper byte positioning

		Memory Layout Example:
		---------------------
		For OpConstant with operand 256
		[OpConstant][0x01][0x00]
		    ^         ^     ^
			|         |     |
			|         +-----+-- 2 bytes for operand (256 in big-endian)
			+-- 1 byte for opcode

		Returns:
		- []byte: Fully encoded instruction ready for VM execution
*/
func Make(op Opcode, operands ...int) []byte {
	// Lookup the opcode definition to get metadata about operand sizes
	def, ok := definitions[op]

	if !ok {
		// Return empty byte slice if opcode is not recognized
		// This prevents undefined behavior in the VM
		return []byte{}
	}

	// Calculate total instruction length:
	// 1 byte for opcode + sum of all operand widths
	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	// Allocate a byte slice for the complete instruction
	// This eliminates need for later resizing
	instruction := make([]byte, instructionLen)

	// Write the opcode as the first byte
	instruction[0] = byte(op)

	// Keeps track of where we are in the instruction while writing operands
	offset := 1

	// Encode each operand according to its width specification
	for i, o := range operands {
		// Get the byte width for this specific operand
		width := def.OperandWidths[i]

		// Encode operand based on its width specification
		switch width {
		case 2:
			// For 2-byte operands, encode as big-endian uint16
			// This allows values from o to 65535
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		// Move offset forward by width bytes
		offset += width
	}

	return instruction
}
