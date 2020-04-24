package agent

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type labelSuite struct {
	label  string
	parsed *ParsedLabel
}

var labelSuites = []labelSuite{
	{
		label: "service=prod:foo.bar.baz",
		parsed: &ParsedLabel{
			Service: "prod",
			Name:    "foo.bar.baz",
			Options: map[string]struct{}{},
		},
	},
	{
		label: "host=abcdefg:boo.foo.uoo",
		parsed: &ParsedLabel{
			HostID:  "abcdefg",
			Name:    "boo.foo.uoo",
			Options: map[string]struct{}{},
		},
	},
	{
		label: "host=foo:hoge;emit_zero",
		parsed: &ParsedLabel{
			HostID: "foo",
			Name:   "hoge",
			Options: map[string]struct{}{
				"emit_zero": {},
			},
		},
	},
	{
		label: "zzz:foo.bar.baz",
	},
	{
		label: "zzz=goo:foo.bar.baz",
	},
	{
		label: "foo.bar.baz",
	},
}

func TestParseLabel(t *testing.T) {
	for _, s := range labelSuites {
		p, err := parseLabel(s.label)
		if s.parsed != nil {
			if diff := cmp.Diff(p, s.parsed); diff != "" {
				t.Errorf("unexpected parseLine diff:%s", diff)
			}
		} else {
			if err == nil {
				t.Errorf("parseLabel(%v) must return error", s.label)
			}
		}
	}
}
