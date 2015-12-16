package main

/*
   Collberg-Thomborson Raddix Watermarking algorithm
   in golang for self-traing & fun
*/

import (
	"bytes"
	"fmt"
	"log"
)

type N struct {
	id string
	v  []*N
}

type RadixGraph struct {
	nodes []N
	radix int
}

const (
	DIGIT = 0
	NEXT  = 1
)

// Creates a new Radix Graph, encoding the selected value
// with the radix base. Mainly this is just a circular linked
// list where each node represents a digit.
// So the node has two pointers: the first is just a link to
// the next node (circular), the second is the value of the
// digit as the distance within the circular linked list from
// the second pointer to the node itself.

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

	rg := RadixGraph{make([]N, nodecount), radix}

	for c := 0; c < nodecount; c++ {
		rg.nodes[c].id = fmt.Sprintf("id%v", c)
	}

	valuecopy = value
	nodeno := 0
	for valuecopy > 0 {
		base := valuecopy % radix
		index := (nodeno + (nodecount - base)) % nodecount
		rg.nodes[nodeno].v = []*N{
			&(rg.nodes[index]),
			&(rg.nodes[(nodeno+1)%nodecount]),
		}
		valuecopy = valuecopy / radix
		nodeno++
	}
	for nodeno < nodecount {
		rg.nodes[nodeno].v = []*N{
			&(rg.nodes[nodeno]),
			&(rg.nodes[(nodeno+1)%nodecount]),
		}
		nodeno++
	}
	return &rg
}

// Compute the value of the radix graph
func (rg *RadixGraph) CalcCT() int {
	digit := func(nb *N) int {
		d := 0
		walker := nb.v[DIGIT]
		for walker != nb {
			d++
			walker = walker.v[NEXT]
		}
		return d
	}
	pow, value := 1, 0
	for n := 0; n < len(rg.nodes); n++ {
		value = value + pow*digit(&(rg.nodes[n]))
		pow = pow * rg.radix
	}
	return value
}

type PendingAssigment struct {
	from    *N
	to      *N
	toIndex int
}

func dump(aliases map[*N]string, pas []PendingAssigment) {
	for _, pa := range pas {
		fmt.Printf("// pending %v[%v] = %v\n", pa.to.id, pa.toIndex, pa.from.id)
	}
	for node, alias := range aliases {
		fmt.Printf("// alias   %v -> %v\n", node.id, alias)
	}
	fmt.Printf("\n")
}

// Create pseudocode to create the Radix Graph preserving
//   the order, defering the creation of inexistent nodes
//   to the moment of existance
func (rg *RadixGraph) CreateCode() {

	pending := make([]PendingAssigment, 0)
	aliases := make(map[*N]string)

	for n := 0; n < len(rg.nodes); n++ {
		fmt.Printf("var id%v N\n", n)
		aliases[&(rg.nodes[n])] = rg.nodes[n].id
		for v := 0; v < len(rg.nodes[n].v); v++ {
			// write assigment alias(n[c]).v[n] =  alias(n[c].v[n])
			added := false
			if from, hasfrom := aliases[rg.nodes[n].v[v]]; hasfrom {
				if to, hasto := aliases[&(rg.nodes[n])]; hasto {
					fmt.Printf("%v.v[%v] = %v\n", to, v, from)
					added = true
				}
			}
			if !added {
				// cannot be added at this time, defer assigment until all info is present
				pending = append(pending, PendingAssigment{
					from:    rg.nodes[n].v[v],
					to:      &(rg.nodes[n]),
					toIndex: v,
				})
				// but assign to itself to not reveal information about
				// node construction. it could be improved to assign to
				// an existing random node
				fmt.Printf("%v.v[%v] = %v\n", rg.nodes[n].id, v, rg.nodes[n].id)
			}
		}

		// check pending assigments and create code for it variables are
		//   available. downward loop for preserving slice when removing element
		for p := len(pending) - 1; p >= 0; p-- {
			if from, hasfrom := aliases[pending[p].from]; hasfrom {
				if to, hasto := aliases[pending[p].to]; hasto {
					fmt.Printf("%v.v[%v] = %v\n", to, pending[p].toIndex, from)
					pending = append(pending[:p], pending[p+1:]...)
				}
			}
		}

		dump(aliases, pending)
	}

	if len(pending) > 0 {
		log.Fatal("Oops! Algorithm failed! Pending assigment list > 0!")
	}
}

// Create graphviz http://www.graphviz.org/ graph
func (rg *RadixGraph) CreateDot() []byte {
	var b bytes.Buffer
	b.WriteString("digraph G {")
	for n := 0; n < len(rg.nodes); n++ {
		for v := 0; v < len(rg.nodes[n].v); v++ {
			line := fmt.Sprintf("%v -> %v [label=%v]\n", rg.nodes[n].id, rg.nodes[n].v[v].id, v)
			b.WriteString(line)
		}
	}
	b.WriteString("}")
	return b.Bytes()
}
