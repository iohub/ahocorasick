package cedar

import (
	"bytes"
	"container/list"
	"io/ioutil"
)

// Matcher Aho Corasick Matcher
type Matcher struct {
	da     *Cedar
	output []*list.List
	fails  []int
}

// MItem matched item in Aho Corasick Matcher
type MItem struct {
	Key   []byte
	Value interface{}
	Freq  int
}

type MPos struct {
	Pos   int
	OutID int
}

func NewMatcher() *Matcher {
	return &Matcher{da: NewCedar()}
}

// Export aho-corasick dfa structures to graphviz file
func (m *Matcher) DumpGraph(fname string) {
	out := &bytes.Buffer{}
	da := m.da
	dumpDFAHeader(out)
	da.dumpTrie(out)
	m.dumpDFAFail(out)
	dumpFinish(out)
	ioutil.WriteFile(fname, out.Bytes(), 0666)
}

// Insert a byte sequence to double array trie inner matcher
func (m *Matcher) Insert(bs []byte, val interface{}) {
	m.da.Insert(bs, val)
}

// Cedar return a cedar trie instance
func (m *Matcher) Cedar() *Cedar {
	return m.da
}

// Compile trie to aho-corasick
func (m *Matcher) Compile() {
	nLen := len(m.da.array)
	// alloc fails, output table space
	if len(m.fails) != 0 || len(m.output) != 0 {
		panic("Matcher already Compiled")
	}
	m.fails = make([]int, nLen)
	for id := 0; id < nLen; id++ {
		m.fails[id] = -1
	}
	m.output = make([]*list.List, nLen)
	fs := 0
	m.fails[fs] = fs
	m.convert()
}

// Matching multiple subsequence in seq return MPos (matched position and output id)
func (m *Matcher) Match(seq []byte) []MPos {
	if len(m.fails) == 0 || len(m.output) == 0 {
		panic("Matcher must be compiled before searching!")
	}
	req := []MPos{}
	nid := 0
	da := m.da
	for i, b := range seq {
		for {
			if da.hasLabel(nid, b) {
				nid, _ = da.child(nid, b)
				if da.isEnd(nid) {
					req = append(req, MPos{Pos: i, OutID: nid})
				}
				break
			}
			if nid == 0 {
				break
			}
			nid = m.fails[nid]
		}
	}
	return req
}

// Export the key, value from MPos
func (m *Matcher) ExportMItem(seq []byte, mp []MPos) []MItem {
	req := []MItem{}
	for _, p := range mp {
		req = append(req, m.mItemOf(seq, p.Pos, p.OutID)...)
	}
	return req
}

func (m *Matcher) mItemOf(seq []byte, offset, id int) []MItem {
	req := []MItem{}
	if m.output[id] == nil {
		return req
	}
	l := m.output[id]
	for e := l.Front(); e != nil; e = e.Next() {
		len := e.Value.(nvalue).len
		val := e.Value.(nvalue).Value
		if len == 0 {
			continue
		}
		bs := seq[offset-len+1 : offset+1]
		req = append(req, MItem{Key: bs, Value: val})
	}
	return req
}

func (m *Matcher) addOutput(nid int, nval nvalue) {
	if m.output[nid] == nil {
		m.output[nid] = &list.List{}
	}
	l := m.output[nid]
	// fmt.Printf("push sublen:%d\n", len)
	l.PushBack(nval)
}

func (m *Matcher) convert() {
	q := &list.List{}
	da, ro := m.da, 0
	m.fails[ro] = ro
	chds := m.da.childs(ro)
	for _, c := range chds {
		m.fails[c.ID] = ro
		q.PushBack(c)
	}
	var fid int
	for q.Len() != 0 {
		e := q.Front()
		q.Remove(e)
		nid := e.Value.(ndesc).ID
		if da.isEnd(nid) {
			vk, _ := da.vKeyOf(nid)
			m.addOutput(nid, da.vals[vk])
		}
		chds := da.childs(nid)
		for _, c := range chds {
			q.PushBack(c)
			for fid = nid; fid != ro; fid = m.fails[fid] {
				fs := m.fails[fid]
				if da.hasLabel(fs, c.Label) {
					fid, _ = da.child(fs, c.Label)
					break
				}
			}
			m.fails[c.ID] = fid
			if da.isEnd(fid) {
				vk, _ := da.vKeyOf(fid)
				m.addOutput(c.ID, da.vals[vk])
			}
		}
	}
}

func (m *Matcher) dumpDFAFail(out *bytes.Buffer) {
	nLen := len(m.da.array)
	for i := 0; i < nLen; i++ {
		fs := m.fails[i]
		if fs != -1 {
			dumpDFALink(out, i, fs, '*', "red")
		}
	}
}
