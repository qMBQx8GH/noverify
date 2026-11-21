<?php

use JetBrains\PhpStorm\Internal\PhpStormStubsElementAvailable;

// The tests are run for PHP 8.1.

function f(
  int $hour,
  #[PhpStormStubsElementAvailable(from: '8.0')] int $seconds,
) {}

function f1(
  int $hour,
  #[PhpStormStubsElementAvailable(to: '7.4')] int $seconds,
) {}

function f2(
  int $hour,
  #[PhpStormStubsElementAvailable('8.0')] int $seconds,
) {}

// The same parameter with a different signature for the different versions (like fsockopen() in the stubs).
function f3(
  int $hour,
  #[PhpStormStubsElementAvailable(from: '5.3', to: '7.4')] int $old,
  #[PhpStormStubsElementAvailable(from: '8.0')] int $new,
  int $last,
) {}

// The bounds are inclusive.
function f4(
  #[PhpStormStubsElementAvailable(from: '8.1')] int $a,
  #[PhpStormStubsElementAvailable(to: '8.1')] int $b,
  #[PhpStormStubsElementAvailable(from: '8.1', to: '8.1')] int $c,
) {}

// Not yet available / not available any more.
function f5(
  #[PhpStormStubsElementAvailable(from: '8.2')] int $a,
  #[PhpStormStubsElementAvailable(to: '8.0')] int $b,
  #[PhpStormStubsElementAvailable(from: "5.4", to: "8.0")] int $c,
  int $d,
) {}

function main() {
  f(); // want `Too few arguments for f, expecting 2, saw 0`
  f1(); // want `Too few arguments for f1, expecting 1, saw 0`
  f2(); // want `Too few arguments for f2, expecting 2, saw 0`
  f3(); // want `Too few arguments for f3, expecting 3, saw 0`
  f4(); // want `Too few arguments for f4, expecting 3, saw 0`
  f5(); // want `Too few arguments for f5, expecting 1, saw 0`
}
