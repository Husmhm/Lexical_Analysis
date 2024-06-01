package main
import (
    "io"
    "bufio"
	"unicode"
	"os"
	"fmt"
	"os/exec"
	"runtime"

)


type Token int
var Error bool
var iskeyword bool
const (
	EOF = iota
	ILLEGAL
	IDENTIFIER
	INT
	FLOAT
	KEYWORD
	SEMICOLON // ;

	OPENPARANTESIS // (
	CLOSEPRANTESIS // )
	OPENOCOLUD     // {
	CLOSEOCOLUD    // }
	ADD            // +
	SUB            // -
	MUL            // *
	DIV            // /
	// Error

	ASSIGN // =
	MORE
	MOREEQUAL

	LESS
	LESSEQUAL
	EQUAL_EQUAL
	PLUS_PLUS
	MINUS_MINUS
)

var tokens = []string{
	EOF:        "EOF",
	ILLEGAL:    "ILLEGAL",
	IDENTIFIER: "ID",
	INT:        "INT",
	FLOAT:      "FLOAT",
	KEYWORD:	"KEYWORD",
	SEMICOLON:  "SEMICOLON",

	// Infix ops
	OPENPARANTESIS: "OPENPARANTESIS",
	CLOSEPRANTESIS: "CLOSEPRANTESIS",
	OPENOCOLUD:     "OPENOCOLUD",
	CLOSEOCOLUD:    "CLOSEOCOLUD",
	ADD:            "PLUS",
	SUB:            "SUB",
	MUL:            "STAR",
	DIV:            "DIVIDE",
	ASSIGN: "EQUAL",
	MORE:"MORE",
	MOREEQUAL:"MOREEQUAL",
	LESS:"LESS",
	LESSEQUAL:"LESSEQUAL",
	EQUAL_EQUAL:"EQUAL_EQUAL",
	PLUS_PLUS:"PLUS_PLUS",
	MINUS_MINUS:"MINUS_MINUS",

}

func (t Token) String() string {
	return tokens[t]
}

type Position struct {
	line   int
	column int
}

type Lexer struct {
    pos            Position
    reader         *bufio.Reader
    quotationCount int

}

func NewLexer(reader io.Reader) *Lexer {
	return &Lexer{
		pos:    Position{line: 1, column: 0},
		reader: bufio.NewReader(reader),
		quotationCount: 0,
	}
}

func (l *Lexer) Lex() (Position, Token, string) {
    for {
        r, _, err := l.reader.ReadRune()
        if err != nil {
            if err == io.EOF {
                return l.pos, EOF, ""
            }
            panic(err)
        }
        l.pos.column++

        switch r {
        case '\n':
            l.resetPosition()
        case ';':
            return l.pos, SEMICOLON, ";"
        case '+':
			l.pos.column++
			z, _, _ := l.reader.ReadRune()
			if z =='+'{
				return l.pos,PLUS_PLUS,"++"
			}else{
				return l.pos,ADD,"+"
			}
        case '-':
			l.pos.column++
			z, _, _ := l.reader.ReadRune()
			if z =='-'{
				return l.pos,MINUS_MINUS,"--"
			}else{
				return l.pos,SUB,"-"
			}
        case '*':
            return l.pos, MUL,"*"
		case '>':
			l.pos.column++
			z, _, _ := l.reader.ReadRune()
			if z =='='{
				return l.pos,MOREEQUAL,">="
			}else{
				l.pos.column--
				return l.pos,MORE,">"
			}
		case '<':
			l.pos.column++
			z, _, _ := l.reader.ReadRune()
			if z =='='{
				return l.pos,LESSEQUAL,"<="
			}else{
				l.pos.column--
				return l.pos,LESS,"<"
			}
			
        case '/':
            l.pos.column++
            z, _, err := l.reader.ReadRune()
            if err != nil {
                panic(err)
            }
            if z == '/' {
                // Single-line comment
                for {
                    r, _, err := l.reader.ReadRune()
                    if err != nil || r == '\n' {
                        // Comment ends or EOF
                        l.resetPosition()
                        break
                    }
                }
                continue
            }   
			if z == '*' {
                // Multi-line comment
				l.multy_comment()
            } else {
                // Division operator
                l.backup()
                return l.pos, DIV, "/"
            }
            
		case '(':
			return l.pos, OPENPARANTESIS, "("
		case ')':
			return l.pos, CLOSEPRANTESIS, ")"
		case '{':
			return l.pos,OPENOCOLUD,"{"
		case '}':
			return l.pos,CLOSEOCOLUD,"}"
        case '=':
			l.pos.column++
			z, _, _ := l.reader.ReadRune()
			if z =='='{
				return l.pos,EQUAL_EQUAL,"=="
			}else{
				l.pos.column--
				return l.pos, ASSIGN, "="
			}
		case '"':
			l.quotationCount ++
        default:
            if unicode.IsSpace(r) {
                continue // nothing to do here, just move on
            }else if unicode.IsDigit(r) {
				// backup and let lexInt rescan the beginning of the int
				startPos := l.pos
				l.backup()
				lit,hasDecimal := l.lexNumber()
				if hasDecimal == true {
					return startPos,FLOAT,lit
				}else {
				return startPos, INT, lit}
			} else if r == '_' || unicode.IsLetter(r) {
				// backup and let lexIdent rescan the beginning of the ident
				if r == '_' {
					startPos := l.pos
					lit,_ := l.lexIdent()
					lit2 := "_"+lit
					return startPos, IDENTIFIER, lit2	
				}
				startPos := l.pos
				l.backup()
				lit,s := l.lexIdent()
				if s == true {
					return startPos, KEYWORD, lit
				}else{return startPos, IDENTIFIER, lit}

			}	
			
        }
		
	}
}

func (l *Lexer) resetPosition() {
	l.pos.line++
	l.pos.column = 0
}

func (l *Lexer)  multy_comment(){
	for{
		l.pos.column++
		r, _, _ := l.reader.ReadRune()
		if r =='*'{
			for{
				l.pos.column++
				r, _, _ := l.reader.ReadRune()
				if r == '/'{
					l.pos.line++
					l.pos.column = 0
					return
				}else if  unicode.IsLetter(r){
					l.lexIdent()
					continue
				}else if unicode.IsSpace(r) {
					continue
				}else if unicode.IsDigit(r) {
					l.lexNumber()
					continue
				}
			}
		}
	}

}

func (l *Lexer) backup() {
	if err := l.reader.UnreadRune(); err != nil {
		panic(err)
	}
	
	l.pos.column--
}
func (l *Lexer) nextchar() {
	if err := l.reader.UnreadRune(); err != nil {
		panic(err)
	}
	
	l.pos.column++
}
func (l *Lexer) lexNumber() (string,bool) {
	Error = false
	var lit string
	var hasDecimal bool

	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				// at the end of the number
				return lit , hasDecimal
			}
		}

		l.pos.column++
		if unicode.IsDigit(r) {
			lit = lit + string(r)
		} else if r == '.' && !hasDecimal {
			l.pos.column++
			r, _, _ := l.reader.ReadRune()
			if unicode.IsDigit(r) {
				lit = lit +"."+string(r)
				hasDecimal = true
			}else{
				Error = true
			}

		} else {
			// scanned something not in the number
			l.backup()
			return lit , hasDecimal
		}

	}
}
func (l *Lexer) lexIdent() (string, bool) {
	var lit string
	Error = false
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				// at the end of the identifier
				return lit , Error
			}
		}
		
        l.pos.column++
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			lit = lit + string(r)
		} else if r =='_' || r == ',' {
			Error = true
		}else {
			// scanned something not in the identifier
			l.backup()
			s := IsKeyword(lit)
			return lit , s
		}
	}
	
}
func IsKeyword(lit string) bool{
	iskeyword = false
	if lit == "int" || lit=="float" ||lit == "if"|| lit =="while"|| lit=="print" || lit=="scan"||lit=="else"{
		iskeyword =true
	}
	return iskeyword
}


func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func main() {
	file, err := os.Open("input_3_error.txt")
	if err != nil {
		panic(err)
	}
	
	
	lexer := NewLexer(file)
	for {
		if Error== true {
			clearScreen()
			fmt.Println("Error")
			break
			
		}else{
		_, tok, lit := lexer.Lex()
		if tok == EOF {
			break
		}
		fmt.Printf("<%s\t\t%s>\n", tok, lit)}

	}
	if lexer.quotationCount%2 !=0 {
		clearScreen()
		// fmt.Println(lexer.quotationCount)
		fmt.Println("Error")
	}
}
