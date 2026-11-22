package types

import "testing"

func TestCanBeNull(t *testing.T) {
	tests := []struct {
		name string
		typ  Map
		want bool
	}{
		{"not inferred", Map{}, true},
		{"null", NewMap("null"), true},
		{"mixed", NewMap("mixed"), true},
		{"string or null", NewMap("string|null"), true},
		{"class or null", NewMap(`\Foo|null`), true},
		{"int", NewMap("int"), false},
		{"string", NewMap("string"), false},
		{"class", NewMap(`\Foo`), false},

		// An array is an array, whatever its elements are.
		{"mixed[]", NewMap(WrapArrayOf("mixed")), false},
		{"null[]", NewMap(WrapArrayOf("null")), false},
		{"mixed[][]", NewMap(WrapArrayOf(WrapArrayOf("mixed"))), false},
		{"string[]", NewMap(WrapArrayOf("string")), false},
		{"mixed[] resolved", NewMap("mixed[]"), false},
		{"null[] resolved", NewMap("null[]"), false},

		// But an array that can be null.
		{"mixed[] or null", NewMap(WrapArrayOf("mixed")).Append(NewMap("null")), true},
		{"array or mixed", NewMap(WrapArrayOf("int")).Append(NewMap("mixed")), true},
	}

	for _, test := range tests {
		if have := CanBeNull(test.typ); have != test.want {
			t.Errorf("%s: CanBeNull = %v, want %v", test.name, have, test.want)
		}
	}

	// IsTypeNullable does not know about the arrays (it is used for the elements and for the parameters):
	// it is the reason why CanBeNull exists.
	if !IsTypeNullable(NewMap(WrapArrayOf("mixed"))) {
		t.Errorf("IsTypeNullable is expected to look for the word mixed in the type of an array")
	}
}
