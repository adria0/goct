package main

/*
   Collberg-Thomborson Raddix Watermarking algorithm
*/

import (
	"bytes"
	"fmt"
)

type N struct {
	id    string
	value *N
	next  *N
}

type RadixGraph struct {
	first *N
	radix int
}

func NewRadixGraph(value int, radix int) *RadixGraph {
	valuecopy := value
	nodecount := 0
	for valuecopy > 0 {
		nodecount++
		valuecopy = valuecopy / radix
	}
	if nodecount < radix {
		nodecount = radix
	}

	nodes := make([]N, nodecount)

	for c := 0; c < nodecount; c++ {
		nodes[c].id = fmt.Sprintf("id%v", c)
	}

	valuecopy = value
	nodeno := 0
	for valuecopy > 0 {
		base := valuecopy % radix
		index := (nodeno + (nodecount - base)) % nodecount
		nodes[nodeno].value = &(nodes[index])
		nodes[nodeno].next = &(nodes[(nodeno+1)%nodecount])
		valuecopy = valuecopy / radix
		nodeno++
	}
	for nodeno < nodecount {
		nodes[nodeno].value = &(nodes[nodeno])
		nodes[nodeno].next = &(nodes[(nodeno+1)%nodecount])
		nodeno++
	}
	radixGraph := RadixGraph{&(nodes[0]), radix}
	return &radixGraph
}

func (r *RadixGraph) CalcCT() int {
	digit := func(nb *N) int {
		d := 0
		walker := nb.value
		for walker != nb {
			d++
			walker = walker.next
		}
		return d
	}
	value := digit(r.first)
	pow := r.radix
	walker := r.first.next
	for walker != r.first {
		value = value + pow*digit(walker)
		pow = pow * r.radix
		walker = walker.next
	}
	return value
}

func (r *RadixGraph) CreateDot() []byte {
	var b bytes.Buffer
	b.WriteString("digraph G {")
	for walker := r.first; walker.next != r.first; walker = walker.next {
		b.WriteString(" " + walker.id + " -> " + walker.value.id + " [label=v]")
		b.WriteString(" " + walker.id + " -> " + walker.next.id + " [label=n]")
	}
	b.WriteString("}")
	return b.Bytes()
}
