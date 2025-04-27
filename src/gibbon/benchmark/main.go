package main

import (
	"flag"
	"fmt"
	"gibbon-lang/src/gibbon/compiler"
	"gibbon-lang/src/gibbon/evaluator"
	"gibbon-lang/src/gibbon/lexer"
	"gibbon-lang/src/gibbon/object"
	"gibbon-lang/src/gibbon/parser"
	"gibbon-lang/src/gibbon/vm"
	"time"
)

var engine = flag.String("engine", "vm", "use 'vm' or 'eval'")

var input = `
let fibonacci = fn(x) {
if (x == 0) {
0
} else {
if (x == 1) {
return 1;
} else {
fibonacci(x - 1) + fibonacci(x - 2);
}
}
};
fibonacci(35);
`

func main() {
	flag.Parse()

	var duration time.Duration
	var result object.Object

	l := lexer.NewLexer(input)

	p := parser.NewParser(l)
	program := p.ParseProgram()

	if *engine == "vm" {
		comp := compiler.New()
		err := comp.Compile(program)

		if err != nil {
			fmt.Printf("compiler error: %s", err)
		}

		machine := vm.New(comp.Bytecode())

		start := time.Now()

		err = machine.Run()

		if err != nil {
			fmt.Printf("vm error: %s", err)
			return
		}

		duration = time.Since(start)
		result = machine.LastPoppedStackElem()
	} else {
		env := object.NewEnviroment()
		start := time.Now()
		result = evaluator.Eval(program, env)
		duration = time.Since(start)
	}

	fmt.Printf("engine=%s, result=%s, duration=%s\n", *engine, result.Inspect(), duration)

}
