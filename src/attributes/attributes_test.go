package attributes

import (
	"testing"

	"github.com/VKCOM/php-parser/pkg/version"
)

func TestVersionInRange(t *testing.T) {
	tests := []struct {
		phpVersion string
		from, to   string
		want       bool
	}{
		// from
		{"8.2", "8.0", "", true},
		{"8.2", "8.2", "", true},
		{"8.2", "8.3", "", false},
		{"7.4", "8.0", "", false},
		{"7.4", "5.3", "", true},

		// to
		{"7.4", "", "7.4", true},
		{"8.2", "", "7.4", false},
		{"8.2", "", "8.2", true},
		{"8.2", "", "8.5", true},
		{"7.4", "", "7.3", false},

		// from and to
		{"7.4", "5.3", "7.4", true},
		{"8.0", "5.3", "7.4", false},
		{"8.2", "8.0", "8.2", true},
		{"8.2", "8.3", "8.5", false},
		{"8.2", "5.4", "8.0", false},
		{"5.6", "5.3", "5.6", true},

		// no bounds and the bounds that can't be parsed are not checked
		{"8.2", "", "", true},
		{"8.2", "not a version", "", true},
		{"8.2", "", "not a version", true},
	}

	for _, test := range tests {
		v, err := version.New(test.phpVersion)
		if err != nil {
			t.Fatal(err)
		}
		if have := versionInRange(v, test.from, test.to); have != test.want {
			t.Errorf("PHP %s, from %q, to %q: have %v, want %v", test.phpVersion, test.from, test.to, have, test.want)
		}
	}

	if !versionInRange(nil, "8.0", "8.1") {
		t.Errorf("the unknown PHP version must not be checked")
	}
}
