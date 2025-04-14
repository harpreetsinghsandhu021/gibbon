package repl

import (
	"bufio"
	"fmt"
	"gibbon-lang/src/gibbon/lexer"
	"gibbon-lang/src/gibbon/token"
	"io"
)

const PROMPT = ">> "

// Reads from the input source until encountering a newline, take the just read line and pass it to an instance
// of the lexer and finally print all the tokens the lexer gives us until we encounter EOF.
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.NewLexer(line)
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Printf("%+v\n", tok)
		}
	}
}
