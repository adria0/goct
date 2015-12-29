package goct

import "testing"
import "bytes"
import "fmt"
import "github.com/robertkrimen/otto"

func Assert(t *testing.T, what string, assertion bool) {
	if !assertion {
		t.Log(what + " failed")
		t.Fail()
	}
}

func TestCodeCreation(t *testing.T) {
	testCodeCreationWithNumberAndRadixAndSeed(t, 912918, 3, 8)
}

func testCodeCreationWithNumberAndRadixAndSeed(t *testing.T, number int, radix int, seed int64) {

	rx := NewRadixGraph(number, radix)
	stmts := rx.CreateCode(seed)

	var b bytes.Buffer

	jsfunc := `
	function digit(id){d=0;w=id.v0;while(w!=id){d++;w=w.v1;}return d}
	function value(base,id){acc=digit(id);pow=base;w=id.v1;
	while(w!=id){acc+=digit(w)*pow;pow*=base;w=w.v1;}
	return acc}
	`
	b.WriteString(jsfunc)

	dumpIndex := func(varExpr VarExpr) string {
		ret := fmt.Sprintf("id%v", varExpr.Node.Id)
		for _, i := range varExpr.Indexes {
			ret = fmt.Sprintf("%v.v%v", ret, i)
		}
		return ret
	}
	for c := range stmts {
		if stmts[c].Assig != nil {
			b.WriteString(fmt.Sprintf("  %v=%v;\n", dumpIndex(stmts[c].Assig.Dst), dumpIndex(stmts[c].Assig.Src)))
		}
		if stmts[c].Newnode != nil {
			b.WriteString(fmt.Sprintf("id%v={};\n", stmts[c].Newnode.Node.Id))
		}
	}
	b.WriteString(fmt.Sprintf("value(%v,id0)", radix))

	fmt.Println(b.String())

	vm := otto.New()
	if ret, err := vm.Run(b.String()); err != nil {
		t.Errorf("for %v/%v %v", number, radix, err)
	} else {
		if intval, _ := ret.ToInteger(); int(intval) != number {
			t.Errorf("Bad response, was %v", intval)
		}
	}
}
