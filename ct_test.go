package main

import "testing"

func Assert(t *testing.T, what string, assertion bool) {
	if !assertion {
		t.Log(what + " failed")
		t.Fail()
	}
}

func TestCt(t *testing.T) {
	rx := NewRadixGraph(270175, 7)
	Assert(t, "270175", rx.CalcCT() == 270175)
}
