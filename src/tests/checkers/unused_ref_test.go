package checkers_test

import (
	"strings"
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

// An assignment to a parameter passed by reference is the result of the function
// for the caller, it is never an unused variable, even after an if.
func TestUnusedByRefParam(t *testing.T) {
	// The code is not clean for the other checks, only `unused` is checked here.
	have := unusedReports(t, `<?php
class A {
  /** */
  public static function afterIf(&$content) {
    if (!$content) {
      return;
    }
    $content = strtolower($content);
  }

  /** */
  public static function afterIfWithoutRead(&$out, $cond) {
    if ($cond) {
      return;
    }
    $out = 1;
  }

  /** */
  public static function afterLoop(&$out, array $items) {
    foreach ($items as $item) {
      echo $item;
    }
    $out = 1;
  }

  /** */
  public static function afterSwitchAndTry(&$out, $v) {
    switch ($v) {
      case 1:
        echo 1;
        break;
    }
    try {
      echo 2;
    } catch (\Exception $e) {
      echo $e->getMessage();
    }
    $out = 1;
  }

  /** */
  public function nested(&$out, $a, $b) {
    if ($a) {
      if ($b) {
        $out = 1;
      }
      $out = 2;
    } elseif ($b) {
      $out = 3;
    } else {
      $out = 4;
    }
    $out = 5;
  }
}

function plain(&$p, $cond) {
  if ($cond) {
    echo 1;
  }
  $p = 1;
}

$closure = function (&$p, $cond) {
  if ($cond) {
    echo 1;
  }
  $p = 1;
};

$counter = 0;
$byUse = function ($cond) use (&$counter) {
  if ($cond) {
    echo 1;
  }
  $counter = 2;
};
$byUse(true);
echo $counter;

$arrow = fn(&$p, $cond) => $cond ? ($p = 1) : ($p = 2);
`)

	if len(have) != 0 {
		t.Errorf("unexpected reports: %q", have)
	}
}

func unusedReports(t *testing.T, code string) []string {
	test := linttest.NewSuite(t)
	test.AddFile(code)

	var messages []string
	for _, r := range test.RunLinter().Reports {
		if r.CheckName == "unused" {
			messages = append(messages, r.Message)
		}
	}
	return messages
}

func TestUnusedStillReportedForLocals(t *testing.T) {
	have := unusedReports(t, `<?php
function local($cond) {
  if ($cond) {
    echo 1;
  }
  $unused = 1;
}

function byValue($param, $cond) {
  if ($cond) {
    echo 1;
  }
  $param = 1;
}

// A variable created by a reference argument in the condition is a usual local variable.
function createdByReference($subject) {
  if (preg_match('/a/', $subject, $matches)) {
    echo 1;
  }
  $matches = [];
}

$value = 1;
$byValueUse = function ($cond) use ($value) {
  if ($cond) {
    echo 1;
  }
  $value = 2;
};
$byValueUse(true);
`)

	for _, name := range []string{"unused", "param", "matches", "value"} {
		found := false
		for _, msg := range have {
			if strings.HasPrefix(msg, "Variable $"+name+" is unused") {
				found = true
			}
		}
		if !found {
			t.Errorf("$%s must be reported as unused, reports: %q", name, have)
		}
	}

	// Nothing but these variables is reported.
	for _, msg := range have {
		known := false
		for _, name := range []string{"unused", "param", "matches", "value"} {
			if strings.HasPrefix(msg, "Variable $"+name+" is unused") {
				known = true
			}
		}
		if !known {
			t.Errorf("unexpected report: %q", msg)
		}
	}
}
