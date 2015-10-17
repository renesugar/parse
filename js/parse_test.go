package js // import "github.com/tdewolff/parse/js"

import (
	"bytes"
	"fmt"
	"testing"
)

func printParse(input string) {
	p := NewParser(bytes.NewBufferString(input))
	for {
		gt, tokens := p.Next()
		fmt.Println(gt, tokens)
		if gt == ErrorGrammar {
			break
		}
	}
}

////////////////////////////////////////////////////////////////

func TestParser(t *testing.T) {
	printParse("{;}")
}
