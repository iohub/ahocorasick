package cedar

import (
	"bytes"
	"container/list"
	"io/ioutil"
)

const (
	// TokenBufferSize size of match token buffer
	TokenBufferSize = 4096
	// MatchBufferSize size of match position buffer
	MatchBufferSize = 4096
)

// Matcher Aho Corasick Matcher
type Matcher struct {
	da       *Cedar
	output   []outNode
	fails    []int
	compiled bool
	buf      *mbuf
}

type mbuf struct {
	token  []MatchToken
	at     []matchAt
	ti, ai int
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
			token: make([]MatchToken, MatchBufferSize),
			at:    make([]matchAt, MatchBufferSize),
			ti:    0,
			ai:    0,
		},
	}
}

func (mb *mbuf) reset() {
	mb.ai, mb.ti = 0, 0
}

func (mb *mbuf) addToken(mt MatchToken) {
	if mb.ti >= TokenBufferSize {
		panic("ERROR buffer overflow")
	}
	mb.token[mb.ti] = mt
	mb.ti++
}

func (mb *mbuf) addAt(mt matchAt) {
	if mb.ai >= MatchBufferSize {
		panic("ERROR buffer overflow")
	}
	mb.at[mb.ai] = mt
	mb.ai++
}

// DumpGraph dumps aho-corasick dfa structures to graphviz file
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
	if m.compiled {
		panic(ErrAlreadyCompiled)
	}
	m.fails = make([]int, nLen)
	for id := 0; id < nLen; id++ {
		m.fails[id] = -1
	}

	m.output = make([]outNode, nLen)
	m.fails[0] = 0
	// build fail function, generate NFA
	m.buildFails()
	// build output function, generate DFA
	m.buildOutputs()
	m.compiled = true
}

// Match multiple subsequence in seq and return tokens
func (m *Matcher) Match(seq []byte) []MatchToken {
	if !m.compiled {
		panic(ErrNotCompile)
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
	mb := m.buf.at
	for i := 0; i < m.buf.ai; i++ {
		m.matchOf(seq, mb[i].At, mb[i].OutID)
	}
	tb := m.buf.token
	return tb[:m.buf.ti]
}

// TokenOf extract matched token in seq
func (m *Matcher) TokenOf(seq []byte, t MatchToken) []byte {
	key := seq[t.At-t.KLen+1 : t.At+1]
	return key
}

func (m *Matcher) matchOf(seq []byte, offset, id int) {
	for e := &m.output[id]; e != nil; e = e.Link {
		nval := m.da.vals[e.vKey]
		if nval.len == 0 {
			continue
		}
		m.buf.addToken(MatchToken{Value: nval.Value, At: offset, KLen: nval.len})
	}
}

func (m *Matcher) addOutput(nid, fid int) {
	m.output[nid].Link = &m.output[fid]
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
			m.output[nid].vKey = vk
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

func (m *Matcher) dumpDFAFail(out *bytes.Buffer) {
	nLen := len(m.da.array)
	for i := 0; i < nLen; i++ {
		fs := m.fails[i]
		if fs != -1 {
			dumpDFALink(out, i, fs, '*', "red")
		}
	}
}
