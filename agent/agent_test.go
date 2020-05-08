package agent

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type labelSuite struct {
	label  string
	parsed Label
}

var labelSuites = []labelSuite{
	{
		label: "service=prod:foo.bar.baz",
		parsed: Label{
			Service: "prod",
			Name:    "foo.bar.baz",
		},
	},
	{
		label: "host=abcdefg:boo.foo.uoo",
		parsed: Label{
			HostID: "abcdefg",
			Name:   "boo.foo.uoo",
		},
	},
	{
		label: "host=foo:hoge;emit_zero",
		parsed: Label{
			HostID:   "foo",
			Name:     "hoge",
			EmitZero: true,
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
		if s.parsed.Name != "" {
			if diff := cmp.Diff(p, s.parsed); diff != "" {
				t.Errorf("unexpected parseLine diff:%s", diff)
			}
			if s.label != p.String() {
				t.Errorf("failed to roundtrip %s to %s", s.label, p.String())
			}
		} else {
			if err == nil {
				t.Errorf("parseLabel(%v) must return error", s.label)
			}
		}
	}
}
