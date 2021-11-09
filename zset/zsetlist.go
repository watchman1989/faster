package zset

import (
	"fmt"
	"math/rand"
	"strings"
)

const (
	defaultMaxLevel = 2
	defaultP = 4
)

type ZsetLevel struct {
	Forward *ZsetNode
	Span int64
}


type ZsetNode struct {
	Score int64
	Member string
	Backward *ZsetNode
	Levels []*ZsetLevel
}

type ZsetList struct {
	Header *ZsetNode
	Tail *ZsetNode
	Length int64
	Level int
}

//
func newZsetListNode(level int, member string, score int64) *ZsetNode {
	zNode := &ZsetNode{
		Score: score,
		Member: member,
		Levels: make([]*ZsetLevel, 0),
	}

	for i := 0; i < level; i++ {
		zNode.Levels = append(zNode.Levels, &ZsetLevel{})
	}
	return zNode
}

//
func NewZsetList() *ZsetList {
	z := &ZsetList{
		Level: 1,
		Length: 0,
		Header: newZsetListNode(defaultMaxLevel, "header",  0),
	}
	for i := 0; i < defaultMaxLevel; i++ {
		z.Header.Levels[i].Forward = nil
		z.Header.Levels[i].Span = 0
	}
	z.Header.Backward = nil
	z.Tail = nil
	return z
}

//get random level
func ZRandomLevel() int {
	level := 1
	for {
		if rand.Int() & 0xFFFF < 0xFFFF / defaultP {
			level++
		} else {
			break
		}
	}
	if level < defaultMaxLevel {
		return level
	} else {
		return defaultMaxLevel
	}
}


func (z *ZsetList) getRankUpdateList(score int64, member string) (*ZsetNode, [defaultMaxLevel]int64, [defaultMaxLevel]*ZsetNode) {
	var rankList [defaultMaxLevel] int64
	var updateList [defaultMaxLevel] *ZsetNode
	var zn *ZsetNode
	//get rank and updates
	zn = z.Header
	for i := z.Level - 1; i >= 0; i-- {
		if i == z.Level -1 {
			rankList[i] = 0
		} else {
			rankList[i] = rankList[i + 1]
		}
		for ; zn.Levels[i].Forward != nil && (zn.Levels[i].Forward.Score < score || (zn.Levels[i].Forward.Score == score && strings.Compare(zn.Levels[i].Forward.Member, member) < 0)); {
				rankList[i] += zn.Levels[i].Span
				zn = zn.Levels[i].Forward
		}
		updateList[i] = zn
	}
	return zn, rankList, updateList
}


func (z *ZsetList) zInsert(member string, score int64) {
	fmt.Printf("insert %s %d\n", member, score)
	_, rankList, updateList := z.getRankUpdateList(score, member)
	/*
	for k, v := range updateList {
		fmt.Printf("updateList %d:\n", k)
		printNode(v)
	}
	fmt.Println()
	*/
	//init
	level := ZRandomLevel()
	//fmt.Printf("random level is %d\n", level)
	if level > z.Level {
		for i := z.Level; i < level; i++ {
			rankList[i] = 0
			updateList[i] = z.Header
			updateList[i].Levels[i].Span = z.Length
		}
		z.Level = level
	}
	//insert node
	zn := newZsetListNode(level, member, score)
	for i := 0; i < level; i++ {
		zn.Levels[i].Forward = updateList[i].Levels[i].Forward
		updateList[i].Levels[i].Forward = zn

		zn.Levels[i].Span = updateList[i].Levels[i].Span - (rankList[0] - rankList[i])
		updateList[i].Levels[i].Span = (rankList[0] - rankList[i]) + 1
	}
	//fmt.Printf("z.Header: %v, z.Tail: %v, z.Level: %d, z.Length: %d\n", z.Header, z.Tail, z.Level, z.Length)
	//update span
	for i := level; i < z.Level; i++ {
		updateList[i].Levels[i].Span++
	}
	//modify header and tail
	if updateList[0] == z.Header {
		zn.Backward = nil
	} else {
		zn.Backward = updateList[0]
	}
	if zn.Levels[0].Forward != nil {
		zn.Levels[0].Forward.Backward = zn
	} else {
		z.Tail = zn
	}
	//length +1
	z.Length++
}

/*
func (z *ZsetList) zDeleteNode(zn *ZsetNode, updateList [defaultMaxLevel]*ZsetNode) {
	for i := 0; i < z.Level; i++ {
		if updateList[i].Levels[i].Forward == zn {
			updateList[i].Levels[i].Span += zn.Levels[i].Span - 1
			updateList[i].Levels[i].Forward = zn.Levels[i].Forward
		} else {
			updateList[i].Levels[i].Span -= 1
		}
	}
	if zn.Levels[0].Forward != nil {
		zn.Levels[0].Forward.Backward = zn.Backward
	} else {
		z.Tail = zn.Backward
	}
	for ; z.Level > 1 && z.Header.Levels[z.Level - 1].Forward == nil; {
		z.Level--
	}
	z.Length--
}


func (z *ZsetList) zDelete(score int64, member string) {
	zn, _, updateList := z.getRankUpdateList(score, member)
	zn = zn.Levels[0].Forward
	if zn != nil && zn.Score == score && strings.Compare(zn.Member, member) == 0 {
		z.zDeleteNode(zn, updateList)
	}
	return
}

func (z *ZsetList) zUpdateScore(member string, oldScore int64, newScore int64) {
	zn, _, updateList := z.getRankUpdateList(oldScore, member)
	zn = zn.Levels[0].Forward
	if zn == nil || zn.Score != oldScore || strings.Compare(zn.Member, member) != 0 {
		return
	}
	if (zn.Backward == nil || zn.Backward.Score < newScore) && (zn.Levels[0].Forward == nil || zn.Levels[0].Forward.Score > newScore) {
		zn.Score = newScore
		return
	}
	z.zDeleteNode(zn, updateList)
	z.zInsert(member, newScore)
}



func (z *ZsetList) zGetRank(score int64, member string) int64 {
	rank := int64(0)
	zn := z.Header
	for i := z.Level - 1; i >= 0; i-- {
		for ; zn.Levels[i].Forward != nil && zn.Levels[i].Forward.Score < score || (zn.Levels[i].Forward.Score == score && strings.Compare(zn.Levels[i].Forward.Member, member) <= 0); {
			zn = zn.Levels[i].Forward
			rank += zn.Levels[i].Span
		}
		if strings.Compare(zn.Member, member) == 0 {
			return rank
		}
	}
	return rank
}

func (z *ZsetList) zGetMemberByRank(rank int64) *ZsetNode {
	zn := z.Header
	s := int64(0)
	for i := z.Level - 1; i >= 0; i-- {
		for ; zn.Levels[i].Forward != nil && (s + zn.Levels[i].Span) <= rank; {
			s += zn.Levels[i].Span
			zn = zn.Levels[i].Forward
		}
		if s == rank {
			return zn
		}
	}
	return nil
}
*/


func printNode(node *ZsetNode) {
	if node == nil {
		return
	}
	fmt.Printf("****Node: %s, score: %d\n", node.Member, node.Score)
	if node.Backward == nil {
		fmt.Printf("Backward: nil\n")
	} else {
		fmt.Printf("Backward: (member: %s, score: %d)\n", node.Backward.Member, node.Backward.Score)
	}
	for k, v := range node.Levels {
		if v == nil {
			fmt.Printf("level %d: nil\n", k)
		} else {
			if v.Forward != nil {
				fmt.Printf("level %d: [Forward: (member: %s, score: %d)]\n", k, v.Forward.Member, v.Forward.Score)
			}
		}
	}
	fmt.Printf("******************************\n")
	return
}


func (z *ZsetList) PrintMeta() {
	fmt.Printf("####meta level = %d, length = %d\n", z.Level, z.Length)
	fmt.Printf("header->%s\n", z.Header.Member)
	if z.Tail != nil {
		fmt.Printf("tail->%s\n", z.Tail.Member)
	}
	fmt.Printf("###############################\n")
}

func (z *ZsetList) Print() {
	z.PrintMeta()
	for i := z.Level - 1; i >= 0; i-- {
		fmt.Printf("+++++++++++++level %d++++++++++++++++\n", i)
		zn := z.Header
		for ; zn != nil; {
			//printNode(zn)
			fmt.Printf("%s:%d  ", zn.Member, zn.Score)
			zn = zn.Levels[i].Forward
		}
	}
}

