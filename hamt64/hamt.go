/*
Package hamt64 implements a 64 node wide Hashed Array Mapped Trie. The hash key
is 60 bits wide and broken into 10 numbers of 6 bits each. Those 6bit numbers
allows us to index into a 64 node array. Each node is either a leaf or another
64 node table. So the 60bit hash allows us to index into a B+ Tree with a
branching factor of 64 and a Maximum depth of 10.

The basic insertion operation is to calculate a 60 bit hash value from your key
(a string in the case you use key.StringKey), then split it into ten 6bit
 numbers. These ten numbers represent a path thru the tree. For each level we
use the coresponding number as an index into the 64 cell array. If the cell is
empty we create a  leaf node there. If the cell is occupide by another table
we continue walking up the tree. If the cell is occupide by a leaf we promote
that cell to a new table and put the current leaf and new one into that table
in cells corresponding to that new level. If we are at the maximun depth of
the  tree and there is already a leaf there we insert our key,value pair into
that leaf.

The retrieval operation is a similar tree walk guided by the ten 6bit numbers
till we find a leaf with the key,value pair in it.

The deletion operation is a walk to find the key, then delete the key from the
leaf. An empty leaf is removed from it's table. If the table has only one other
leaf in that level we will remove that leaf, replace the table in it's parent
table placing that last leaf down one level.

*/
package hamt64

import (
	"fmt"
	"log"

	"github.com/lleo/go-hamt-key"
)

// nBits constant is the number of bits(6) a 60bit hash value is split into
// to provied the index of a HAMT.
const nBits uint = key.BitsPerLevel60

// maxDepth constant is the maximum depth(6) of nBits values that constitute
// the path in a HAMT, from [0..maxDepth] for a total of maxDepth+1(10) levels.
// nBits*(maxDepth+1) == HASHBITS (ie 6*(6+1) == 60).
const maxDepth uint = key.MaxDepth60

// tableCapacity constant is the number of table entries in a each node of
// a HAMT datastructure; its value is 2^6 == 64.
const tableCapacity uint = uint(1 << nBits)

// downgradeThreshold constant is the number of nodes a fullTable has shrunk to,
// before it is converted to a compressedTable.
const downgradeThreshold uint = tableCapacity / 8

// upgradeThreshold constan is the number of nodes a compressedTable has grown to,
// before it is converted to a fullTable.
const upgradeThreshold uint = tableCapacity / 2

//func indexMask(depth uint) uint64 {
//	return uint64(uint8(1<<nBits)-1) << (depth * nBits)
//}
//
//func index(h60 uint64, depth uint) uint {
//	var idxMask = indexMask(depth)
//	var idx = uint((h60 & idxMask) >> (depth * nBits))
//	return idx
//}
//
//func hashPathString(hashPath uint64, depth uint) string {
//	if depth == 0 {
//		return "/"
//	}
//	var strs = make([]string, depth)
//
//	for d := uint(0); d < depth; d++ {
//		var idx = index(hashPath, d)
//		strs[d] = fmt.Sprintf("%02d", idx)
//	}
//
//	return "/" + strings.Join(strs, "/")
//}
//
//func hash60String(h60 uint64) string {
//	return hashPathString(h60, maxDepth)
//}
//
//func hashPathMask(depth uint) uint64 {
//	return uint64(1<<(depth*nBits)) - 1
//}
//
//func buildHashPath(hashPath uint64, idx, depth uint) uint64 {
//	return hashPath | uint64(idx<<(depth*nBits))
//}

// Configuration contants to be passed to `hamt64.New(int) *Hamt`.
const (
	// HybridTables indicates the structure should use compressedTable
	// initially, then upgrad to fullTable when appropriate.
	HybridTables = iota //0
	// CompTablesOnly indicates the structure should use compressedTables ONLY.
	// This was intended just save space, but also seems to be faster; CPU cache
	// locality maybe?
	CompTablesOnly //1
	// FullTableOnly indicates the structure should use fullTables ONLY.
	// This was intended to be for speed, as compressed tables use a software
	// bitCount function to access individual cells. Turns out, not so much.
	FullTablesOnly //2
)

// TableOptionName is a map of the table option value Hybrid, CompTablesOnly,
// or FullTableOnly to a string representing that option.
//      var options = hamt64.FullTablesOnly
//      hamt64.TableOptionName[hamt64.FullTablesOnly] == "FullTablesOnly"
var TableOptionName = make(map[int]string, 3)

func init() {
	TableOptionName[HybridTables] = "HybridTables"
	TableOptionName[CompTablesOnly] = "CompTablesOnly"
	TableOptionName[FullTablesOnly] = "FullTablesOnly"
}

//Hamt is a Hashed Array Map Trie data structure. It has a branching factor of
//64 and is at most 10 nodes deep. See:
//https://en.wikipedia.org/wiki/Hash_array_mapped_trie
type Hamt struct {
	root            tableI
	nentries        int
	grade, fullinit bool
}

//New creates a new hamt64.Hamt data structure with the table option set to
//either:
//
//`hamt64.HybridTables`:
//Initially start out with compressedTable, but when the table is half full
//upgrade to fullTable. If a fullTable shrinks to tableCapacity/8(4) entries
//downgrade to compressedTable.
//
//`hamt64.CompTablesOnly`:
//Use compressedTable ONLY with no up/downgrading to/from fullTable. This
//uses the least amount of space.
//
//`hamt64.FullTablesOnly`:
//Only use fullTable no up/downgrading from/to compressedTables. This is
//the fastest performance.
func New(opt int) *Hamt {
	var h = new(Hamt)
	if opt == CompTablesOnly {
		h.grade = false
		h.fullinit = false
	} else if opt == FullTablesOnly {
		h.grade = false
		h.fullinit = true
	} else /* opt == HybridTables */ {
		h.grade = true
		h.fullinit = false
	}
	return h
}

// IsEmpty Hamt method returns a boolean indicating if this Hamt structure has
// no entries.
func (h *Hamt) IsEmpty() bool {
	return h.root == nil
}

func (h *Hamt) Nentries() int {
	return h.nentries
}

func (h Hamt) find(k key.Key) (path pathT, leaf leafI, idx uint) {
	if h.IsEmpty() {
		return nil, nil, 0
	}

	path = newPathT()
	var curTable = h.root

	var h60 = k.Hash60()
	var depth uint
	var curNode nodeI

DepthIter:
	for depth = 0; depth <= maxDepth; depth++ {
		path.push(curTable)
		idx = h60.Index(depth)
		curNode = curTable.get(idx)

		switch n := curNode.(type) {
		case nil:
			leaf = nil
			break DepthIter
		case leafI:
			leaf = n
			break DepthIter
		case tableI:
			if depth == maxDepth {
				log.Panicf("SHOULD NOT BE REACHED; depth,%d == maxDepth,%d & tableI entry found; %s", depth, maxDepth, n)
			}
			curTable = n
			// exit switch then loop for
		default:
			log.Panicf("SHOULD NOT BE REACHED: depth=%d; curNode unknown type=%T;", depth, curNode)
		}
	}

	return
}

// Get Hamt method looks up a given key in the Hamt data structure.
// BenchHamt64:
//func (h *Hamt) Get(k key.Key) (val interface{}, found bool) {
//	var _, leaf, _ = h.find(k)
//
//	if leaf == nil {
//		return //nil, false
//	}
//
//	val, found = leaf.get(k)
//	return
//}

func (h *Hamt) Get(k key.Key) (val interface{}, found bool) {
	if h.IsEmpty() {
		return //nil, false
	}

	var h60 = k.Hash60()

	var curTable = h.root //ISA tableI

	for depth := uint(0); depth <= maxDepth; depth++ {
		var idx = h60.Index(depth)
		var curNode = curTable.get(idx) //nodeI

		if curNode == nil {
			return //nil, false
		}

		if leaf, isLeaf := curNode.(leafI); isLeaf {
			val, found = leaf.get(k)
			return //val, found
		}

		if depth == maxDepth {
			panic("SHOULD NOT HAPPEN")
		}
		curTable = curNode.(tableI)
	}

	panic("SHOULD NEVER BE REACHED")
}

// Put Hamt method inserts a given key/val pair into the Hamt data structure.
// It returns a boolean indicating if the key/val was inserted or whether or
// not the key already existed and the val was merely overwritten.
func (h *Hamt) Put(k key.Key, v interface{}) bool {
	var depth uint
	var hashPath key.HashVal60

	if h.IsEmpty() {
		h.root = h.newRootTable(depth, hashPath, newFlatLeaf(k, v))
		h.nentries++
		return true
	}

	var path = newPathT()
	var curTable = h.root

	for depth = 0; depth <= maxDepth; depth++ {
		var idx = k.Hash60().Index(depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			curTable.set(idx, newFlatLeaf(k, v))
			h.nentries++

			// upgrade?
			if h.grade {
				_, isCompressedTable := curTable.(*compressedTable)
				if isCompressedTable && curTable.nentries() >= upgradeThreshold {
					curTable = upgradeToFullTable(hashPath, curTable.entries())
					if depth == 0 {
						h.root = curTable
					} else {
						parentTable := path.peek()
						parentIdx := k.Hash60().Index(depth - 1)
						parentTable.set(parentIdx, curTable)
					}
				}
			}

			return true //inserted
		}

		if curLeaf, isLeaf := curNode.(leafI); isLeaf {
			if curLeaf.Hash60() == k.Hash60() {
				// This is a minor optimization but since these two leaves
				// will collide all the way up the to maxDepth, we can
				// choose to create the collisionLeaf hear and now.

				// Accumulate collisionLeaf
				colLeaf, inserted := curLeaf.put(k, v)
				if inserted {
					curTable.set(idx, colLeaf)
					h.nentries++
				}
				return inserted
			}

			if depth == maxDepth {
				// this test should be delete cuz it is logically impossible
				if curLeaf.Hash60() != k.Hash60() {
					// This should not happen cuz we had to walk up maxDepth
					// levels to get here.
					panic("WTF!!!")
				}

				// Accumulate collisionLeaf
				insLeaf, inserted := curLeaf.put(k, v)
				if inserted {
					curTable.set(idx, insLeaf)
					h.nentries++
				}
				return inserted
			}

			hashPath = hashPath.BuildHashPath(idx, depth)
			var collisionTable = h.newTable(depth+1, hashPath, curLeaf, newFlatLeaf(k, v))
			curTable.set(idx, collisionTable)
			h.nentries++

			return true
		}

		hashPath = hashPath.BuildHashPath(idx, depth)
		path.push(curTable)
		curTable = curNode.(tableI)
	}

	//log.Println(path)
	//log.Printf("k=%s, v=%v", k, v)

	panic("WTF!!")
}

// Del Hamt Method removes a given key from the Hamt data structure. It returns
// a pair of values: the value stored and a boolean indicating if the key was
// even found and deleted.
func (h *Hamt) Del(k key.Key) (interface{}, bool) {
	if h.IsEmpty() {
		return nil, false
	}

	var h60 = k.Hash60()
	var depth uint
	var hashPath key.HashVal60

	var path = newPathT()
	var curTable = h.root

	for depth = 0; depth <= maxDepth; depth++ {
		var idx = h60.Index(depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			return nil, false
		}

		if curLeaf, isLeaf := curNode.(leafI); isLeaf {
			val, delLeaf, deleted := curLeaf.del(k)
			if !deleted {
				return nil, false
			}
			// else a leaf key/value was deleted
			h.nentries--

			// If curLeaf was a collisionLeaf,
			// then delLeaf is either a slimmed down collisionLeaf or a flatLeaf.
			// If curLeaf was a flatLeaf then delLeaf is nil.
			curTable.set(idx, delLeaf)

			// downgrade?
			if h.grade {
				if delLeaf == nil {
					_, isFullTable := curTable.(*fullTable)
					if isFullTable && curTable.nentries() <= downgradeThreshold {
						curTable = downgradeToCompressedTable(hashPath, curTable.entries())
						if depth == 0 {
							h.root = curTable
						} else {
							parentTable := path.peek()
							parentIdx := h60.Index(depth - 1)
							parentTable.set(parentIdx, curTable)
						}
					}
				}
			}

			// If curTable has only one entry and that entry is a leafI,
			// then collapse that leafI down to the position curTable holds
			// in the parent Table; repeat test and collapse for parent table.
			//
			// These are identical for conditionals:
			//  curTable != h.root AND len(path) > 0 AND depth > 0
			//
			for curTable.nentries() == 1 && depth > 0 {
				// _ = ASSERT && Assert(curTable != h.root, "curTable == h.root")
				// _ = ASSERT && Assert(depth == len(path), "depth != len(path)")

				var node = curTable.entries()[0].node
				var leaf, isLeaf = node.(leafI)
				if !isLeaf {
					// We only collapse leafs
					break
				}

				// Collapse leaf down to where curTable is in parentTable

				var parentTable = path.pop()
				depth-- // OR depth = len(path)

				parentIdx := curTable.Hash60().Index(depth)
				parentTable.set(parentIdx, leaf)

				curTable = parentTable
			}

			// TODO: I should keep this table rather than throwing it away.
			// Instead using h.root == nil to detect emptyness, we should
			// trust our accounting and use h.nentries == 0.
			if curTable == h.root && curTable.nentries() == 0 {
				h.root = nil
			}

			return val, true
		} //if isLeaf

		// curNode is not nil
		// curNode is not a leafI
		// curNode MUST be a tableI
		hashPath = hashPath.BuildHashPath(idx, depth)
		path.push(curTable)
		curTable = curNode.(tableI)
	} //for depth loop

	//log.Printf("WTF! this should never be called; k=%s", k)
	return nil, false
}

// String returns a string representation of the Hamt string.
func (h *Hamt) String() string {
	return fmt.Sprintf("Hamt{ nentries: %d, root: %s }", h.nentries, h.root.LongString("", 0))
}

// LongString returns a complete listing of the entire Hamt data structure.
func (h *Hamt) LongString(indent string) string {
	var str string
	if h.root != nil {
		str = indent + fmt.Sprintf("Hamt{ nentries: %d, root:\n", h.nentries)
		str += indent + h.root.LongString(indent, 0)
		str += indent + "} //Hamt"
	} else {
		str = indent + fmt.Sprintf("Hamt{ nentries: %d, root: nil }", h.nentries)
	}
	return str
}
