package phpgrep

import (
	"testing"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/conf"
	phperrors "github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
)

// mustParsePHP8 parses the expression like the linter parses the analyzed code.
// (Patterns are still parsed as PHP 7.4, see parseutil.)
func mustParsePHP8(t *testing.T, expr string) ir.Node {
	var errs []*phperrors.Error
	root, err := parser.Parse([]byte("<?php "+expr+";"), conf.Config{
		Version:          &version.Version{Major: 8, Minor: 2},
		ErrorHandlerFunc: func(e *phperrors.Error) { errs = append(errs, e) },
	})
	if err != nil || len(errs) != 0 {
		t.Fatalf("parse `%s`: %v %v", expr, err, errs)
	}
	stmt := irconv.ConvertNode(root.(*ast.Root).Stmts[0])
	if n, ok := stmt.(*ir.ExpressionStmt); ok {
		return n.Expr
	}
	return stmt
}

// Captured nodes are compared with eqNode, so the nodes of the analyzed code can get
// there. It used to panic on the nullsafe property fetch and method call (`?->`).
func TestMatchPHP8Nodes(t *testing.T) {
	var c Compiler

	run := func(want bool, tests []*matcherTest) {
		for _, test := range tests {
			m := mustCompile(t, &c, test.pattern)
			_, have := m.Match(mustParsePHP8(t, test.input))
			if have != want {
				t.Errorf("match results mismatch:\npattern: %q\ninput: %q\nhave: %v\nwant: %v",
					test.pattern, test.input, have, want)
			}
		}
	}

	run(true, []*matcherTest{
		{`$x ? $x : $y`, `$a?->b ? $a?->b : 1`},
		{`$x ? $x : $y`, `$a?->b() ? $a?->b() : 1`},
		{`$x ? $x : $y`, `$a?->b(1, $c) ? $a?->b(1, $c) : 1`},
		{`$x ? $x : $y`, `$a?->b?->c ? $a?->b?->c : 1`},
		{`$x ? $x : $y`, `$a?->B() ? $a?->b() : 1`},
		{`$x ?? $x`, `$o?->method() ?? $o?->method()`},
	})

	run(false, []*matcherTest{
		{`$x ? $x : $y`, `$a?->b ? $a?->c : 1`},
		{`$x ? $x : $y`, `$a?->b ? $a->b : 1`},
		{`$x ? $x : $y`, `$a->b ? $a?->b : 1`},
		{`$x ? $x : $y`, `$a?->b() ? $a?->b(1) : 1`},
		{`$x ? $x : $y`, `$a?->b() ? $c?->b() : 1`},
	})
}
