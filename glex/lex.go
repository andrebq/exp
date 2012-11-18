package glex

import (
	"regexp"
	"strings"
)

// Lexer errors
type Error string

// Error interface
func (e Error) Error() string { return string(e) }

// Stringer interface
func (e Error) String() string { return string(e) }

const (
	NoMatch = Error("No match was found in the list")
	Eof     = Error("End of input")
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
