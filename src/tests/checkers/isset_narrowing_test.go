package checkers_test

import (
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

// checkNullSafetyLines runs the linter and compares the lines with a notNullSafetyVariable
// report with the lines marked with `// REPORT` in the code.
func checkNullSafetyLines(t *testing.T, code string) {
	t.Helper()

	test := linttest.NewSuite(t)
	test.AddFile(code)

	reported := map[int]bool{}
	for _, r := range test.RunLinter().Reports {
		if r.CheckName == "notNullSafetyVariable" {
			reported[r.Line] = true
		}
	}

	for i, line := range strings.Split(code, "\n") {
		want := strings.Contains(line, "// REPORT")
		if have := reported[i+1]; have != want {
			t.Errorf("line %d: reported=%v, want %v: %s", i+1, have, want, strings.TrimSpace(line))
		}
	}
}

// isset() is false for null, so the variable is not null inside of the if.
func TestIssetRemovesNull(t *testing.T) {
	checkNullSafetyLines(t, `<?php
class R {
  public ?R $next = null;
  /** @return int */
  public function m() { return 1; }
}

function plain(?R $x) { if (isset($x)) { echo $x->m(); } }
function withAnd(?R $x, $y) { if (isset($x) && $y) { echo $x->m(); } }
function inCondition(?R $x) { if (isset($x) && $x->m() > 0) { return 1; } return 0; }
function ternary(?R $x) { return isset($x) ? $x->m() : 0; }
function twoVariables(?R $x, ?R $z) { if (isset($x, $z)) { echo $x->m(), $z->m(); } }
function property(?R $x) { if (isset($x->next)) { echo $x->m(); } }
function arrayElement(?array $arr) { if (isset($arr['k'])) { echo count($arr); } }
function elseBranchOfNegated(?R $x) { if (!isset($x)) { echo 1; } else { echo $x->m(); } }
function inElseif(?R $x, $y) { if ($y) { echo 1; } elseif (isset($x)) { echo $x->m(); } }

// The inner if/else does not bring null back.
function innerIfElse(?R $x, $c) {
  if (isset($x)) {
    if ($c) { $a = 1; } else { $b = 2; }
    echo $x->m(), $a ?? 0, $b ?? 0;
  }
}

// The shape of the real code: a variable from an array that can be null.
/** @param R[] $arr */
function fromArray(array $arr, $c) {
  /** @var R|null $x */
  $x = $arr['k'] ?? null;
  if (isset($x)) {
    if ($c) { $a = 1; } else { $b = 2; }
    echo $x->m(), $a ?? 0, $b ?? 0;
  }
}

// null is still possible here.
function noCheck(?R $x) { echo $x->m(); } // REPORT
function afterIf(?R $x) { if (isset($x)) { echo 1; } echo $x->m(); } // REPORT
function elseBranch(?R $x) { if (isset($x)) { echo 1; } else { echo $x->m(); } } // REPORT
function orOperator(?R $x, $y) { if (isset($x) || $y) { echo $x->m(); } } // REPORT
function negatedAnd(?R $x, $y) { if (!isset($x) && $y) { echo $x->m(); } } // REPORT
function negatedAndAfter(?R $x, $y) { if (!isset($x) && $y) { return; } echo $x->m(); } // REPORT
function otherVariable(?R $x, ?R $z) { if (isset($z)) { echo $x->m(); } } // REPORT
function onlyOneOfTwo(?R $x, ?R $z) { if (isset($x) || isset($z)) { echo $x->m(); } } // REPORT
`)
}

// The other forms of the check work the same way, they are the reference for isset().
func TestNotNullChecksRemoveNull(t *testing.T) {
	checkNullSafetyLines(t, `<?php
class R {
  /** @return int */
  public function m() { return 1; }
}

function notNull(?R $x) { if ($x !== null) { echo $x->m(); } }
function truthy(?R $x) { if ($x) { echo $x->m(); } }
function instanceOf1(?R $x) { if ($x instanceof R) { echo $x->m(); } }
function ternary(?R $x) { return $x ? $x->m() : 0; }
function inCondition(?R $x) { if ($x !== null && $x->m() > 0) { return 1; } return 0; }
`)
}

// A variable declared with @var keeps its type when it is assigned (the assigned type is added to it),
// but a check still narrows it inside of the branch.
func TestNarrowingOfVarDocVariables(t *testing.T) {
	checkNullSafetyLines(t, `<?php
class R {
  /** @return int */
  public function m() { return 1; }
}

/** @param R[] $arr */
function viaIsset(array $arr) {
  /** @var R|null $x */
  $x = $arr['k'] ?? null;
  if (isset($x)) { echo $x->m(); }
}

/** @param R[] $arr */
function viaNotNull(array $arr) {
  /** @var R|null $x */
  $x = $arr['k'] ?? null;
  if ($x !== null) { echo $x->m(); }
}

/** @param R[] $arr */
function viaTruthiness(array $arr) {
  /** @var R|null $x */
  $x = $arr['k'] ?? null;
  if ($x) { echo $x->m(); }
}

/** @param R[] $arr */
function viaInstanceof(array $arr) {
  /** @var R|null $x */
  $x = $arr['k'] ?? null;
  if ($x instanceof R) { echo $x->m(); }
}

/** @param R[] $arr */
function elseBranchOfNotNull(array $arr) {
  /** @var R|null $x */
  $x = $arr['k'] ?? null;
  if ($x === null) { echo 1; } else { echo $x->m(); }
}

/** @param R[] $arr */
function elseBranchOfIsNull(array $arr) {
  /** @var R|null $x */
  $x = $arr['k'] ?? null;
  if (is_null($x)) { echo 1; } else { echo $x->m(); }
}

/** @param R[] $arr */
function issetWithInnerIf(array $arr, $c) {
  /** @var R|null $x */
  $x = $arr['k'] ?? null;
  if (isset($x)) {
    if ($c) { $a = 1; } else { $b = 2; }
    echo $x->m(), $a ?? 0, $b ?? 0;
  }
}

/** @param R[] $arr */
function noCheck(array $arr) {
  /** @var R|null $x */
  $x = $arr['k'] ?? null;
  echo $x->m(); // REPORT
}

/** @param R[] $arr */
function afterIf(array $arr) {
  /** @var R|null $x */
  $x = $arr['k'] ?? null;
  if (isset($x)) { echo 1; }
  echo $x->m(); // REPORT
}

/** @param R[] $arr */
function assignedInBranch(array $arr) {
  /** @var R|null $x */
  $x = $arr['k'] ?? null;
  if (isset($x)) {
    $x = null; // the assigned type is added: R|null
    echo $x->m(); // REPORT
  }
}

/** @param R[] $arr */
function elseBranch(array $arr) {
  /** @var R|null $x */
  $x = $arr['k'] ?? null;
  if (isset($x)) { echo 1; } else { echo $x->m(); } // REPORT
}
`)
}
