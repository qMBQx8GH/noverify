package types

import (
	"strings"
)

func IsClass(s string) bool {
	return strings.HasPrefix(s, `\`) && !IsShape(s) && !IsArray(s) && !IsClosure(s)
}

func IsShape(s string) bool {
	return strings.HasPrefix(s, `\shape$`)
}

func IsBoolean(s string) bool {
	return strings.HasPrefix(s, "bool") || strings.HasPrefix(s, "boolean")
}

func IsClosure(s string) bool {
	return strings.HasPrefix(s, `\Closure`)
}

func IsAnonClass(s string) bool {
	return strings.HasPrefix(s, `\anon$`)
}

func IsClosureFromPHPDoc(s string) bool {
	return strings.HasPrefix(s, `\Closure`) && !strings.ContainsRune(s, '.')
}

func IsArray(s string) bool {
	return strings.HasSuffix(s, `[]`)
}

func IsTrivial(s string) bool {
	return trivial[s]
}

func IsScalar(s string) bool {
	return scalar[s]
}

func IsAlias(s string) bool {
	_, has := aliases[s]
	return has
}

func IsTypeNullable(typ Map) bool {
	isNullable := false
	if typ.m == nil || typ.Empty() {
		return true // We consider that if the type is not inferred, then it is mixed
	}
	typ.Iterate(func(t string) {
		if strings.Contains(t, "null") || strings.Contains(t, "mixed") {
			isNullable = true
		}
	})
	return isNullable
}

// CanBeNull reports whether a value of the type can be null.
//
// It is stricter than IsTypeNullable, which looks for the words "null" and "mixed" in the
// type and says "true" for what is not null:
//   - an array: `mixed[]` or `null[]` is an array, it is not null even if the elements can be;
//   - a lazy type (a call of a method, a property, an element...): it describes how to
//     compute the type, the words in it are the names of the classes and the methods,
//     for example the call of `getResult()` on `mixed` has the word "mixed" in it, and
//     a method can be called `isNull()`.
//
// IsTypeNullable is still used for the parameters and for the elements of the arrays.
func CanBeNull(typ Map) bool {
	if typ.m == nil || typ.Empty() {
		return true // We consider that if the type is not inferred, then it is mixed
	}

	canBeNull := false
	typ.Iterate(func(t string) {
		if isArrayType(t) || isLazyType(t) {
			return
		}
		if strings.Contains(t, "null") || strings.Contains(t, "mixed") {
			canBeNull = true
		}
	})
	return canBeNull
}

// IsArrayMap reports whether all the types of the map are arrays (`T[]`, `array`, array shapes
// and lazy arrays) and the map is not empty.
func IsArrayMap(typ Map) bool {
	if typ.Empty() {
		return false
	}
	all := true
	typ.Iterate(func(t string) {
		if !(isArrayType(t) || t == "array" || t == "empty_array" || IsShape(t) || (len(t) > 0 && t[0] == WArrayOf)) {
			all = false
		}
	})
	return all
}

// isArrayType reports whether the type is a resolved array: `T[]`.
func isArrayType(t string) bool {
	return strings.HasSuffix(t, "[]")
}

// isLazyType reports whether the type is a lazy ("wrapped") one: it starts with the code
// of the kind (WArrayOf, WInstanceMethodCall, ...) that is a control character,
// and the names of the types never start with it.
func isLazyType(t string) bool {
	return len(t) > 0 && t[0] < ' '
}

func IsTypeMixed(typ Map) bool {
	return typ.Contains("mixed")
}

func Alias(s string) (string, bool) {
	alias, has := aliases[s]
	return alias, has
}

func ArrayType(typ string) string {
	return strings.TrimSuffix(typ, "[]")
}

var trivial = map[string]bool{
	"bool":     true,
	"callable": true,
	"float":    true,
	"int":      true,
	"mixed":    true,
	"object":   true,
	"resource": true,
	"string":   true,
	"void":     true,
	"iterable": true,
	"never":    true,

	"null":  true,
	"true":  true,
	"false": true,
}

var scalar = map[string]bool{
	"bool":   true,
	"float":  true,
	"int":    true,
	"string": true,

	"true":  true,
	"false": true,
}

var aliases = map[string]string{
	"integer": "int",
	"long":    "int",

	"boolean": "bool",

	"real":   "float",
	"double": "float",

	"callback": "callable",
}
