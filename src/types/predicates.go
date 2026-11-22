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
// It differs from IsTypeNullable for the arrays: a value of the type `mixed[]` or `null[]` is an
// array, it is not null even if the elements can be. IsTypeNullable says "true" for them,
// because it looks for the words "null" and "mixed" in the type (it is used for the elements of
// an array too, and for the parameters).
func CanBeNull(typ Map) bool {
	if typ.m == nil || typ.Empty() {
		return true // We consider that if the type is not inferred, then it is mixed
	}

	canBeNull := false
	typ.Iterate(func(t string) {
		if isArrayType(t) {
			return
		}
		if strings.Contains(t, "null") || strings.Contains(t, "mixed") {
			canBeNull = true
		}
	})
	return canBeNull
}

// isArrayType reports whether the type is an array: `T[]` or the lazy ArrayOf(T)
// (the types of the scope are lazy until they are resolved).
func isArrayType(t string) bool {
	return strings.HasSuffix(t, "[]") || (len(t) > 0 && t[0] == WArrayOf)
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
