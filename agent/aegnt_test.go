package agent

import "testing"

type labelSuite struct {
	label   string
	service string
	hostID  string
	name    string
	isValid bool
}

var labelSuites = []labelSuite{
	labelSuite{
		"service=prod:foo.bar.baz",
		"prod", "", "foo.bar.baz",
		true,
	},
	labelSuite{
		"host=abcdefg:boo.foo.uoo",
		"", "abcdefg", "boo.foo.uoo",
		true,
	},
	labelSuite{
		"zzz:foo.bar.baz",
		"", "", "",
		false,
	},
	labelSuite{
		"zzz=goo:foo.bar.baz",
		"", "", "",
		false,
	},
	labelSuite{
		"foo.bar.baz",
		"", "", "",
		false,
	},
}

func TestParseLabel(t *testing.T) {
	for _, s := range labelSuites {
		service, hostID, name, err := parseLabel(s.label)
		if s.isValid {
			if s.service != service || s.hostID != hostID || s.name != name {
				t.Errorf("unexpected parseLine result: %#v", s)
			}
		} else {
			if err == nil {
				t.Errorf("parseLabel(%v) must return error", s.label)
			}
		}
	}
}
