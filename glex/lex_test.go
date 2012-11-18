package glex

import (
	"testing"
)

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
