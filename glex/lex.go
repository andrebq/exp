package glex
/*Copyright (c) 2012 André Luiz Alves Moraes

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.*/

import (
	"io/ioutil"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Lexer errors
type Error string

// Error interface
func (e Error) Error() string { return string(e) }

// Stringer interface
func (e Error) String() string { return string(e) }

const (
	NoMatch              = Error("No match was found in the list")
	Eof                  = Error("End of input")
	ErrInvalidUtf8String = Error("Input string isn't a valid UTF-8 sequence")
)

// Interface that define's a Match
type Match struct {
	// text of the match
	Text string

	// token number
	Token int
}

var (
	// used just to return a empty match
	// avoid doing one allocation for every call to Match methods
	emptyMatch Match
)

// Represent a rule for the lexer
type Rule struct {
	// regular expression for the given match
	exp *regexp.Regexp

	// Token for the given matches
	token int
}

// Create a new rule based on the given regular expression.
//
// If the re don't match the start of the string (^), it will have the (^) character inserted.
func NewRule(re string, token int) (*Rule, error) {
	if !strings.HasPrefix(re, "^") {
		re = "^" + re
	}
	r := &Rule{token: token}
	var err error
	r.exp, err = regexp.Compile(re)
	return r, err
}

// Check if the rule can match the given string
//
// When the input string is empty, err will hold Eof, if no match is found, err will hold NoMatch
//
// On any other cases, this will return a Match, the rest of the string and nil.
func (r *Rule) Match(input string) (Match, string, error) {
	if len(input) == 0 {
		return emptyMatch, input, Eof
	}

	loc := r.exp.FindStringIndex(input)
	if loc == nil {
		return emptyMatch, input, NoMatch
	}

	m := Match{Text: input[loc[0]:loc[1]], Token: r.token}
	return m, input[loc[1]:], nil
}

// Hold the list of rules
type RuleList []*Rule

// Try to match with one of the rules from the list
//
// This method simply call's the Match method of each Rule in the list.
func (l RuleList) Match(input string) (Match, string, error) {
	if len(input) == 0 {
		return emptyMatch, input, Eof
	}

	for _, r := range l {
		m, tail, err := r.Match(input)
		if err == nil {
			return m, tail, err
		}
	}

	return emptyMatch, input, NoMatch
}

// Include the given expression/token pair into this
// rule list.
//
// If there are erros in the epxression, the rule isn't
// included and the error is returned.
func (l *RuleList) NewRule(exp string, token int) (*Rule, error) {
	r, err := NewRule(exp, token)
	if err != nil {
		return nil, err
	}
	*l = append(*l, r)
	return r, nil
}

// Hold the information to perform the lexical scanning
// of a given input
type Lexer struct {
	rules     RuleList
	fullInput string
	current   string
}

// Create a new lexer from the given input string
func NewLexer(input string) (*Lexer, error) {
	if !utf8.ValidString(input) {
		err := ErrInvalidUtf8String
		return nil, err
	}

	return &Lexer{rules: make(RuleList, 0), fullInput: input, current: input}, nil
}

// Create a new lexer loading it's contents from the given file
// file contents MUST BE valid UTF-8 strings.
func NewLexerFromFile(file string) (*Lexer, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if !utf8.Valid(data) {
		err = ErrInvalidUtf8String
		return nil, err
	}

	input := string(data)
	return &Lexer{rules: make(RuleList, 0), fullInput: input, current: input}, nil
}

// Try to find a match for the current input and move the cursor, if the returned token
// is < 0, the input is consumed and the lexer tries to advance again.
//
// If not match is found, return's an error. Later the function IsEof can be used to check
// if the error represent the end of input.
func (l *Lexer) Next() (Match, error) {
	for {
		// this loop will consume input until it find's an error
		// or a token >= 0.
		m, tail, err := l.rules.Match(l.current)
		if err == nil {
			// move the cursor only if didn't found error
			l.current = tail
			if m.Token < 0 {
				// search for anoter token
				continue
			}
			return m, err
		} else {
			return emptyMatch, err
		}
	}
	panic("not reached")
	return emptyMatch, nil
}

// Return true only if the error is Eof
func (l *Lexer) Eof(err error) bool {
	return err == Eof
}

// Include a new rule into the lexer, same logic used
// by RuleList
func (l *Lexer) NewRule(exp string, token int) error {
	_, err := l.rules.NewRule(exp, token)
	return err
}
