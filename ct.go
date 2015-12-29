package goct

/*
   Collberg-Thomborson Raddix Watermarking algorithm
   in golang for self-traing & fun
*/

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
)

type N struct {
	Id int
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
		rg.nodes[c].Id = c
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
	src      *N
	dst      *N
	dstIndex int
}

type VarExpr struct {
	Node    *N
	Indexes []int
}

type AssigStmt struct {
	Src VarExpr
	Dst VarExpr
}

type NewNodeStmt struct {
	Node *N
}

type Stmt struct {
	Newnode *NewNodeStmt
	Assig   *AssigStmt
}

// Create pseudocode to create the Radix Graph preserving
//   the order, defering the creation of inexistent nodes
//   to the moment of existance.
func (rg *RadixGraph) CreateCode(seed int64) []Stmt {

	// Shuffle the nodes tho create some random code
	//  we fix the first node because is needed when
	//  rebuilding the graph starting from id0

	rand := rand.New(rand.NewSource(seed))
	permutation := rand.Perm(len(rg.nodes) - 1)
	permutatedNodes := make([]*N, len(rg.nodes))
	permutatedNodes[0] = &(rg.nodes[0])
	for n := 0; n < len(rg.nodes)-1; n++ {
		permutatedNodes[n+1] = &(rg.nodes[permutation[n]+1])
	}

	// Statements = LOCs
	stmts := make([]Stmt, 0)

	// Pending assigments to be resolved later
	pending := make([]PendingAssigment, 0)

	// Aliases are nodes that could be rewritten in terms of id0
	aliases := make(map[int]VarExpr)

	for n := 0; n < len(permutatedNodes); n++ {
		node := permutatedNodes[n]
		stmts = append(stmts, Stmt{Newnode: &NewNodeStmt{node}})

		aliases[node.Id] = VarExpr{node, []int{}}
		for v := 0; v < len(node.v); v++ {

			// Write assigment alias(n[c]).v[n] =  alias(n[c].v[n])
			if src, hassrc := aliases[node.v[v].Id]; hassrc {
				if dst, hasdst := aliases[node.Id]; hasdst {

					newindex := make([]int, len(dst.Indexes), 1+len(dst.Indexes))
					copy(newindex, dst.Indexes)
					newindex = append(newindex, v)

					dst = VarExpr{dst.Node, newindex}
					stmts = append(stmts, Stmt{Assig: &AssigStmt{
						Src: src,
						Dst: dst,
					}})

					continue
				}
			}

			// Cannot be added at this time, defer assigment until all info is present
			pending = append(pending, PendingAssigment{
				src:      node.v[v],
				dst:      node,
				dstIndex: v,
			})

		}

		// check pending assigments and create code for it variables are
		//   available. downward loop for preserving slice when removing element
		for p := len(pending) - 1; p >= 0; p-- {
			if src, hassrc := aliases[pending[p].src.Id]; hassrc {
				if dst, hasdst := aliases[pending[p].dst.Id]; hasdst {

					newindex := make([]int, len(dst.Indexes), 1+len(dst.Indexes))
					copy(newindex, dst.Indexes)
					newindex = append(newindex, pending[p].dstIndex)
					dst = VarExpr{dst.Node, newindex}

					stmts = append(stmts, Stmt{Assig: &AssigStmt{
						Src: src,
						Dst: dst,
					}})

					pending = append(pending[:p], pending[p+1:]...)
					if dst.Node.Id == permutatedNodes[0].Id && len(src.Indexes) == 0 {
						aliases[src.Node.Id] = dst
					}
				}
			}
		}
	}

	if len(pending) > 0 {
		log.Fatal("Oops! Algorithm failed! Pending assigment list > 0!")
	}

	return stmts
}

// Create graphviz http://www.graphviz.org/ graph
func (rg *RadixGraph) CreateDot() []byte {
	var b bytes.Buffer
	b.WriteString("digraph G {")
	for n := 0; n < len(rg.nodes); n++ {
		for v := 0; v < len(rg.nodes[n].v); v++ {
			line := fmt.Sprintf("%v -> %v [label=%v]\n", rg.nodes[n].Id, rg.nodes[n].v[v].Id, v)

			b.WriteString(line)
		}
	}
	b.WriteString("}")
	return b.Bytes()
}
