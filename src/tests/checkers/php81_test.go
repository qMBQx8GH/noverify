package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestPHP81EnumMembers(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
enum Suit: string {
  case Hearts = 'H';
  case Spades = 'S';

  const Default = self::Hearts;

  /** */

  public function label(): string { return ucfirst($this->name) . $this->value; }

  /** */

  public static function make(): self { return self::Hearts; }
}

function f(Suit $s): string { return $s->value; }

echo Suit::Hearts->value;
echo Suit::Hearts->name;
echo Suit::Hearts->label();
echo f(Suit::Spades);
echo Suit::from('H')->label();
echo Suit::tryFrom('X')?->label();
echo count(Suit::cases());
echo Suit::Default->name;
echo Suit::make()->name;
`)
}

func TestPHP81EnumPure(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
enum Status {
  case Active;
  case Inactive;
}

function f(Status $s): string { return $s->name; }

echo f(Status::Active);
echo count(Status::cases());
`)
}

func TestPHP81EnumInterface(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
interface HasLabel {
  /** */
  public function label(): string;
}

enum Level: int implements HasLabel {
  case Low = 1;
  case High = 2;

  /** */

  public function label(): string { return $this->name; }
}

function f(HasLabel $l): string { return $l->label(); }

echo f(Level::Low);
echo Level::from(2)->value;
`)
}

func TestPHP81EnumErrors(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
enum Suit: string {
  case Hearts = 'H';
}

echo Suit::Clubs->value;
echo Suit::Hearts->nope;
echo Suit::missing();
`)
	test.Expect = []string{
		`Class constant \Suit::Clubs does not exist`,
		`Property {\Suit}->nope does not exist`,
		`Call to undefined method \Suit::missing()`,
	}
	test.RunAndMatch()
}

func TestPHP81PureEnumHasNoValue(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
enum Status {
  case Active;
}

echo Status::Active->value;
echo Status::from('a');
`)
	test.Expect = []string{
		`Property {\Status}->value does not exist`,
		`Call to undefined method \Status::from()`,
	}
	test.RunAndMatch()
}

func TestPHP81FirstClassCallable(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function foo(int $a): int { return $a; }

class A {
  /** */
  public function m(int $x): int { return $x; }
  /** */
  public static function s(int $x): int { return $x; }
}

$a = new A();
$f = foo(...);
$g = $a->m(...);
$h = A::s(...);
$i = strlen(...);

echo $f(1), $g(2), $h(3), $i('abc');
`)
}

func TestPHP81FirstClassCallableDoesNotHideRealCalls(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function foo(int $a): int { return $a; }

$f = foo(...);
$r = foo();
echo $f(1), $r;
`)
	test.Expect = []string{
		`Too few arguments for foo, expecting 1, saw 0`,
	}
	test.RunAndMatch()
}

func TestPHP81ReadonlyValid(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
class Point {
  public readonly int $x;

  public function __construct(public readonly int $y, int $x) {
    $this->x = $x;
  }

  /** */

  public function withY(int $y): self {
    return new self($y, $this->x);
  }
}

$p = new Point(1, 2);
echo $p->x, $p->y, $p->withY(3)->y;
`)
}

func TestPHP81ReadonlyModification(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class Point {
  /**
   * @param int   $x
   * @param int[] $items
   */
  public function __construct(public readonly int $x, public readonly array $items) {}

  /** */

  public function inc(): void {
    $this->x++;
  }
}

$p = new Point(1, []);
$p->x = 2;
$p->x += 2;
$p->x--;
$p->items[] = 1;
unset($p->x);
`)
	test.Expect = []string{
		`Cannot modify readonly property \Point::$x`,
		`Cannot modify readonly property \Point::$x`,
		`Cannot modify readonly property \Point::$x`,
		`Cannot modify readonly property \Point::$x`,
		`Cannot modify readonly property \Point::$items`,
		`Cannot modify readonly property \Point::$x`,
	}
	test.RunAndMatch()
}

func TestPHP81ReadonlyDeclaration(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
  public readonly $untyped;
  public readonly int $withDefault = 5;
  public static readonly int $static;

  public function __construct(public readonly $promoted) {}
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

func TestPHP81IntersectionTypes(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
interface A {
  /** */
  public function a(): int;
}
interface B {
  /** */
  public function b(): int;
}

function f(A&B $x): int { return $x->a() + $x->b(); }
`)
}

func TestPHP82ReadonlyClass(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
final readonly class Point {
  public function __construct(public int $x, public int $y) {}
}

$p = new Point(1, 2);
echo $p->x, $p->y;
`)
}

func TestPHP81EnumReservedCaseNames(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
enum Mode: int {
  case Use = 1;
  case Empty = 2;
  case Default = 3;
  case List = 4;
  case Static = 5;
}

echo Mode::Use->value, Mode::Empty->name, Mode::from(3)->name;
echo count(Mode::cases());
`)
}

func TestPHP81EnumAsIdentifier(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
/** */
class Enum {
  const X = 1;
}

/** */
class Child extends Enum {}

echo Enum::X, (new Child())::X;
`)
}

func TestPHP81FirstClassCallableIsClosure(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
function foo(int $a): int { return $a; }

class A {
  /** */
  public function m(int $x): string { return (string)$x; }
  /** */
  public static function s(int $x): bool { return $x > 0; }
}

function takesClosure(\Closure $c): void { echo $c(1); }

$a = new A();
$f = strlen(...);
$g = foo(...);
$h = $a->m(...);
$i = A::s(...);

$bound = $f->bindTo(null);
echo $g->call(new A(), 1);
echo $h->bindTo($a) !== null;
echo $i instanceof \Closure;

takesClosure($f);
takesClosure($g);
takesClosure($h);
takesClosure($i);
`)
}

func TestPHP81FirstClassCallableTypeErrors(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
function foo(int $a): int { return $a; }

class A {
  /** */
  public function m(int $x): string { return (string)$x; }
  /** */
  public static function s(int $x): bool { return $x > 0; }
}

$a = new A();

// These are the closures, not the results of the calls.
$f = foo(...);
echo $f->nope();
$g = $a->m(...);
echo $g->nope();
$h = A::s(...);
echo $h->nope();

// A real call has the type of the returned value.
$x = foo(1);
echo $x->nope();
$y = $a->m(1);
echo $y->nope();
$z = A::s(1);
echo $z->nope();
`)
	test.Expect = []string{
		`Call to undefined method {\Closure}->nope()`,
		`Call to undefined method {\Closure}->nope()`,
		`Call to undefined method {\Closure}->nope()`,
		`Call to undefined method {int}->nope()`,
		`Call to undefined method {string}->nope()`,
		`Call to undefined method {bool}->nope()`,
	}
	test.RunAndMatch()
}
