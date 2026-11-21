package attributes

import (
	"github.com/VKCOM/php-parser/pkg/version"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/meta"
	"github.com/VKCOM/noverify/src/types"
)

const (
	LevelTypeAware       = `\JetBrains\PhpStorm\Internal\LanguageLevelTypeAware`
	ElementAvailable     = `\JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable`
	DeprecatedAnnotation = `\JetBrains\PhpStorm\Deprecated`

	// AllowDynamicPropertiesAttribute allows to create dynamic properties of the class and
	// of its children without the deprecation of PHP 8.2.
	AllowDynamicPropertiesAttribute = `\AllowDynamicProperties`
)

// AllowDynamicProperties reports whether the class is marked with #[\AllowDynamicProperties].
func AllowDynamicProperties(groups []*ir.AttributeGroup, state *meta.ClassParseState) (found bool) {
	Each(groups, func(attr *ir.Attribute) bool {
		if Name(attr, state) == AllowDynamicPropertiesAttribute {
			found = true
			return false
		}
		return true
	})

	return found
}

func TypeAware(groups []*ir.AttributeGroup, state *meta.ClassParseState) (typ types.Map) {
	Each(groups, func(attr *ir.Attribute) bool {
		if Name(attr, state) == LevelTypeAware && len(attr.Args) > 1 {
			defaultType := NamedStringArgument(attr, "default")
			if defaultType != "" {
				typ = types.NewMap(defaultType)
				return false
			}
		}

		return true
	})

	return typ
}

// Available reports whether the element (a parameter of a stub) exists in the PHP version.
//
// The stubs describe a function with several signatures using
// #[PhpStormStubsElementAvailable(from: '5.3', to: '7.4')] and #[PhpStormStubsElementAvailable('8.0')]
// (the only positional argument is `from`). The element is available if
// from <= phpVersion <= to, a missing bound is not checked.
func Available(groups []*ir.AttributeGroup, state *meta.ClassParseState, phpVersion *version.Version) (res bool) {
	res = true

	Each(groups, func(attr *ir.Attribute) bool {
		if Name(attr, state) != ElementAvailable || len(attr.Args) == 0 {
			return true
		}

		from := NamedStringArgument(attr, "from")
		if from == "" {
			if arg := attr.Arg(0); arg.Name == nil {
				if str, ok := arg.Expr.(*ir.String); ok {
					from = str.Value
				}
			}
		}
		to := NamedStringArgument(attr, "to")

		res = versionInRange(phpVersion, from, to)
		return false
	})

	return res
}

// versionInRange reports whether from <= v <= to. The empty bounds, the bounds that
// can't be parsed and the unknown version v are not checked.
func versionInRange(v *version.Version, from, to string) bool {
	if v == nil {
		return true
	}

	if from != "" {
		if fromVersion, err := version.New(from); err == nil && v.Less(fromVersion) {
			return false
		}
	}

	if to != "" {
		// `to: '7.4'` means every 7.4.x release, so compare the major and minor parts only.
		if toVersion, err := version.New(to); err == nil && v.Greater(toVersion) {
			return false
		}
	}

	return true
}

func Deprecated(groups []*ir.AttributeGroup, state *meta.ClassParseState) (info meta.DeprecationInfo, ok bool) {
	Each(groups, func(attr *ir.Attribute) bool {
		if Name(attr, state) == DeprecatedAnnotation {
			info.Reason = NamedStringArgument(attr, "reason")
			info.Since = NamedStringArgument(attr, "since")
			info.Replacement = NamedStringArgument(attr, "replacement")

			info.Deprecated = true
		}
		return true
	})

	return info, info.Deprecated
}
