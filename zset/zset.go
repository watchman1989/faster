package zset


type Zset struct {
	zList *ZsetList
	zMap map[string]int64
}

