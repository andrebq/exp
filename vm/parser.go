package main

import (
	"fmt"
	"unicode/utf8"
)

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

// return a string that skip all whitespace chars
// on the left of in
func discardWhiteSpace(in string) string {
	for len(in) > 0 {
		r, size := utf8.DecodeRuneInString(in)
		if isSpace(r) {
			in = in[size:]
			continue
		}
		break
	}
	return in
}

// read the ( that opens a expression, 
// return the text representing the node,
// the tail of the input string,
// and a bool representing if the value is valid
func openExp(in string) (string, string, bool) {
	orig := in
	// ( is a valid ascii, no need to work with
	// runes where
	if len(in) == 0 || in[0] != '(' {
		return "", orig, false
	}
	return in[:1], in[1:], true
}

// same as openExp but searchs for the ')'
func closeExp(in string) (string, string, bool) {
	orig := in
	if len(in) == 0 || in[0] != ')' {
		return "", orig, false
	}
	return in[:1], in[1:], true
}

// check if the input is a valid symbol.
// a symbol is anything that starts with a letter or _
// and don't have any whitespace between
// this-is-a-valid-symbol
// this!is_another?crazy_symbol
// THIS
func symbol(in string) (string, string, bool) {
	orig := in
	sym, sz := utf8.DecodeRuneInString(in)
	if isSpace(sym) || isDigit(sym) {
		// a sym MUST START with something different
		// from a digit or space
		return "", orig, false
	}
	in = in[sz:]
	// okay, go ahread and read everything until you find
	// a space
	for len(in) > 0 {
		r, w := utf8.DecodeRuneInString(in)
		if isSpace(r) {
			break
		} else {
			// not a whitespace
			// move the sz counter by w bytes
			// and use the tail of input
			in = in[w:]
			sz += w
		}
	}
	// the first space found is kept intact
	return orig[0:sz], in, true
}

// check if the input string matches the required test
func is(input string, test func(string)(string, string,bool)) (string, string, bool) {
	text, tail, ok := test(input)
	if ok {
		return text, tail, ok
	} else {
		return "", input, false
	}
}

func getNode(input string) (string, *Node, error) {
	input = discardWhiteSpace(input)
	if _, tail, ok:= is(input, openExp); ok {
		return tail, NewNode(OEXP, "("), nil
	} else if _, tail, ok := is(input, closeExp); ok {
		return tail, NewNode(CEXP, ")"), nil
	} else if sym, tail, ok := is(input, symbol); ok {
		return tail, NewNode(SYM, sym), nil
	} else {
		return input, nil, errorf("input isn't valid. tail: %q", tail)
	}
}

func errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
