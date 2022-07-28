package valuecomponents

import (
	"testing"

	"github.com/halimath/expect-go"
)

func TestParseToken(t *testing.T) {
	tab := map[string]string{
		"foo":     "foo",
		"foo 99":  "foo",
		"   asd ": "",
	}

	for in, want := range tab {
		got := ParseToken(in)
		expect.That(t, got).Is(expect.Equal(want))
	}
}

func TestParseQuotedString(t *testing.T) {
	tab := map[string]string{
		`foo`:      "",
		`"foo" 99`: "foo",
		`  "asd" `: "",
	}

	for in, want := range tab {
		got, err := ParseQuotedString(in)
		expect.That(t, err, expect.StopImmediately{}).Is(expect.NoError())
		expect.That(t, got).Is(expect.Equal(want))
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

	for in, want := range tab {
		got, err := ParseValueList(in)
		expect.That(t, err, expect.StopImmediately{}).Is(expect.NoError())
		expect.That(t, got).Is(expect.DeepEqual(want))
	}
}
