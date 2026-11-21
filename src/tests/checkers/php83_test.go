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
