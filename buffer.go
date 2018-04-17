package cedar

// FixedBuffer fixed reuse buffer for zero alloc
type FixedBuffer struct {
	b   interface{}
	idx int
	cap int
	op  iBufferOP
}

type iBufferOP interface {
	assign(fb *FixedBuffer, val interface{})
	init(fb *FixedBuffer, n int)
}

func (fb *FixedBuffer) push(t interface{}) {
	if fb.idx >= fb.cap {
		panic("ERROR buffer overflow")
	}
	fb.op.assign(fb, t)
	fb.idx++
}

func (fb *FixedBuffer) reset() {
	fb.idx = 0
}

// NewFixedBuffer alloc & init a fixed buffer
func NewFixedBuffer(n int, op iBufferOP) *FixedBuffer {
	fb := &FixedBuffer{
		// b:   make([]interface{}, n),
		idx: 0,
		cap: n,
		op:  op,
	}
	fb.op.init(fb, n)
	return fb
}
