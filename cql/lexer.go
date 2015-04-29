package cql

import (
	"bufio"
	"bytes"
	"os"
	"unicode/utf8"
)

//
// Yeah, ok. This is not a real lexer. Maybe calling it a lexer will annoy me enough to
// come back and write a proper CQL lexer/parser one day?
//
// IT IS ALSO IMPORTANT TO NOTE: that, at the moment, this will lock up if there is a comment
// at the end of the file i.e: one with no new line character after it.
//
func ReadCQLFile(file *os.File) (statements []string, err error) {

	s := bufio.NewScanner(file)
	s.Split(scanCQLExpressions)

	for s.Scan() {
		statements = append(statements, s.Text())
	}
	return statements, nil
}

func scanCQLExpressions(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip whitespace
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !isSpace(r) {
			break
		}
	}
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	buf := bytes.Buffer{}
	var r rune
	for width, i := 0, start; i < len(data); i += width {
		r, width = utf8.DecodeRune(data[i:])
		switch r {
		case '/':
			r2, width2 := utf8.DecodeRune(data[i+width:])
			if r2 == '/' { // Remove single line C style comment
				for i += width2; r2 != '\n'; i += width2 {
					r2, width2 = utf8.DecodeRune(data[i+width:])
				}
				continue
			}
			if r2 == '*' { // TODO: Multi line comments are legal too, let's not forget. But, frankly, I really can't be arsed to deal with them.
				panic("Ahh shit. This is a multiline comment and we don't handle that right now")
			}
		case '-':
			r2, width2 := utf8.DecodeRune(data[i+width:])
			if r2 == '-' { // Remove single line SQL style comment
				for i += width2; r2 != '\n'; i += width2 {
					r2, width2 = utf8.DecodeRune(data[i+width:])
				}
				continue
			}
		case ';':
			return i + width, buf.Bytes(), nil
		}
		buf.WriteRune(r)
	}

	if atEOF && len(data) > start {
		return len(data), buf.Bytes(), nil
	}
	return 0, nil, nil
}

func isSpace(r rune) bool {
	if r <= '\u00FF' {
		switch r {
		case ' ', '\t', '\n', '\v', '\f', '\r':
			return true
		case '\u0085', '\u00A0':
			return true
		}
		return false
	}
	if '\u2000' <= r && r <= '\u200a' {
		return true
	}
	switch r {
	case '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000':
		return true
	}
	return false
}
