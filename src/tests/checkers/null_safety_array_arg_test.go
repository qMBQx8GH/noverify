package checkers_test

import (
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

// checkLinesOf runs the linter and compares the lines that have a report of the check
// with the lines marked with `// REPORT` in the code.
func checkLinesOf(t *testing.T, checkName, code string) {
	t.Helper()

	test := linttest.NewSuite(t)
	test.AddFile(code)

	reported := map[int]bool{}
	for _, r := range test.RunLinter().Reports {
		if r.CheckName == checkName {
			reported[r.Line] = true
		}
	}

	for i, line := range strings.Split(code, "\n") {
		want := strings.Contains(line, "// REPORT")
		if have := reported[i+1]; have != want {
			t.Errorf("%s, line %d: reported=%v, want %v: %s", checkName, i+1, have, want, strings.TrimSpace(line))
		}
	}
}

// An array is never null, whatever its elements are: the elements of the array
// that are of an unknown (mixed) type do not make the variable nullable.
//
// The parameter of a method that implements an interface has a type that is taken from
// the interface, so it does not "allow null" by itself.
func TestArrayOfMixedIsNotNull(t *testing.T) {
	checkLinesOf(t, "notNullSafetyFunctionArgumentVariable", `<?php
interface Filter {
  /** */
  public function filter();
}

class Items implements Filter {
  /** */
  public function filter() {}
  /** */
  public function add($params) {}
}

function f(Items $items, $other, ?string $maybe, $c) {
  $headers = [['id' => 'URL', 'content' => $other, 'default' => true]];
  $items->add($headers);

  $flat = [$other, $other];
  $items->add($flat);

  $nested = [[$other]];
  $items->add($nested);

  // But null is null.
  $items->add($maybe); // REPORT
  $items->add($other); // REPORT
  $x = null;
  $items->add($x); // REPORT

  // An array that can be null too.
  $a = $c ? [$other] : null;
  $items->add($a); // REPORT
}
`)
}

// A parameter without a type of a method that implements an interface takes the type from
// the interface; if the interface has no type either, any value is fine.
func TestUntypedParamOfImplementingMethod(t *testing.T) {
	checkLinesOf(t, "notSafeCall", `<?php
interface Filter {
  /** */
  public function filter();
}

class NoTypes implements Filter {
  /** */
  public function filter() {}
  /** */
  public function add($params) {}
}

interface Typed {
  /**
   * @param string $value
   */
  public function set($value);
}

class FromInterface implements Typed {
  /** */
  public function set($value) {}
}

function f(NoTypes $list, FromInterface $typed) {
  $ints = [1, 2];
  $strings = [['id' => 'URL']];
  $list->add($ints);
  $list->add($strings);

  $typed->set('text');
  $typed->set($ints); // REPORT
}
`)
}

func TestLazyTypeNamesAreNotNull(t *testing.T) {
	checkLinesOf(t, "notNullSafetyFunctionArgumentVariable", `<?php
class Res {
  public function getResult() { return 1; }
  public function isNull(): int { return 1; }
}
function take(int $x) { return $x; }
/** @param mixed $m */
function f($m, Res $r) {
  $a = $m->getResult();
  take($a);
  $b = $r->isNull();
  take($b);
}
`)
}
