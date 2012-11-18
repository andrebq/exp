package glex

import (
	"testing"
)

func TestRule(t *testing.T) {
	r, err := NewRule("abba",1)
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
