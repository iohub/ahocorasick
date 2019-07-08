package cedar

import (
	"reflect"
	"sort"
	"testing"
)

func TestMatchVals(t *testing.T) {
	word := ""
	m := NewMatcher()
	for i := 0; i < 5; i++ {
		word += "a"
		m.Insert([]byte(word), i)
	}
	m.Compile()

	ret := m.MatchVals([]byte(word + word))
	ids := []int{}
	for _, v := range ret {
		ids = append(ids, v.(int))
	}
	sort.Ints(ids)
	expect := []int{0, 1, 2, 3, 4}
	if !reflect.DeepEqual(ids, expect) {
		t.Fatalf("ids=%+v, expected %+v", ids, expect)
	}
}
