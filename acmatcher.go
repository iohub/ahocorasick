package cedar

import (
	"bytes"
	"container/list"
	"io/ioutil"
)

const (
	DefaultTokenBufferSize = 4096
	DefaultMatchBufferSize = 4096
)

// Matcher Aho Corasick Matcher
type Matcher struct {
	da       *Cedar
	outputs  []outNode
	fails    []int
	compiled bool
	matched  bool
	buf      *mbuf
}

type mbuf struct {
	token          []MatchToken
	at             []matchAt
	nextIdx, atIdx int
}

// MatchToken matched words in Aho Corasick Matcher
type MatchToken struct {
	KLen  int // len of key
	Value interface{}
	At    int // match position of source text
	Freq  uint
}

type matchAt struct {
	At    int
	OutID int
}

type outNode struct {
	Link *outNode
	vKey int
}

// NewMatcher new an aho corasick matcher
func NewMatcher() *Matcher {
	return &Matcher{
		da:       NewCedar(),
		compiled: false,
		buf: &mbuf{
			token:   make([]MatchToken, DefaultTokenBufferSize),
			at:      make([]matchAt, DefaultMatchBufferSize),
			nextIdx: 0,
			atIdx:   0,
		},
	}
}

func (mb *mbuf) reset() {
	mb.nextIdx, mb.atIdx = 0, 0
}

func (mb *mbuf) grow() {
	idx := mb.atIdx
	mb.at = append(mb.at, make([]matchAt, idx, idx)...)
}

// safely grow
func (mb *mbuf) addAt(mt matchAt) {
	if mb.atIdx >= len(mb.at) {
		mb.grow()
	}
	mb.at[mb.atIdx] = mt
	mb.atIdx++
}

// DumpGraph dumps aho-corasick dfa structures to graphviz file
func (m *Matcher) DumpGraph(fname string) {
	out := &bytes.Buffer{}
	da := m.da
	dumpDFAHeader(out)
	da.dumpTrie(out)
	m.dumpDFAFails(out)
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
	if m.compiled {
		return
	}
	nLen := len(m.da.array)
	m.fails = make([]int, nLen)
	for id := 0; id < nLen; id++ {
		m.fails[id] = -1
	}

	m.outputs = make([]outNode, nLen)
	m.fails[0] = 0
	// build fail function, generate NFA
	m.buildFails()
	// build output function, generate DFA
	m.buildOutputs()
	m.compiled = true
}

// Match multiple subsequence in seq and return tokens
func (m *Matcher) Match(seq []byte) {
	if !m.compiled {
		m.Compile()
	}
	nid := 0
	da := m.da
	m.buf.reset()
	for i, b := range seq {
		for {
			if da.hasLabel(nid, b) {
				nid, _ = da.child(nid, b)
				if da.isEnd(nid) {
					m.buf.addAt(matchAt{OutID: nid, At: i})
				}
				break
			}
			if nid == 0 {
				break
			}
			nid = m.fails[nid]
		}
	}
	m.matched = true
}

func (m *Matcher) HasNext() bool {
	return m.matched && m.buf.nextIdx < m.buf.atIdx
}

func (m *Matcher) NextMatchItem(content []byte) []MatchToken {
	token := []MatchToken{}
	if !m.HasNext() {
		return token
	}
	at := m.buf.at[m.buf.nextIdx]
	for e := &m.outputs[at.OutID]; e != nil; e = e.Link {
		nVal := m.da.vals[e.vKey]
		if nVal.Len == 0 {
			continue
		}
		token = append(token, MatchToken{Value: nVal.Value, At: at.At, KLen: nVal.Len})
	}
	m.buf.nextIdx++
	return token
}

// TokenOf extract matched token in seq
func (m *Matcher) TokenOf(seq []byte, t MatchToken) []byte {
	return seq[t.At-t.KLen+1 : t.At+1]
}

func (m *Matcher) addOutput(nid, fid int) {
	m.outputs[nid].Link = &m.outputs[fid]
}

func (m *Matcher) buildOutputs() {
	da := m.da
	for nid, fid := range m.fails {
		if fid == -1 || !da.isEnd(fid) {
			continue
		}
		da.toEnd(nid)
		m.addOutput(nid, fid)
	}
}

func (m *Matcher) buildFails() {
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
			m.outputs[nid].vKey = vk
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
		}
	}
}

func (m *Matcher) dumpDFAFails(out *bytes.Buffer) {
	nLen := len(m.da.array)
	for i := 0; i < nLen; i++ {
		fs := m.fails[i]
		if fs != -1 {
			dumpDFALink(out, i, fs, '*', "red")
		}
	}
}
