package js // import "github.com/tdewolff/parse/js"

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/tdewolff/parse"
)

func printParse(input string) {
	p := NewParser(bytes.NewBufferString(input))
	for {
		gt, tokens := p.Next()
		fmt.Println(gt, tokens)
		if gt == ErrorGrammar {
			err := p.Err()
			if serr, ok := err.(*parse.SyntaxError); ok {
				fmt.Printf("%d: %s\n", serr.Line, err)
			} else {
				fmt.Println(err)
			}
			break
		}
	}
}

////////////////////////////////////////////////////////////////

func TestParser(t *testing.T) {
	printParse("function a(b,c){var d = 5;}")
}
