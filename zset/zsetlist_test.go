package zset

import (
	"fmt"
	"testing"
)


func TestZRandomLevel(t *testing.T) {
	for i := 0; i < 200; i++ {
		t.Logf("%d ", ZRandomLevel())
	}
}

func TestNewZsetList(t *testing.T) {
	z := NewZsetList()
	z.Print()
	z.zInsert("a", 10)
	z.Print()
	z.zInsert("b", 20)
	z.Print()
	z.zInsert("c", 5)
	z.zInsert("d", 75)
	z.zInsert("e", 90)
	z.zInsert("f", 20)
	z.zInsert("g", 27)
	z.zInsert("h", 17)
	z.zInsert("i", 83)
	z.zInsert("j", 69)
	z.zInsert("k", 60)
	z.Print()



	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
}
