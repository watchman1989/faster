package zset

import "testing"

func TestNewZsetList(t *testing.T) {
	z := NewZsetList()
	z.zInsert("a", 10)
	z.Print()
	z.zInsert("b", 120)
	z.Print()
	z.zInsert("c", 1)
	z.Print()
	z.zInsert("d", 75)
	z.zInsert("e", 90)
	z.zInsert("f", 20)
	z.zInsert("g", 27)
	z.zInsert("h", 17)
	z.zInsert("i", 83)
	z.zInsert("j", 69)
	z.zInsert("k", 60)
	z.Print()
}
