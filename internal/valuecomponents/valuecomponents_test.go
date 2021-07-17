package valuecomponents

import (
	"testing"

	"github.com/go-test/deep"
)

func TestParseToken(t *testing.T) {
	tab := map[string]string{
		"foo":     "foo",
		"foo 99":  "foo",
		"   asd ": "",
	}

	for in, exp := range tab {
		act := ParseToken(in)
		if act != exp {
			t.Errorf("'%s': expected '%s' but got '%s'", in, exp, act)
		}
	}
}

func TestParseQuotedString(t *testing.T) {
	tab := map[string]string{
		`foo`:      "",
		`"foo" 99`: "foo",
		`  "asd" `: "",
	}

	for in, exp := range tab {
		act, err := ParseQuotedString(in)
		if err != nil {
			t.Error(err)
		}
		if act != exp {
			t.Errorf("'%s': expected '%s' but got '%s'", in, exp, act)
		}
	}
}

func TestParseFieldValueComponents(t *testing.T) {
	tab := map[string]ValueList{
		`foo`: {
			Value{
				Primary: "foo",
				Pairs:   map[string]string{},
			},
		},
		`"foo"`: {
			Value{
				Primary: "foo",
				Pairs:   map[string]string{},
			},
		},
		`foo, bar`: {
			Value{
				Primary: "foo",
				Pairs:   map[string]string{},
			},
			Value{
				Primary: "bar",
				Pairs:   map[string]string{},
			},
		},
		`foo; charset=UTF-8`: {
			Value{
				Primary: "foo",
				Pairs: map[string]string{
					"charset": "UTF-8",
				},
			},
		},
		`proto=https; host=example.com; for=5.6.7.84, for=5.6.7.8; proto=http`: {
			Value{
				Primary: "",
				Pairs: map[string]string{
					"proto": "https",
					"host":  "example.com",
					"for":   "5.6.7.84",
				},
			},
			Value{
				Primary: "",
				Pairs: map[string]string{
					"proto": "http",
					"for":   "5.6.7.8",
				},
			},
		},
	}

	for in, exp := range tab {
		act, err := ParseValueList(in)
		if err != nil {
			t.Errorf("'%s': got error %s", in, err)
		} else if diff := deep.Equal(exp, act); diff != nil {
			t.Errorf("'%s': got diff: %s", in, diff)
		}
	}
}
