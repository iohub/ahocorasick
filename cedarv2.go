package cedar

type nodev2 struct {
	base  int
	check int
}

type ninfov2 struct {
	sibling uint8
	child   uint8
}

func newNinfov2() ninfov2 {
	return ninfov2{
		sibling: 0,
		child:   0,
	}
}

type blockv2 struct {
	prev   int
	next   int
	num    uint16
	reject uint16
	trial  int
	ehead  int
}

func newBlockv2() blockv2 {
	return blockv2{
		prev:   0,
		next:   0,
		num:    256,
		reject: 257,
		trial:  0,
		ehead:  0,
	}
}

type DaTrie struct {
	array     []nodev2
	ninfo     []ninfov2
	block     []blockv2
	bheadF    int
	bheadC    int
	bheadO    int
	capacity  uint32
	size      uint32
	no_delete int
	reject    []uint16 // 257
}
