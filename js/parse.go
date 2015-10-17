package js // import "github.com/tdewolff/parse/js"

import (
	"io"
	"strconv"
)

// GrammarType determines the type of grammar.
type GrammarType uint32

// GrammarType values.
const (
	ErrorGrammar GrammarType = iota // extra token when errors occur
	EmptyStatementGrammar
	BlockGrammar
)

// String returns the string representation of a GrammarType.
func (tt GrammarType) String() string {
	switch tt {
	case ErrorGrammar:
		return "Error"
	case EmptyStatementGrammar:
		return "EmptyStatement"
	case BlockGrammar:
		return "Block"
	}
	return "Invalid(" + strconv.Itoa(int(tt)) + ")"
}

////////////////////////////////////////////////////////////////

// State is the state function the parser currently is in.
type State func(TokenType, []byte) GrammarType

// Token is a single TokenType and its associated data.
type Token struct {
	TokenType
	Data []byte
}

// Parser is the state for the parser.
type Parser struct {
	l   *Lexer
	err error

	state []State

	buf []Token
	pos int
	n   int
}

// NewParser returns a new CSS parser from an io.Reader. isStylesheet specifies whether this is a regular stylesheet (true) or an inline style attribute (false).
func NewParser(r io.Reader) *Parser {
	p := &Parser{
		l: NewLexer(r),
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
	p.buf = p.buf[:0]
	p.l.Free(p.n)
	p.n = 0

	tt, data := p.next()
	state := p.state[len(p.state)-1](tt, data)
	return state, p.buf
}

func (p *Parser) next() (TokenType, []byte) {
	if p.pos < len(p.buf) {
		t := p.buf[p.pos]
		p.pos++
		return t.TokenType, t.Data
	}
	tt, data, n := p.l.Next()
	p.n += n
	for tt == WhitespaceToken || tt == CommentToken {
		tt, data, n = p.l.Next()
		p.n += n
	}
	p.buf = append(p.buf, Token{tt, data})
	p.pos++
	return tt, data
}

func (p *Parser) rewind(n int) {
	p.pos -= n
}

func (p *Parser) pushState(s State) {
	p.state = append(p.state, s)
}

func (p *Parser) popState() {
	p.state = p.state[:len(p.state)-1]
}

////////////////////////////////////////////////////////////////

func (p *Parser) parseProgram(tt TokenType, data []byte) GrammarType {
	//if tt == IdentifierToken && parse.Equal(data, []byte("function")) {
	//	return p.parseFunction(tt, data)
	//} else
	if tt == ErrorToken {
		return ErrorGrammar
	}
	return p.parseStatement(tt, data)
}

func (p *Parser) parseStatement(tt TokenType, data []byte) GrammarType {
	if tt == PunctuatorToken {
		if data[0] == ';' {
			return EmptyStatementGrammar
		} else if data[0] == '{' {
			p.pushState(p.parseStatement)
			return BlockGrammar
		} else if data[0] == '}' {
			p.popState()
			return BlockGrammar
		} else {
			return ErrorGrammar
		}
	}
	//  else if tt == IdentifierToken {

	// }
	return ErrorGrammar
}
