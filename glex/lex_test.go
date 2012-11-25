package glex
/*Copyright (c) 2012 André Luiz Alves Moraes

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.*/

import (
	"testing"
)

// Ensure that all errors are nil, if not,
// log all of them into t
func must(t *testing.T, errors ...error) (valid bool) {
	valid = true
	for _, e := range errors {
		if e != nil {
			valid = false
			t.Errorf("Unexpected error: %v", e)
		}
	}
	return
}

func TestRule(t *testing.T) {
	r, err := NewRule("abba", 1)
	if err != nil {
		t.Fatalf("Error compiling regexp. %v", err)
	}
	input := "abba c abba"
	m, input, err := r.Match(input)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if m.Token != 1 {
		t.Errorf("Expecting token value of %v got %v", 1, m.Token)
	}

	if m.Text != "abba" {
		t.Errorf("Expecting match text %v but got %v", "abba", m.Text)
	}

	if input != " c abba" {
		t.Errorf("Invalid tail, should be %v got %v", " c abba", input)
	}
}

func TestRuleList(t *testing.T) {
	rl := make(RuleList, 0)
	_, err := rl.NewRule("abba", 1)
	if err != nil {
		t.Fatalf("Error compiling expression. %v", err)
	}
	_, err = rl.NewRule("cba", 2)
	if err != nil {
		t.Fatalf("Erro compiling expression. %v", err)
	}
	input := "abbacba"
	m, input, err := rl.Match(input)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if m.Text != "abba" {
		t.Errorf("Expecting %v got %v", "abba", m.Text)
	}
	m, input, err = rl.Match(input)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if m.Text != "cba" {
		t.Errorf("Expecting %v got %v", "cba", m.Text)
	}
}

// Make sure that the lexer return the expected token's while scanning data
func TestLexerLiteral(t *testing.T) {
	l, err := NewLexer("abba c d")
	if err != nil {
		t.Errorf("Unable to build the parser. Cause: %v", err)
	}

	must(t, l.NewRule("abba", 1), l.NewRule("c", 2), l.NewRule("d", 3), l.NewRule("\\s+", -1))
	expected := []Match{Match{"abba", 1}, Match{"c", 2}, Match{"d", 3}}
	for _, m := range expected {
		actual, err := l.Next()
		if err != nil {
			t.Errorf("Unexpected error while lexing... %v", err)
		}
		if m.Token != actual.Token {
			t.Errorf("Expecting %v got %v", m, actual)
		}
	}
}
