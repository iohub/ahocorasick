package cedar

import (
	"fmt"
	"testing"
)

func TestInsertInt(t *testing.T) {
	fmt.Printf("\ntesting int type value...\n")
	cd := NewCedar()
	words := []string{
		"she", "hers", "her", "he",
	}
	for i, word := range words {
		cd.Insert([]byte(word), i)
	}
	ids := cd.PrefixMatch([]byte("hers"), 0)
	for _, id := range ids {
		k, _ := cd.Key(id)
		v, _ := cd.Get(k)
		fmt.Printf("key:%s val:%d\n", string(k), v.(int))
	}
	cd.DumpGraph("datrie.gv")
}

func TestInsertFloat(t *testing.T) {
	fmt.Printf("\ntesting int float32 value...\n")
	cd := NewCedar()
	words := []string{
		"she", "hers", "her", "he",
	}
	for i, word := range words {
		v := float32(i) + 1.880001
		cd.Insert([]byte(word), v)
	}
	ids := cd.PrefixMatch([]byte("hers"), 0)
	for _, id := range ids {
		k, _ := cd.Key(id)
		v, _ := cd.Get(k)
		fmt.Printf("key:%s val:%f\n", string(k), v.(float32))
	}
	cd.DumpGraph("datrie.gv")
}

