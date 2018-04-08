package cedar

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
)

// defines max & min value of chinese CJK code
const (
	valueLimit = int(^uint(0) >> 1)
	CJKZhMin   = '\u4E00'
	CJKZhMax   = '\u9FFF'
	acsiiMax   = '\u007F'
	asciiA     = 'A'
	asciiz     = 'z'
)

type node struct {
	Value int
	Check int
}

type nvalue struct {
	len   int
	Value interface{}
}

type ndesc struct {
	Label byte
	ID    int
}

func (n *node) base() int { return -(n.Value + 1) }

type ninfo struct {
	Sibling byte
	Child   byte
	End     bool
}

type block struct {
	Prev, Next, Num, reject, Trial, Ehead int
}

func (b *block) init() {
	b.Num = 256
	b.reject = 257
}

// Cedar encapsulates a fast and compressed double array trie for words query
type Cedar struct {
	array    []node
	infos    []ninfo
	blocks   []block
	vals     map[int]nvalue
	vkey     int
	reject   [257]int
	bheadF   int
	bheadC   int
	bheadO   int
	capacity int
	size     int
	ordered  bool
	maxTrial int
}

// NewCedar new a Cedar instance
func NewCedar() *Cedar {
	da := Cedar{
		array:    make([]node, 256),
		infos:    make([]ninfo, 256),
		blocks:   make([]block, 1),
		capacity: 256,
		size:     256,
		vals:     make(map[int]nvalue),
		vkey:     1,
		ordered:  true,
		maxTrial: 1,
	}

	da.array[0] = node{-2, 0}
	for i := 1; i < 256; i++ {
		da.array[i] = node{-(i - 1), -(i + 1)}
	}
	da.array[1].Value = -255
	da.array[255].Check = -1

	da.blocks[0].Ehead = 1
	da.blocks[0].init()

	for i := 0; i <= 256; i++ {
		da.reject[i] = i + 1
	}

	return &da
}

func (da *Cedar) vKey() int {
	k := da.vkey
	for {
		k = (k + 1) % da.capacity
		if _, ok := da.vals[k]; !ok {
			break
		}
	}
	da.vkey = k
	return k
}

func (da *Cedar) keyLen(id int) int {
	val, err := da.vKeyOf(id)
	if err != nil {
		return 0
	}
	if v, ok := da.vals[val]; ok {
		return v.len
	}
	return 0
}

// Get value by key, insert the key if not exist
func (da *Cedar) get(key []byte, from, pos int) int {
	for ; pos < len(key); pos++ {
		if value := da.array[from].Value; value >= 0 && value != valueLimit {
			to := da.follow(from, 0)
			da.array[to].Value = value
		}
		from = da.follow(from, key[pos])
	}
	to := from
	if da.array[from].Value < 0 {
		to = da.follow(from, 0)
	}
	return to
}

func (da *Cedar) follow(from int, label byte) int {
	base := da.array[from].base()
	to := base ^ int(label)
	if base < 0 || da.array[to].Check < 0 {
		hasChild := false
		if base >= 0 {
			hasChild = (da.array[base^int(da.infos[from].Child)].Check == from)
		}
		to = da.popEnode(base, label, from)
		da.pushSibling(from, to^int(label), label, hasChild)
	} else if da.array[to].Check != from {
		to = da.resolve(from, base, label)
	} else if da.array[to].Check == from {
	} else {
		panic("Cedar: internal error, should not be here")
	}
	return to
}

func (da *Cedar) popBlock(bi int, headIn *int, last bool) {
	if last {
		*headIn = 0
	} else {
		b := &da.blocks[bi]
		da.blocks[b.Prev].Next = b.Next
		da.blocks[b.Next].Prev = b.Prev
		if bi == *headIn {
			*headIn = b.Next
		}
	}
}

func (da *Cedar) pushBlock(bi int, headOut *int, empty bool) {
	b := &da.blocks[bi]
	if empty {
		*headOut, b.Prev, b.Next = bi, bi, bi
	} else {
		tailOut := &da.blocks[*headOut].Prev
		b.Prev = *tailOut
		b.Next = *headOut
		*headOut, *tailOut, da.blocks[*tailOut].Next = bi, bi, bi
	}
}

func (da *Cedar) addBlock() int {
	if da.size == da.capacity {
		da.capacity *= 2

		oldarray := da.array
		da.array = make([]node, da.capacity)
		copy(da.array, oldarray)

		oldNinfo := da.infos
		da.infos = make([]ninfo, da.capacity)
		copy(da.infos, oldNinfo)

		oldBlock := da.blocks
		da.blocks = make([]block, da.capacity>>8)
		copy(da.blocks, oldBlock)
	}

	da.blocks[da.size>>8].init()
	da.blocks[da.size>>8].Ehead = da.size

	da.array[da.size] = node{-(da.size + 255), -(da.size + 1)}
	for i := da.size + 1; i < da.size+255; i++ {
		da.array[i] = node{-(i - 1), -(i + 1)}
	}
	da.array[da.size+255] = node{-(da.size + 254), -da.size}

	da.pushBlock(da.size>>8, &da.bheadO, da.bheadO == 0)
	da.size += 256
	return da.size>>8 - 1
}

func (da *Cedar) transferBlock(bi int, headIn, headOut *int) {
	da.popBlock(bi, headIn, bi == da.blocks[bi].Next)
	da.pushBlock(bi, headOut, *headOut == 0 && da.blocks[bi].Num != 0)
}

func (da *Cedar) popEnode(base int, label byte, from int) int {
	e := base ^ int(label)
	if base < 0 {
		e = da.findPlace()
	}
	bi := e >> 8
	n := &da.array[e]
	b := &da.blocks[bi]
	b.Num--
	if b.Num == 0 {
		if bi != 0 {
			da.transferBlock(bi, &da.bheadC, &da.bheadF)
		}
	} else {
		da.array[-n.Value].Check = n.Check
		da.array[-n.Check].Value = n.Value
		if e == b.Ehead {
			b.Ehead = -n.Check
		}
		if bi != 0 && b.Num == 1 && b.Trial != da.maxTrial {
			da.transferBlock(bi, &da.bheadO, &da.bheadC)
		}
	}
	n.Value = valueLimit
	n.Check = from
	if base < 0 {
		da.array[from].Value = -(e ^ int(label)) - 1
	}
	return e
}

func (da *Cedar) pushEnode(e int) {
	bi := e >> 8
	b := &da.blocks[bi]
	b.Num++
	if b.Num == 1 {
		b.Ehead = e
		da.array[e] = node{-e, -e}
		if bi != 0 {
			da.transferBlock(bi, &da.bheadF, &da.bheadC)
		}
	} else {
		prev := b.Ehead
		next := -da.array[prev].Check
		da.array[e] = node{-prev, -next}
		da.array[prev].Check = -e
		da.array[next].Value = -e
		if b.Num == 2 || b.Trial == da.maxTrial {
			if bi != 0 {
				da.transferBlock(bi, &da.bheadC, &da.bheadO)
			}
		}
		b.Trial = 0
	}
	if b.reject < da.reject[b.Num] {
		b.reject = da.reject[b.Num]
	}
	da.infos[e] = ninfo{}
}

// hasChild: wherether the `from` node has children
func (da *Cedar) pushSibling(from, base int, label byte, hasChild bool) {
	c := &da.infos[from].Child
	keepOrder := *c == 0
	if da.ordered {
		keepOrder = label > *c
	}
	if hasChild && keepOrder {
		c = &da.infos[base^int(*c)].Sibling
		for da.ordered && *c != 0 && *c < label {
			c = &da.infos[base^int(*c)].Sibling
		}
	}
	da.infos[base^int(label)].Sibling = *c
	*c = label
}

func (da *Cedar) popSibling(from, base int, label byte) {
	c := &da.infos[from].Child
	for *c != label {
		c = &da.infos[base^int(*c)].Sibling
	}
	*c = da.infos[base^int(*c)].Sibling
}

func (da *Cedar) consult(baseN, baseP int, cN, cP byte) bool {
	cN = da.infos[baseN^int(cN)].Sibling
	cP = da.infos[baseP^int(cP)].Sibling
	for cN != 0 && cP != 0 {
		cN = da.infos[baseN^int(cN)].Sibling
		cP = da.infos[baseP^int(cP)].Sibling
	}
	return cP != 0
}

func (da *Cedar) hasLabel(id int, label byte) bool {
	_, err := da.child(id, label)
	return err == nil
}

func (da *Cedar) child(id int, label byte) (int, error) {
	base := da.array[id].base()
	cid := base ^ int(label)
	if cid < 0 || cid >= da.size || da.array[cid].Check != id {
		return -1, errors.New("cann't find child in node")
	}
	return cid, nil
}

func (da *Cedar) childs(id int) []ndesc {
	req := []ndesc{}
	base := da.array[id].base()
	s := da.infos[id].Child
	if s == 0 && base > 0 {
		s = da.infos[base].Sibling
	}
	for s != 0 {
		to := base ^ int(s)
		if to < 0 {
			break
		}
		req = append(req, ndesc{ID: to, Label: s})
		s = da.infos[to].Sibling
	}
	return req
}

func (da *Cedar) setChild(base int, c byte, label byte, flag bool) []byte {
	child := make([]byte, 0, 257)
	if c == 0 {
		child = append(child, c)
		c = da.infos[base^int(c)].Sibling
	}
	if da.ordered {
		for c != 0 && c <= label {
			child = append(child, c)
			c = da.infos[base^int(c)].Sibling
		}
	}
	if flag {
		child = append(child, label)
	}
	for c != 0 {
		child = append(child, c)
		c = da.infos[base^int(c)].Sibling
	}
	return child
}

func (da *Cedar) findPlace() int {
	if da.bheadC != 0 {
		return da.blocks[da.bheadC].Ehead
	}
	if da.bheadO != 0 {
		return da.blocks[da.bheadO].Ehead
	}
	return da.addBlock() << 8
}

func (da *Cedar) findPlaces(child []byte) int {
	bi := da.bheadO
	if bi != 0 {
		bz := da.blocks[da.bheadO].Prev
		nc := len(child)
		for {
			b := &da.blocks[bi]
			if b.Num >= nc && nc < b.reject {
				for e := b.Ehead; ; {
					base := e ^ int(child[0])
					for i := 0; da.array[base^int(child[i])].Check < 0; i++ {
						if i == len(child)-1 {
							b.Ehead = e
							return e
						}
					}
					e = -da.array[e].Check
					if e == b.Ehead {
						break
					}
				}
			}
			b.reject = nc
			if b.reject < da.reject[b.Num] {
				da.reject[b.Num] = b.reject
			}
			bin := b.Next
			b.Trial++
			if b.Trial == da.maxTrial {
				da.transferBlock(bi, &da.bheadO, &da.bheadC)
			}
			if bi == bz {
				break
			}
			bi = bin
		}
	}
	return da.addBlock() << 8
}

func (da *Cedar) resolve(fromN, baseN int, labelN byte) int {
	toPN := baseN ^ int(labelN)
	fromP := da.array[toPN].Check
	baseP := da.array[fromP].base()

	flag := da.consult(baseN, baseP, da.infos[fromN].Child, da.infos[fromP].Child)
	var children []byte
	if flag {
		children = da.setChild(baseN, da.infos[fromN].Child, labelN, true)
	} else {
		children = da.setChild(baseP, da.infos[fromP].Child, 255, false)
	}
	var base int
	if len(children) == 1 {
		base = da.findPlace()
	} else {
		base = da.findPlaces(children)
	}
	base ^= int(children[0])
	var from int
	var nbase int
	if flag {
		from = fromN
		nbase = baseN
	} else {
		from = fromP
		nbase = baseP
	}
	if flag && children[0] == labelN {
		da.infos[from].Child = labelN
	}
	da.array[from].Value = -base - 1
	for i := 0; i < len(children); i++ {
		to := da.popEnode(base, children[i], from)
		newto := nbase ^ int(children[i])
		if i == len(children)-1 {
			da.infos[to].Sibling = 0
		} else {
			da.infos[to].Sibling = children[i+1]
		}
		if flag && newto == toPN { // new node has no child
			continue
		}
		n := &da.array[to]
		nn := &da.array[newto]
		n.Value = nn.Value
		if n.Value < 0 && children[i] != 0 {
			// this node has children, fix their check
			c := da.infos[newto].Child
			da.infos[to].Child = c
			da.array[n.base()^int(c)].Check = to
			c = da.infos[n.base()^int(c)].Sibling
			for c != 0 {
				da.array[n.base()^int(c)].Check = to
				c = da.infos[n.base()^int(c)].Sibling
			}
		}
		if !flag && newto == fromN { // parent node moved
			fromN = to
		}
		if !flag && newto == toPN {
			da.pushSibling(fromN, toPN^int(labelN), labelN, true)
			da.infos[newto].Child = 0
			nn.Value = valueLimit
			nn.Check = fromN
		} else {
			da.pushEnode(newto)
		}
	}
	if flag {
		return base ^ int(labelN)
	}
	return toPN
}

func valueOfRune(r rune) uint16 {
	v := uint32(r)
	if v >= CJKZhMin {
		return uint16(v - CJKZhMin + asciiz + 1)
	}
	return uint16(v)
}

func runeOfValue(v uint16) rune {
	if v >= asciiz {
		return rune(v - 1 - asciiz + CJKZhMin)
	}
	return rune(v)
}

func (da *Cedar) isEnd(id int) bool {
	if da.infos[id].End {
		return true
	}
	return da.infos[id].Child == 0
}

func (da *Cedar) toEnd(id int) {
	da.infos[id].End = true
}

func dumpDFAHeader(out *bytes.Buffer) {
	out.WriteString("digraph DFA {\n")
	out.WriteString("\tnode [color=lightblue2 style=filled]\n")
}

func dumpFinish(out *bytes.Buffer) {
	out.WriteString("}\n")
}

func dumpDFALink(out *bytes.Buffer, fid int, tid int, val uint16, color string) {
	r := runeOfValue(val)
	out.WriteString(fmt.Sprintf("\t\"node(%d)\" -> \"node(%d)\" [label=\"(%c)\" color=%s]\n", fid, tid, r, color))
}

func (da *Cedar) dumpTrie(out *bytes.Buffer) {
	end := "END"
	for id := 0; id < da.size; id++ {
		pid := da.array[id].Check
		if pid < 0 {
			continue
		}
		pbase := da.array[pid].base()
		label := pbase ^ id
		if label == 0 {
			label = '0'
		}
		dumpDFALink(out, pid, id, uint16(label), "black")
		if da.isEnd(id) {
			out.WriteString(fmt.Sprintf("\t\"%s\" -> \"END(%d)\"\n", end, id))
			end = fmt.Sprintf("END(%d)", id)
		}
	}
}

// DumpGraph dumps inner data structures for graphviz
func (da *Cedar) DumpGraph(fname string) {
	out := &bytes.Buffer{}
	dumpDFAHeader(out)
	da.dumpTrie(out)
	dumpFinish(out)
	ioutil.WriteFile(fname, out.Bytes(), 0666)
}
