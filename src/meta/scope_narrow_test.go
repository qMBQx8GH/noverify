package meta

import (
	"testing"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/types"
)

func varType(t *testing.T, sc *Scope, name string) types.Map {
	t.Helper()
	typ, ok := sc.GetVarNameType(name)
	if !ok {
		t.Fatalf("$%s is not in the scope", name)
	}
	return typ
}

func TestNarrowVarReplacesTheTypeOfPHPDocVariable(t *testing.T) {
	nullable := types.NewMap(`\R`).Append(types.NewMap("null"))

	// A variable declared with @var: ReplaceVar adds to the type, NarrowVar replaces it.
	sc := NewScope()
	sc.AddVarFromPHPDoc("x", nullable, "@var")

	sc.ReplaceVarName("x", types.NewMap(`\R`), "narrow with ReplaceVar", VarAlwaysDefined)
	if !varType(t, sc, "x").Contains("null") {
		t.Fatalf("ReplaceVar must not remove null from a @var variable (the assignment adds to it)")
	}

	sc.NarrowVar(&ir.SimpleVar{Name: "x"}, types.NewMap(`\R`), "narrow", VarAlwaysDefined)
	if typ := varType(t, sc, "x"); typ.Contains("null") || !typ.Contains(`\R`) {
		t.Errorf("NarrowVar must replace the type: %s", typ)
	}

	// It is still the variable of @var: an assignment adds to the type.
	sc.ReplaceVarName("x", types.NewMap("int"), "assign", VarAlwaysDefined)
	if typ := varType(t, sc, "x"); !typ.Contains("int") || !typ.Contains(`\R`) {
		t.Errorf("an assignment must add to the type of a @var variable: %s", typ)
	}
}

func TestNarrowVarOfAnOrdinaryVariable(t *testing.T) {
	sc := NewScope()
	sc.AddVarName("x", types.NewMap(`\R`).Append(types.NewMap("null")), "param", VarAlwaysDefined)

	sc.NarrowVarName("x", types.NewMap(`\R`), "narrow", VarAlwaysDefined)
	if varType(t, sc, "x").Contains("null") {
		t.Errorf("null must be removed")
	}

	// An ordinary variable is replaced by an assignment.
	sc.ReplaceVarName("x", types.NewMap("int"), "assign", VarAlwaysDefined)
	if typ := varType(t, sc, "x"); typ.Contains(`\R`) || !typ.Contains("int") {
		t.Errorf("an assignment must replace the type: %s", typ)
	}

	// A variable that does not exist is created, like ReplaceVar does.
	sc.NarrowVarName("y", types.NewMap("string"), "narrow", VarAlwaysDefined)
	if !varType(t, sc, "y").Contains("string") {
		t.Errorf("$y must be created")
	}

	// A node that is not a variable is ignored.
	sc.NarrowVar(&ir.String{Value: "x"}, types.NewMap("bool"), "narrow", VarAlwaysDefined)
}
