package main

import (
	"fmt"
	"os"
	"unicode"
)

type Token struct {
	Value string
	Type  string
}

type Lexer struct {
	input   string
	current int
}

func (l *Lexer) readString() string {
	curr := l.input[l.current]
	str := ""

	for curr != '"' {
		str += string(l.input[l.current])
		l.current++
		curr = l.input[l.current]
	}

	// po to aby wskaznik current nie byl na "
	l.current++
	return str
}

func (l *Lexer) readNumber() string {
	curr := l.input[l.current]
	numb := ""

	for unicode.IsDigit(rune(curr)) {
		numb += string(curr)
		l.current++
		curr = l.input[l.current]
	}

	return numb
}

func (l *Lexer) readIdent() string {
	curr := l.input[l.current]
	numb := ""

	for unicode.IsLetter(rune(curr)) {
		numb += string(curr)
		l.current++
		curr = l.input[l.current]
	}

	return numb
}

func (l *Lexer) NextToken() Token {
	if l.current >= len(l.input) {
		return Token{Value: "end"}
	}

	curr := l.input[l.current]

	switch curr {
	case '{':
		l.current++
		return Token{Value: "{", Type: "LEFTBRACE"}
	case '}':
		l.current++
		return Token{Value: "}", Type: "RIGHTBRACE"}
	case '[':
		l.current++
		return Token{Value: "[", Type: "LEFTBRACKET"}
	case ']':
		l.current++
		return Token{Value: "]", Type: "RIGHTBRACKET"}
	case '"':
		l.current++
		return Token{Value: l.readString(), Type: "STRING"}
	case ':':
		l.current++
		return Token{Value: ":", Type: "COLON"}
	case ',':
		l.current++
		return Token{Value: ",", Type: "COMA"}
	case '\n':
		l.current++
		return Token{Value: "\n", Type: "NEWLINE"}
	case ' ':
		l.current++
		return Token{Value: " ", Type: "WHITESPACE"}
	default:
		if unicode.IsLetter(rune(curr)) {
			ident := l.readIdent()
			return Token{Value: ident, Type: "BOOLEAN"}
		} else if unicode.IsDigit(rune(curr)) {
			numb := l.readNumber()
			return Token{Value: numb, Type: "NUMBER"}
		}
	}

	return Token{Value: "INVALID"}
}

type Parser struct {
	l    Lexer
	curr int
}

func (p *Parser) parseArray() []any {
	arr := []any{}

	for {
		token := p.l.NextToken()
		for token.Type == "WHITESPACE" || token.Type == "NEWLINE" {
			token = p.l.NextToken()
		}

		fmt.Println(token)
		if token.Type == "RIGHTBRACKET" {
			break
		}

        arr = append(arr, token.Value)
		if token.Type != "STRING" && token.Type != "BOOLEAN" && token.Type != "NUMBER" {
			panic("expected string, bool, nubmer in array")
		}

		token = p.l.NextToken()
		for token.Type == "WHITESPACE" || token.Type == "NEWLINE" {
			token = p.l.NextToken()
		}

		fmt.Println(token)
		if token.Type == "RIGHTBRACKET" {
			break
		}

		if token.Type != "COMA" {
			panic("expected coma after element in arry")
		}
	}

	return arr
}

func (p *Parser) parseObject() map[any]any {
	obj := make(map[any]any)

	for {
		token := p.l.NextToken()
		// could be ommited during lexing stage
		// otherwise we have to eat those tokens after each nextTOken call
		for token.Type == "WHITESPACE" || token.Type == "NEWLINE" {
			token = p.l.NextToken()
		}
		fmt.Println(token)

		if token.Type == "RIGHTBRACE" {
			break
		}

		// during object parsing first there HAS to be a key which is a string
		if token.Type != "STRING" {
			panic("expected string as a key in object")
		}
		key := token.Value

		token = p.l.NextToken()
		for token.Type == "WHITESPACE" || token.Type == "NEWLINE" {
			token = p.l.NextToken()
		}
		fmt.Println(token)
		// next we need a ":" sign
		if token.Type != "COLON" {
			panic("expected colon after a key in object")
		}

		// next comes any value
		token = p.l.NextToken()
		for token.Type == "WHITESPACE" || token.Type == "NEWLINE" {
			token = p.l.NextToken()
		}

		fmt.Println(token)
		if token.Type == "LEFTBRACKET" {
			arr := p.parseArray()
            obj[key] = arr
		} else if token.Type == "LEFTBRACE" {
            objct := p.parseObject()
            obj[key] = objct
		} else {
            obj[key] = token.Value
        }

		// at the end comes coma, but only if that's not the last key-value
		token = p.l.NextToken()
		for token.Type == "WHITESPACE" || token.Type == "NEWLINE" {
			token = p.l.NextToken()
		}

		fmt.Println(token)

		if token.Type == "RIGHTBRACE" {
			break
		}

		if token.Type != "COMA" {
			panic("expected coma at the eol")
		}
	}

	return obj
}

func (p *Parser) Parse() map[any]any {
	token := p.l.NextToken()
	fmt.Println(token)
	if token.Type == "LEFTBRACE" {
        return p.parseObject()
	}

    return map[any]any{}
}

// TODO: null 
func main() {
	f, _ := os.ReadFile("tests/step2/invalid.json")
	l := Lexer{input: string(f)}
	parser := Parser{l: l}
    res := parser.Parse()
    fmt.Println(res)
}
