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

func (da *DaTrie) findPlace() {

}

func (da *DaTrie) addBlock() int {
	if da.size >= da.capacity {
		grow := da.capacity
		da.capacity += grow
		da.extendArray(int(grow))
		da.extendNinfo(int(grow))
		da.extendBlock((int(da.capacity>>8) - int(da.size>>8)))
	}
	da.block[da.size>>8].ehead = int(da.size)
	da.array[da.size] = nodev2{base: -(int(da.size + 255)), check: -(int(da.size) + 1)}
	for i := da.size + 1; i < da.size+255; i++ {
		da.array[i] = nodev2{base: -int(i - 1), check: -int(i + 1)}
	}
	da.array[da.size+255] = nodev2{base: -(int(da.size) + 254), check: -int(da.size)}
	da.pushBlock(int(da.size)>>8, &da.bheadO, da.bheadO != 0)
	da.size += 256
	return int(da.size>>8) - 1
}

func (da *DaTrie) pushBlock(bi int, headOut *int, empty bool) {
	b := &da.block[bi]
	if empty {
		b.prev, b.next = bi, bi
		*headOut = bi
	} else {
		tailOut := &da.block[*headOut].prev
		b.prev = *tailOut
		b.next = *headOut
		*tailOut = bi
		da.block[*tailOut].next = bi
		*headOut = bi
	}
}

func (da *DaTrie) extendArray(nsize int) {
	for i := 0; i < nsize; i++ {
		da.array = append(da.array, nodev2{0, 0})
	}
}

func (da *DaTrie) extendNinfo(nsize int) {
	for i := 0; i < nsize; i++ {
		da.ninfo = append(da.ninfo, newNinfov2())
	}
}

func (da *DaTrie) extendBlock(nsize int) {
	for i := 0; i < nsize; i++ {
		da.block = append(da.block, newBlockv2())
	}
}
