package checkers_test

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

// The stubs describe the latest PHP, so the deprecation that starts in a version
// newer than the analyzed one (PHP 8.1 for the test suite) must not be reported.
func TestDeprecatedSinceVersion(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddFile(`<?php
use JetBrains\PhpStorm\Deprecated;

#[Deprecated(since: "7.4")]
function old_function(): void {}

#[Deprecated(since: "8.1")]
function current_function(): void {}

#[Deprecated(since: "8.5")]
function future_function(): void {}

#[Deprecated(since: "8.5", reason: "use something else")]
function future_function_with_reason(): void {}

#[Deprecated]
function no_version(): void {}

#[Deprecated(since: "not a version")]
function bad_version(): void {}

/** */
class A {
  #[Deprecated(since: "8.5")]
  public const FUTURE_CONST = 1;

  #[Deprecated(since: "7.0")]
  public const OLD_CONST = 1;

  /** */
  #[Deprecated(since: "8.5")]
  public static function future(): void {}

  /** */
  #[Deprecated(since: "7.0")]
  public static function old(): void {}

  /** */
  #[Deprecated(since: "8.5")]
  public function futureMethod(): void {}

  /** */
  #[Deprecated(since: "7.0")]
  public function oldMethod(): void {}
}

function f(A $a): void {
  old_function();
  current_function();
  future_function();
  future_function_with_reason();
  no_version();
  bad_version();

  echo A::FUTURE_CONST, A::OLD_CONST;
  A::future();
  A::old();
  $a->futureMethod();
  $a->oldMethod();
}
`)
	test.Expect = []string{
		`Call to deprecated function old_function (since: 7.4)`,
		`Call to deprecated function current_function (since: 8.1)`,
		`Call to deprecated function no_version`,
		`Call to deprecated function bad_version (since: not a version)`,
		`Try to call constant OLD_CONST that was marked as deprecated (since: 7.0) from the class \A`,
		`Call to deprecated static method \A::old() (since: 7.0)`,
		`Call to deprecated method {\A}->oldMethod() (since: 7.0)`,
	}
	test.RunAndMatch()
}
