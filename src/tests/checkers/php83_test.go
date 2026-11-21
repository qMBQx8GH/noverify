package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestPHP83TypedClassConstants(t *testing.T) {
	linttest.SimpleNegativeTest(t, `<?php
/** */
class A {
  public const int NUM = 1;
  public const ?string NAME = null;
  public const int|string MIXED = 1;
  final protected const array LIST = [1, 2];
  const FLOAT_C = 1.5;

  /** @return int[] */
  public function list(): array { return self::LIST; }
}

trait T {
  public const bool FLAG = true;
}

/** */
class B {
  use T;
}

interface I {
  public const mixed ANY = 1;
}

enum E: int {
  const int DEFAULT = 1;

  case One = 1;
}

echo A::NUM, A::NAME, A::MIXED, A::FLOAT_C, B::FLAG, I::ANY, E::DEFAULT;
echo count((new A())->list());
`)
}

func TestPHP83TypedClassConstantTypes(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
class A {
  public const int NUM = 1;
  public const ?string NAME = null;
  public const float RATIO = 1;
}

trait T {
  public const bool FLAG = true;
}

class B {
  use T;
}

echo A::NUM->nope();
echo A::NAME->nope();
echo A::RATIO->nope();
echo B::FLAG->nope();
echo A::MISSING;
`)
	test.Expect = []string{
		`Call to undefined method {int}->nope()`,
		`Call to undefined method {null|string}->nope()`,
		`Call to undefined method {float}->nope()`,
		`Call to undefined method {bool}->nope()`,
		`Class constant \A::MISSING does not exist`,
	}
	test.RunAndMatch()
}

func TestPHP83DynamicClassConstantFetch(t *testing.T) {
	test := linttest.NewPHP83Suite(t)
	test.AddFile(`<?php
class Config {
  const VERSION = '8.3.0';

  /** */
  public static function byName(string $name): string {
    return static::{$name} . self::{'VER' . 'SION'};
  }
}

$name = 'VERSION';
$config = new Config();

echo Config::{$name};
echo Config::{'VER' . 'SION'};
echo $config::{$name};
echo Config::byName($name);

// The expression is analyzed.
echo Config::{$undefinedVariable};
// The old syntax is still checked.
echo Config::MISSING;
`)
	test.Expect = []string{
		`Cannot find referenced variable $undefinedVariable`,
		`Class constant \Config::MISSING does not exist`,
	}
	test.RunAndMatch()
}

func TestPHP83AnonymousReadonlyClass(t *testing.T) {
	test := linttest.NewPHP83Suite(t)
	test.AddFile(`<?php
$valid = new readonly class(5) {
  public function __construct(public int $x) {}
};
echo $valid->x;

$plain = new class(5) {
  public function __construct(public int $x) {}
};
$plain->x = 6;

$modified = new readonly class(5) {
  public function __construct(public int $x) {}
};
$modified->x = 6;
`)
	test.Expect = []string{
		`Cannot modify readonly property`,
	}
	test.RunAndMatch()
}

func TestPHP83Override(t *testing.T) {
	test := linttest.NewPHP83Suite(t)
	test.AddFile(`<?php
interface I {
  /** */
  public function fromInterface(): void;
}

interface J extends I {
  /** */
  public function fromJ(): void;
}

abstract class P implements J {
  /** */
  public function __construct() {}
  /** */
  public function parentMethod(): void {}
  /** */
  private function secret(): void {}
  /** */
  abstract public function abstractMethod(): void;
  /** */
  public function fromInterface(): void {}
}

trait T {
  /** */
  public function implemented(): void {}
  /** */
  abstract public function required(): void;
}

class C extends P {
  use T;

  /** */
  #[\Override]
  public function parentMethod(): void {}
  /** */
  #[\Override]
  public function fromInterface(): void {}
  /** */
  #[\Override]
  public function fromJ(): void {}
  /** */
  #[\Override]
  public function abstractMethod(): void {}
  /** */
  #[\Override]
  public function required(): void {}
  /** */
  #[\Override]
  public function __toString(): string { return ''; }

  /** */
  #[\Override]
  public function typo(): void {}
  /** */
  #[\Override]
  public function secret(): void {}
  /** */
  #[\Override]
  public function implemented(): void {}
  /** */
  #[\Override]
  public function __construct() { parent::__construct(); }
  /** */
  public function withoutAttribute(): void {}
}

class NoParent {
  /** */
  #[\Override]
  public function a(): void {}
}

// The unknown parent: it is not known if the method overrides something.
class Unknown extends Missing {
  /** */
  #[\Override]
  public function a(): void {}
}

enum E implements I {
  case A;

  /** */
  #[\Override]
  public function fromInterface(): void {}
  /** */
  #[\Override]
  public function nope(): void {}
}
`)
	test.Expect = []string{
		`\C::typo() has #[\Override] attribute, but no matching parent method exists`,
		`\C::secret() has #[\Override] attribute, but no matching parent method exists`,
		`\C::implemented() has #[\Override] attribute, but no matching parent method exists`,
		`\C::__construct() has #[\Override] attribute, but no matching parent method exists`,
		`\NoParent::a() has #[\Override] attribute, but no matching parent method exists`,
		`\E::nope() has #[\Override] attribute, but no matching parent method exists`,
		`Class or interface named \Missing does not exist`,
	}
	test.RunAndMatch()
}

func TestPHP83ReadonlyInClone(t *testing.T) {
	code := `<?php
class A {
  /**
   * @param int[]  $items
   * @param object $o
   * @param int    $n
   * @param int    $m
   */
  public function __construct(
    public readonly array $items,
    public readonly object $o,
    public readonly int $n,
    public readonly int $m,
  ) {}

  public function __clone() {
    $this->o = clone $this->o;
    $this->n++;
    $this->m += 1;
    unset($this->items);
    $this->items[] = 1;
  }

  /** */
  public function notClone(): void {
    $this->n++;
    unset($this->m);
  }
}
`

	t.Run("php8.3", func(t *testing.T) {
		test := linttest.NewPHP83Suite(t)
		test.AddFile(code)
		test.Expect = []string{
			`Cannot modify readonly property \A::$items`,
			`Cannot modify readonly property \A::$n`,
			`Cannot modify readonly property \A::$m`,
		}
		test.RunAndMatch()
	})

	t.Run("php8.1", func(t *testing.T) {
		test := linttest.NewSuite(t)
		test.AddFile(code)
		test.Expect = []string{
			`Cannot modify readonly property \A::$n`,
			`Cannot modify readonly property \A::$m`,
			`Cannot modify readonly property \A::$items`,
			`Cannot modify readonly property \A::$items`,
			`Cannot modify readonly property \A::$n`,
			`Cannot modify readonly property \A::$m`,
		}
		test.RunAndMatch()
	})
}

func TestPHP83TypedClassConstantValues(t *testing.T) {
	test := linttest.NewPHP83Suite(t)
	test.AddFile(`<?php
class A {
  const int OK_INT = 1;
  const float OK_FLOAT = 1;
  const float OK_FLOAT_EXPR = 1 + 2;
  const string OK_STRING = 'a' . 'b';
  const bool OK_BOOL = true;
  const ?int OK_NULL = null;
  const array OK_ARRAY = [1, 2];
  const iterable OK_ITERABLE = [];
  const int|string OK_UNION = 'a';
  const mixed OK_MIXED = null;
  const false OK_FALSE = false;
  const int OK_BUILTIN = PHP_INT_MAX;
  const int OK_UNKNOWN = self::OK_INT;

  const int BAD_STRING = '1';
  const int BAD_FLOAT = 1.5;
  const float BAD_FLOAT_STRING = '1.5';
  const string BAD_INT = 1;
  const bool BAD_BOOL = 1;
  const int BAD_NULL = null;
  const array BAD_ARRAY = 1;
  const int BAD_ARRAY_VALUE = [];
  const int|string BAD_UNION = 1.5;
  const true BAD_TRUE = false;
  const string FIRST = 'a', SECOND = 2;

  const void INVALID_VOID = null;
  const never INVALID_NEVER = null;
  const static INVALID_STATIC = null;
  const callable INVALID_CALLABLE = 'strlen';
}
`)
	test.Expect = []string{
		`Cannot use string as value for class constant \A::BAD_STRING of type int`,
		`Cannot use float as value for class constant \A::BAD_FLOAT of type int`,
		`Cannot use string as value for class constant \A::BAD_FLOAT_STRING of type float`,
		`Cannot use int as value for class constant \A::BAD_INT of type string`,
		`Cannot use int as value for class constant \A::BAD_BOOL of type bool`,
		`Cannot use null as value for class constant \A::BAD_NULL of type int`,
		`Cannot use int as value for class constant \A::BAD_ARRAY of type array`,
		`Cannot use array as value for class constant \A::BAD_ARRAY_VALUE of type int`,
		`Cannot use float as value for class constant \A::BAD_UNION of type int|string`,
		`Cannot use bool as value for class constant \A::BAD_TRUE of type true`,
		`Cannot use int as value for class constant \A::SECOND of type string`,
		`Class constant \A cannot have type void`,
		`Class constant \A cannot have type never`,
		`Class constant \A cannot have type static`,
		`Class constant \A cannot have type callable`,
	}
	test.RunAndMatch()
}
