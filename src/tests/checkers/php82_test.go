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

func TestPHP82TraitConstants(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
trait HasStatus {
  public const DEFAULT_STATUS = 'draft';
  protected const LIMIT = 10;
  private const SECRET = 3;
  final public const FIXED = 2;

  /** */
  public function limit(): int { return self::LIMIT + static::FIXED + self::SECRET; }
}

trait Outer {
  use HasStatus;

  public const OUTER = 5;
}

class Article {
  use HasStatus;

  /** */
  public function all(): int { return self::LIMIT + static::FIXED + self::SECRET; }
}

class Post {
  use Outer;
}

class Child extends Article {}

echo Article::DEFAULT_STATUS, Article::FIXED, (new Article())->all(), (new Article())->limit();
echo Post::OUTER, Post::DEFAULT_STATUS, Post::FIXED;
echo Child::DEFAULT_STATUS, Child::FIXED;
`)
}

func TestPHP82TraitConstantsErrors(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
trait HasStatus {
  public const DEFAULT_STATUS = 'draft';
  protected const LIMIT = 10;
  private const SECRET = 3;
}

class Article {
  use HasStatus;
}

class Other {}

echo HasStatus::DEFAULT_STATUS;
echo Article::MISSING;
echo Article::LIMIT;
echo Article::SECRET;
echo Other::DEFAULT_STATUS;
`)
	test.Expect = []string{
		`Cannot access trait constant \HasStatus::DEFAULT_STATUS directly`,
		`Class constant \Article::MISSING does not exist`,
		`Cannot access protected constant \Article::LIMIT`,
		`Cannot access private constant \Article::SECRET`,
		`Class constant \Other::DEFAULT_STATUS does not exist`,
	}
	test.RunAndMatch()
}

func TestPHP82TraitConstantType(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
trait T {
  public const NAME = 'x';
}

class C {
  use T;
}

$name = C::NAME;
echo $name->foo();
`)
	test.Expect = []string{
		`Call to undefined method {string}->foo()`,
	}
	test.RunAndMatch()
}

func TestPHP82AllowDynamicProperties(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
#[\AllowDynamicProperties]
class Allowed {
  public int $a = 1;
}

/** */
class Child extends Allowed {}

/** */
class GrandChild extends Child {}

#[AllowDynamicProperties]
/** */
class GlobalNs {}

$a = new Allowed();
echo $a->dynamic;

$c = new Child();
echo $c->dynamic;

$g = new GrandChild();
echo $g->dynamic;

$n = new GlobalNs();
echo $n->dynamic;

$s = new stdClass();
echo $s->dynamic;
`)
}

func TestPHP82DynamicPropertiesWithoutAttribute(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Plain {
  public int $a = 1;
}

class Child extends Plain {}

#[\AllowDynamicProperties]
class Allowed {}

$p = new Plain();
echo $p->dynamic;

$c = new Child();
echo $c->dynamic;

$a = new Allowed();
echo $a->dynamic;
`)
	test.Expect = []string{
		`Property {\Plain}->dynamic does not exist`,
		`Property {\Child}->dynamic does not exist`,
	}
	test.RunAndMatch()
}

func TestPHP82AllowDynamicPropertiesInNamespace(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
namespace App;

// Without the leading slash it is \App\AllowDynamicProperties, which has no effect.
#[AllowDynamicProperties]
class NotApplied {}

#[\AllowDynamicProperties]
class Applied {}

$n = new NotApplied();
echo $n->dynamic;

$a = new Applied();
echo $a->dynamic;
`)
	test.Expect = []string{
		`Property {\App\NotApplied}->dynamic does not exist`,
	}
	test.RunAndMatch()
}
