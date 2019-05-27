package cedar

import (
	"fmt"
	"os"
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

func TestCedar_Save(t *testing.T) {
	cd := createCedar(t)
	formats := []string{"json", "gob"}
	for _, format := range formats {
		testCedarFormat(t, cd, format)
	}

}

func createCedar(t *testing.T) *Cedar {
	t.Helper()
	cd := NewCedar()
	words := []string{
		"she", "hers", "her", "he",
	}
	for i, word := range words {
		cd.Insert([]byte(word), i)
	}
	return cd
}

func testCedarFormat(t *testing.T, cd *Cedar, format string) {
	t.Helper()
	fileName := "test.tmp"
	f, err := os.Create(fileName)
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(fileName)
	err = cd.Save(f, format)
	if err != nil {
		t.Error(err)
		return
	}

	cd2 := NewCedar()
	err = cd2.LoadFromFile(fileName, format)
	if err != nil {
		t.Error(err)
		return
	}
	keys, nodes, size, cap := cd.Status()
	keys2, nodes2, size2, cap2 := cd2.Status()
	if keys != keys2 {
		t.Errorf("Had %v keys, but reloaded %s has %v", keys, format, keys2)
	}
	if nodes != nodes2 {
		t.Errorf("Had %v nodes, but reloaded %s has %v", nodes, format, nodes2)
	}
	if size != size2 {
		t.Errorf("Had %v size, but reloaded %s has %v", size, format, size2)
	}
	if cap != cap2 {
		t.Errorf("Had %v cap, but reloaded %s has %v", cap, format, cap2)
	}
}
