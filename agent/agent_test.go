package agent

import "testing"

type labelSuite struct {
	label  string
	parsed *parsedLabel
}

var labelSuites = []labelSuite{
	{
		label:  "service=prod:foo.bar.baz",
		parsed: &parsedLabel{service: "prod", name: "foo.bar.baz"},
	},
	{
		label:  "host=abcdefg:boo.foo.uoo",
		parsed: &parsedLabel{hostID: "abcdefg", name: "boo.foo.uoo"},
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
			if p.service != s.parsed.service || p.hostID != s.parsed.hostID || p.name != s.parsed.name {
				t.Errorf("unexpected parseLine got:%#v expected:%#v", p, s.parsed)
			}
		} else {
			if err == nil {
				t.Errorf("parseLabel(%v) must return error", s.label)
			}
		}
	}
}
