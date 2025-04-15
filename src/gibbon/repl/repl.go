package repl

import (
	"bufio"
	"fmt"
	"gibbon-lang/src/gibbon/lexer"
	"gibbon-lang/src/gibbon/parser"
	"io"
)

const PROMPT = ">> "

const MONKEY_FACE = `                                                  
                          ██                     
                       ████    █                 
                    ▓▓█▓████████                 
                 █▓▓▓▓████                       
                  ███▓▓                          
                 █▓▓▓▓                           
                 ▓▓▓▓▓                           
                ▓▓▓▓▓▓                           
               ▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓                  
              █▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓               
              ▓▓▓▓▓▓▓▓▒▒▒▒▒▒▒▒▒▒▓▓▓▓             
              ▓▓▓▓▒▓▓▒▒▒▒▓▒▓▒▒▒▒▓▒▓▓▓▓▓          
               ▓▓▓▓▓▓▒██▓▓▒▓▓█▓▒▒██▓▓▓▓          
               ▓▓▓▓▓▓▓▒▒▓▓▓▓▒▓▒▒▓▓▓▓▓▓▓          
                ▓▓▓▓▒▒▓▓▓▓▓▓▓▓▒▒▒▓▓▓▓▓▓          
                 ███▓▒▓▓▓▓▓▓▓▓▒▒▒▓▓█             
                  ██▓▓▓▓▓▓▓▓▓▒▓▓▓▓▓▓█            
                    █▓▓███████▓▓▓▓▓▓▓█           
                    █▓▓▓▓█▓▓▓▓▓▓▓▓▓▓▓▓           
                   ██▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓          
                   ██▓▓█▓▓▓▓▓▓▓█▓▓▓▓▓▓▓          
                 ████▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓          
              █▓▓▓▓▓▓▓█▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓          
             ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓█          
             █▓▓▓▓▓▓▓▓▓▓▓▒▓▓▓▓▓▓▓▓▓▓▓██          
             ███▓▓▓▓▓▓▓▓▒▒▓█▓▓▓▓▓▓▓███           
             ████▓█▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓██            
                 █▓▓████▓▓▓▓▓▓▓▓▓▓██             
                  █▓▓▓█████▓▓▓▓▓▓█               
                  █▓▓▓▓ ▓▓▓▓▓████                
                  ▓▒▒█▓▓▒▓▓█                     
                  ▓▓▓▓█▓▒▒▓                      
                      █▓▓▓                                                                                    
`

const WELCOME_IMAGE = `
██████  ██ ██████  ██████   ██████  ███    ██ 
██       ██ ██   ██ ██   ██ ██    ██ ████   ██ 
██   ███ ██ ██████  ██████  ██    ██ ██ ██  ██ 
██    ██ ██ ██   ██ ██   ██ ██    ██ ██  ██ ██ 
 ██████  ██ ██████  ██████   ██████  ██   ████ 
`

// Reads from the input source until encountering a newline, take the just read line and pass it to an instance
// of the lexer and finally print all the tokens the lexer gives us until we encounter EOF.
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	fmt.Println(WELCOME_IMAGE)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.NewLexer(line)
		p := parser.NewParser(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Woops! we ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
