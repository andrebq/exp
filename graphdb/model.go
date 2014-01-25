package graphdb

import (
	"database/sql"
	"github.com/lib/pq/hstore"
	"strings"
)

// Attributes define the structure used to represent
// the attributes of a given node.
//
// This map's directly to the hstore column type of postgres,
// with some pre-validation
type Attributes struct {
	data *hstore.Hstore
}

// Put set the value of the given keywork to
// the specified attribute
func NewAttributes() *Attributes {
	return &Attributes{
		data: &hstore.Hstore{
			Map: make(map[string]sql.NullString)}}
}

// Put save the given value under the given keyword
func (a *Attributes) Put(keyword, value string) {
	if !strings.HasPrefix(keyword, ":") {
		keyword = ":" + keyword
	}
	a.data.Map[keyword] = sql.NullString{value, true}
}

// Get return the value at the keyword and a boolean.
//
// If the boolean is true, then the key was found, otherwise
// the key wasn't in the attributes or the value was null
//
//	a.Put(":this/is/a/key/with/namespace", "abc123")
//	value, has := a.Get(":this/is/a/key/with/namespace")
//	if has { /* the value was found */ }
//	else { /* value wasn't fount */ }
//
func (a *Attributes) Get(keyword string) (string, bool) {
	if val, has := a.data.Map[keyword]; has {
		if val.Valid {
			return val.String, true
		}
		return "", false
	}
	return "", false
}
