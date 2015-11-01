package js // import "github.com/tdewolff/parse/js"

import (
	"fmt"
	"io"
	"strconv"

	"github.com/tdewolff/parse"
)

// GrammarType determines the type of grammar.
type GrammarType uint32

// GrammarType values.
const (
	ErrorGrammar GrammarType = iota // extra token when errors occur
	FunctionGrammar
	EmptyStatementGrammar
	BlockGrammar
	VarStatementGrammar
)

// String returns the string representation of a GrammarType.
func (tt GrammarType) String() string {
	switch tt {
	case ErrorGrammar:
		return "Error"
	case FunctionGrammar:
		return "Function"
	case EmptyStatementGrammar:
		return "EmptyStatement"
	case BlockGrammar:
		return "Block"
	case VarStatementGrammar:
		return "VarStatement"
	}
	return "Invalid(" + strconv.Itoa(int(tt)) + ")"
}

////////////////////////////////////////////////////////////////

// State is the state function the parser currently is in.
type State func(Token) GrammarType

// Token is a single TokenType and its associated data.
type Token struct {
	TokenType
	Data []byte

	line   int
	prevWS bool
	prevLT bool
}

// Parser is the state for the parser.
type Parser struct {
	l   *Lexer
	err error

	state []State
	line  int

	buf []Token
	end int
	n   int
}

// NewParser returns a new CSS parser from an io.Reader. isStylesheet specifies whether this is a regular stylesheet (true) or an inline style attribute (false).
func NewParser(r io.Reader) *Parser {
	p := &Parser{
		l:    NewLexer(r),
		line: 1,
	}
	p.state = []State{p.parseProgram}
	return p
}

// Err returns the error encountered during parsing, this is often io.EOF but also other errors can be returned.
func (p *Parser) Err() error {
	if p.err != nil {
		return p.err
	}
	return p.l.Err()
}

// Next returns the next Grammar. It returns ErrorGrammar when an error was encountered. Using Err() one can retrieve the error message.
func (p *Parser) Next() (GrammarType, []Token) {
	copy(p.buf, p.buf[p.end:])
	p.buf = p.buf[:len(p.buf)-p.end]
	p.end = 0

	p.l.Free(p.n)
	p.n = 0

	t := p.next()
	state := p.state[len(p.state)-1](t)
	return state, p.buf[:p.end]
}

func (p *Parser) last() Token {
	return p.buf[len(p.buf)-1]
}

func (p *Parser) next() Token {
	if p.end < len(p.buf) {
		p.end++
		return p.buf[p.end-1]
	}
	prevWS := false
	prevLT := false
	tt, data, n := p.l.Next()
	p.n += n
	for tt == WhitespaceToken || tt == CommentToken || tt == LineTerminatorToken {
		if tt == LineTerminatorToken {
			p.line++
			prevLT = true
		} else if tt == WhitespaceToken {
			prevWS = true
		}
		tt, data, n = p.l.Next()
		p.n += n
	}
	p.buf = append(p.buf, Token{tt, data, p.line, prevWS, prevLT})
	p.end++
	return p.buf[p.end-1]
}

func (p *Parser) back() {
	p.end--
}

func (p *Parser) pushState(s State) {
	p.state = append(p.state, s)
}

func (p *Parser) popState() {
	p.state = p.state[:len(p.state)-1]
}

////////////////////////////////////////////////////////////////

func (p *Parser) parseProgram(t Token) GrammarType {
	if t.TokenType == IdentifierToken && parse.Equal(t.Data, []byte("function")) {
		t = p.next()
		return p.parseFunction(t)
	}
	return p.parseStatement(t)
}

func (p *Parser) parseFunctionBody(t Token) GrammarType {
	if t.TokenType == PunctuatorToken && t.Data[0] == '}' {
		p.popState()
		return FunctionGrammar
	}
	return p.parseProgram(t)
}

func (p *Parser) parseFunction(t Token) GrammarType {
	if t.TokenType != IdentifierToken {
		return p.produceError("unexpected '" + string(t.Data) + "' in function declaration; expected identifier")
	}
	t = p.next()
	if t.TokenType != PunctuatorToken || t.Data[0] != '(' {
		return p.produceError("unexpected '" + string(t.Data) + "' in function declaration; expected (")
	}
	t = p.next()
	fmt.Println("a", t, string(t.Data))
	for t.TokenType == IdentifierToken {
		t = p.next()
		fmt.Println("b", t, string(t.Data))
		if t.TokenType != PunctuatorToken || t.Data[0] != ',' {
			break
		}
		t = p.next()
		fmt.Println("c", t, string(t.Data))
	}
	p.back()
	t = p.next()
	fmt.Println("d", p.end, t, string(t.Data))
	if t.TokenType != PunctuatorToken || t.Data[0] != ')' {
		return p.produceError("unexpected '" + string(t.Data) + "' in function declaration; expected )")
	}
	t = p.next()
	fmt.Println("e", p.end, t, string(t.Data))
	if t.TokenType != PunctuatorToken || t.Data[0] != '{' {
		return p.produceError("unexpected '" + string(t.Data) + "' in function declaration; expected {")
	}
	p.pushState(p.parseFunctionBody)
	return FunctionGrammar
}

func (p *Parser) parseStatement(t Token) GrammarType {
	if t.TokenType == PunctuatorToken {
		if t.Data[0] == ';' {
			return EmptyStatementGrammar
		} else if t.Data[0] == '{' {
			p.pushState(p.parseStatement)
			return BlockGrammar
		} else if t.Data[0] == '}' {
			p.popState()
			return BlockGrammar
		}
	} else if t.TokenType == IdentifierToken {
		if parse.Equal(t.Data, []byte("var")) {
			if !p.consumeVariableDeclarationList() {
				return p.produceError("unexpected '" + string(p.last().Data) + "' in var statement")
			}
			if !p.consumeSemicolon() {
				return p.produceError("unexpected '" + string(p.last().Data) + "' in var statement, expected semicolon")
			}
			return VarStatementGrammar
		}
	} else if t.TokenType == ErrorToken {
		return ErrorGrammar
	}
	return p.produceError("unexpected '" + string(t.Data) + "', expected statement")
}

////////////////////////////////////////////////////////////////

func (p *Parser) produceError(msg string) GrammarType {
	p.err = parse.NewSyntaxError(msg, p.buf[p.end-1].line)
	return ErrorGrammar
}

////////////////////////////////////////////////////////////////

func (p *Parser) consumeVariableDeclarationList() bool {
	if !p.consumeVariableDeclaration() {
		return false
	}
	t := p.next()
	for t.TokenType == PunctuatorToken && t.Data[0] == ',' {
		if !p.consumeVariableDeclaration() {
			break
		}
		t = p.next()
	}
	p.back()
	return true
}

func (p *Parser) consumeVariableDeclaration() bool {
	if t := p.next(); t.TokenType != IdentifierToken {
		p.back()
		return false
	} else if t := p.next(); t.TokenType == PunctuatorToken && t.Data[0] == '=' {
		p.consumeAssignmentExpression()
	}
	return true
}

func (p *Parser) consumeExpression(allowIn bool) bool {
	if !p.consumeAssignmentExpression(allowIn) {
		return false
	}
	t := p.next()
	for t.TokenType == PunctuatorToken && t.Data[0] == ',' {
		if !p.consumeAssignmentExpression(allowIn) {
			break
		}
		t = p.next()
	}
	p.back()
	return true
}

func (p *Parser) consumeAssignmentExpression(allowIn bool) bool {
	p.next()
	return true
}

func (p *Parser) consumeSemicolon() bool {
	t := p.next()
	if t.TokenType == PunctuatorToken {
		if t.Data[0] == ';' {
			return true
		} else if t.Data[0] == '}' {
			p.back()
			return true
		}
	} else if t.TokenType == ErrorToken && p.l.Err() == io.EOF {
		p.back()
		return true
	} else if t.prevLT {
		p.back()
		return true
	}
	p.back()
	return false
}
