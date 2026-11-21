package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestPHP82DNFTypes(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
interface A {
  /** */
  public function a(): int;
}
interface B {
  /** */
  public function b(): int;
}
class C {
  /** */
  public function a(): int { return 1; }
}

function f((A&B)|C $x): int {
  return $x->a();
}

/**
 * @param A $a
 * @param B $b
 */
function g((A&B)|null $x, A $a, B $b): (A&B)|null {
  if ($x === null) {
    return null;
  }
  return $x;
}

class Holder {
  public function __construct(private (A&B)|C|null $item = null) {}

  /** */
  public function item(): (A&B)|C|null { return $this->item; }
}

$h = new Holder();
echo $h->item() === null;
`)
}

func TestPHP82DNFTypesAreResolved(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class C {}

function f((Missing&Other)|C $x): void {}
`)
	test.Expect = []string{
		`Class or interface named \Missing does not exist`,
		`Class or interface named \Other does not exist`,
	}
	test.RunAndMatch()
}

func TestPHP82ReadonlyClassValid(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
final readonly class Point {
  public int $z;

  public function __construct(public int $x, public int $y, int $z = 0) {
    $this->z = $z;
  }

  /** */
  public function withX(int $x): self {
    return new self($x, $this->y);
  }
}

$p = new Point(1, 2);
echo $p->x, $p->y, $p->z, $p->withX(3)->x;
`)
}

func TestPHP82ReadonlyClassModification(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
readonly class Point {
  public int $z;

  public function __construct(public int $x, public int $y) {}
}

$p = new Point(1, 2);
$p->x = 3;
$p->y++;
$p->z = 4;
unset($p->x);
`)
	test.Expect = []string{
		`Cannot modify readonly property \Point::$x`,
		`Cannot modify readonly property \Point::$y`,
		`Cannot modify readonly property \Point::$z`,
		`Cannot modify readonly property \Point::$x`,
	}
	test.RunAndMatch()
}

func TestPHP82ReadonlyClassDeclaration(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
readonly class A {
  public $untyped;
  public int $withDefault = 5;
  public static int $static;

  public function __construct(public $promoted) {}
}
`)
	test.Expect = []string{
		`Readonly property must have a type`,
		`Readonly property $withDefault cannot have a default value`,
		`Static property cannot be readonly`,
		`Readonly property $promoted must have a type`,
	}
	test.RunAndMatch()
}

func TestPHP82NonReadonlyClassIsNotAffected(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Point {
  public $untyped;
  public int $withDefault = 5;
  public static int $static = 1;

  public function __construct(public int $x) {}
}

$p = new Point(1);
$p->x = 2;
$p->untyped = 3;
$p->withDefault++;
`)
}
