package code

import (
	"bytes"
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
const (
	OpConstant Opcode = iota // Uses iota for auto-incrementing opcode values
	OpAdd
	OpPop
	OpSub
	OpMul
	OpDiv
)

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
	OpAdd:      {"OpAdd", []int{}},       // A single byte, single opcode that pops the two topmost elements off the stack, adds them and pushes the result back on the stack
	OpPop:      {"OpPop", []int{}},       // Pop the topmost element off the stack
	OpSub:      {"OpSub", []int{}},
	OpMul:      {"OpMul", []int{}},
	OpDiv:      {"OpDiv", []int{}},
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

/*
Implements the Stringer interface for Instructions type.
It provides a human-readable representation of bytecode instructions.

Detailed Operation:
------------------

1. Instruction Processing:
  - Iterates through instructions sequentially
  - Decodes each instruction and its operands
  - Formats them into readable text

2. Error Handling:
  - Reports invalid opcodes
  - Validates operand counts
  - Continues processing after errors

3. Output Formatting:
  - Includes instruction offset
  - Shows opcode names
  - Displays decoded operand values

Memory Layout Example:
---------------------
Given bytecode: [OpConstant, 0x01, 0x00]
Output: "0000 OpConstant 256"

	        |____||________| |__|
			  |       |        |
			  |       |        +--- Decoded operand value
			  |       +--- Opcode name
			  +-- Instruction offset
*/
func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		// Lookup instruction definition
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		// Read and decode operands
		operands, read := ReadOperands(def, ins[i+1:])

		// Format instruction with offset and decoded values
		if i > 0 {
			fmt.Fprintf(&out, "\t") // Add tab for all lines except the first
		}
		fmt.Fprintf(&out, "%04d %s", i, ins.fmtInstruction(def, operands))

		// Add newline if not the last instruction
		if i+1+read < len(ins) {
			fmt.Fprintf(&out, "\n")
		}

		// Move to next instruction
		i += 1 + read
	}

	return out.String()
}

/*
Formats a single instruction with its operands.

Detailed Operation:
------------------
1. Validation:
  - Checks operand count matches definition
  - Reports mismatch errors

2. Formatting:
  - Handles different operand counts
  - Combines opcode name with operands
  - Provides error messages for invalid cases
*/
func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}

	// Format based on operand count
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		// Format single-operand instruction
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

/*
Decodes operands from a bytecode instruction.

Detailed Operation:
------------------
1. Operand Extraction Phase:
  - Creates a slice to hold decoded operands
  - Size based on operand width specification
  - Maintains offset for multi-operand instructions

2. Width-based Decoding:
  - Handles different operand widths (currently 2-bytes)
  - Uses big-endian byte order for consistency
  - Converts bytes back to integers

3. Offset Tracking:
  - Keeps track of bytes read
  - Essential for instruction parsing
  - Enables sequential instruction reading

Parameters:
  - def *Definition    : Metadata about the instruction's operands
  - ins Instructions  : Raw bytecode containing the operands

Returns:
  - []int: Slice of decoded operand values
  - int:   Number of bytes read (offset)

Memory Layout Example:
---------------------
For OpConstant with operand 256
[0x01][0x00]    -> 256

	^     ^
	|     |
	+-----+-- 2 bytes in big-endian format
*/
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	// Allocate slice for decoded operands
	// Size matches the number of expected operands
	operands := make([]int, len(def.OperandWidths))
	// Track position in instruction byte sequence
	offset := 0

	// Process each operand according to its width
	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			// Handle 2-byte operands:
			// - Read using big-endian byte order
			// - Convert to integer value
			operands[i] = int(ReadUint16(ins[offset:]))
		}

		// Move offset forward by processes width
		offset += width
	}

	return operands, offset
}

/*
Converts two bytes into a uint16 value.

Operation:
---------
- Takes 2 consecutive bytes from instruction
- Interprets them as big-endian unint16
- Essential for reading 2-byte operands
*/
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
