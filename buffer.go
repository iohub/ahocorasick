package cedar

type atBuffer struct {
	buf    []matchAt
	offset int
}

const (
	maxInt   = int(^uint(0) >> 1)
	increase = 128
)

var perGrow = 64

func newAtBuffer(n int) *atBuffer {
	if n <= 0 {
		panic("Invalid buffer cap!")
	}

	return &atBuffer{
		buf:    make([]matchAt, n),
		offset: 0,
	}
}

func (b *atBuffer) tryReslice(n int) bool {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return true
	}
	return false
}

func makeSlice(n int) []matchAt {
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	return make([]matchAt, n)
}

func (b *atBuffer) grow(n int) bool {
	if b.tryReslice(n) {
		return true
	}
	l := len(b.buf)
	c := cap(b.buf)
	if c > maxInt-c-n {
		panic(ErrTooLarge)
	}
	buf := makeSlice(c*2 + l)
	copy(buf, b.buf[:b.offset])
	b.buf = buf

	return true
}

func (b *atBuffer) Append(at matchAt) {
	if len(b.buf) <= b.offset {
		b.grow(perGrow)
		perGrow += increase
	}
	b.buf[b.offset].At = at.At
	b.buf[b.offset].OutID = at.OutID
	b.offset++
}
