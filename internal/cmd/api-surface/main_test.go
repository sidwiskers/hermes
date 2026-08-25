package main

import (
	"go/types"
	"testing"
)

func TestAliasTargetStringCanonicalizesEmptyInterface(t *testing.T) {
	t.Parallel()

	if got := aliasTargetString(types.Universe.Lookup("any").Type(), nil); got != "any" {
		t.Fatalf("aliasTargetString(any) = %q, want any", got)
	}

	empty := types.NewInterfaceType(nil, nil)
	empty.Complete()
	if got := aliasTargetString(empty, nil); got != "any" {
		t.Fatalf("aliasTargetString(interface{}) = %q, want any", got)
	}
}

func TestAliasTargetStringPreservesOtherTypes(t *testing.T) {
	t.Parallel()

	if got := aliasTargetString(types.Typ[types.String], nil); got != "string" {
		t.Fatalf("aliasTargetString(string) = %q, want string", got)
	}
}
