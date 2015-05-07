package parse // import "github.com/tdewolff/parse"

import (
	"bytes"
	"math/rand"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func helperRand(n, m int, chars []byte) [][]byte {
	r := make([][]byte, n)
	for i := range r {
		for j := 0; j < m; j++ {
			r[i] = append(r[i], chars[rand.Intn(len(chars))])
		}
	}
	return r
}

func assertSplitDataURI(t *testing.T, x, e1, e2 string, eerr error) {
	s1, s2, err := SplitDataURI([]byte(x))
	assert.Equal(t, eerr, err, "ok must match in "+x)
	assert.Equal(t, e1, string(s1), "mediatype part must match in "+x)
	assert.Equal(t, e2, string(s2), "data part must match in "+x)
}

////////////////////////////////////////////////////////////////

var wsSlices [][]byte

func init() {
	wsSlices = helperRand(100, 20, []byte("abcdefg \n\r\f\t"))
}

func TestCopy(t *testing.T) {
	foo := []byte("abc")
	bar := Copy(foo)
	foo[0] = 'b'
	assert.Equal(t, "bbc", string(foo))
	assert.Equal(t, "abc", string(bar))
}

func TestToLower(t *testing.T) {
	foo := []byte("Abc")
	bar := ToLower(foo)
	bar[1] = 'B'
	assert.Equal(t, "aBc", string(foo))
	assert.Equal(t, "aBc", string(bar))
}

func TestCopyToLower(t *testing.T) {
	foo := []byte("Abc")
	bar := CopyToLower(foo)
	bar[1] = 'B'
	assert.Equal(t, "Abc", string(foo))
	assert.Equal(t, "aBc", string(bar))
}

func TestEqual(t *testing.T) {
	assert.Equal(t, true, Equal([]byte("abc"), []byte("abc")))
	assert.Equal(t, false, Equal([]byte("abcd"), []byte("abc")))
	assert.Equal(t, false, Equal([]byte("bbc"), []byte("abc")))

	assert.Equal(t, true, EqualCaseInsensitive([]byte("Abc"), []byte("abc")))
	assert.Equal(t, false, EqualCaseInsensitive([]byte("Abcd"), []byte("abc")))
	assert.Equal(t, false, EqualCaseInsensitive([]byte("Bbc"), []byte("abc")))
}

func TestWhitespace(t *testing.T) {
	assert.Equal(t, true, IsAllWhitespace([]byte("\t \r\n\f")))
	assert.Equal(t, false, IsAllWhitespace([]byte("\t \r\n\fx")))
}

func TestReplaceMultipleWhitespace(t *testing.T) {
	multipleWhitespaceRegexp := regexp.MustCompile("\\s+")
	for _, e := range wsSlices {
		reference := multipleWhitespaceRegexp.ReplaceAll(e, []byte(" "))
		assert.Equal(t, string(reference), string(ReplaceMultiple(e, IsWhitespace, ' ')), "must remove all multiple whitespace")
	}
}

func TestNormalizeContentType(t *testing.T) {
	assert.Equal(t, "text/html", string(NormalizeContentType([]byte("text/html"))))
	assert.Equal(t, "text/html;charset=utf-8", string(NormalizeContentType([]byte("text/html; charset=UTF-8"))))
	assert.Equal(t, "text/html;charset=utf-8;param=\" ; \"", string(NormalizeContentType([]byte("text/html; charset=UTF-8 ; param = \" ; \""))))
	assert.Equal(t, "text/html,text/css", string(NormalizeContentType([]byte("text/html, text/css"))))
}

func TestTrim(t *testing.T) {
	assert.Equal(t, "a", string(Trim([]byte("a"), IsWhitespace)))
	assert.Equal(t, "a", string(Trim([]byte(" a"), IsWhitespace)))
	assert.Equal(t, "a", string(Trim([]byte("a "), IsWhitespace)))
	assert.Equal(t, "", string(Trim([]byte(" "), IsWhitespace)))
}

func TestSplitDataURI(t *testing.T) {
	assertSplitDataURI(t, "www.domain.com", "", "", ErrBadDataURI)
	assertSplitDataURI(t, "data:,", "text/plain", "", nil)
	assertSplitDataURI(t, "data:text/xml,", "text/xml", "", nil)
	assertSplitDataURI(t, "data:,text", "text/plain", "text", nil)
	assertSplitDataURI(t, "data:;base64,dGV4dA==", "text/plain", "text", nil)
	assertSplitDataURI(t, "data:image/svg+xml,", "image/svg+xml", "", nil)
}

////////////////////////////////////////////////////////////////

func BenchmarkBytesTrim(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, e := range wsSlices {
			e = bytes.TrimSpace(e)
		}
	}
}

func BenchmarkTrim(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, e := range wsSlices {
			e = Trim(e, IsWhitespace)
		}
	}
}
