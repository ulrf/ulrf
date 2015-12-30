package models

import (
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	okveds *leveldb.DB
)

func NewOkveds() {
	var (
		e error
	)
	okveds, e = NewLevel("data/okveds.level")
	if e != nil {
		panic(e)
	}
}

func searchOkved(okved string, page int) ([]int64, int, error) {
	page--
	bts, e := okveds.Get([]byte(okved), nil)
	if e != nil {
		return nil, 0, e
	}
	res, e := DecodeIntArr(bts)
	total := len(res)
	sliceEnd := (page + 1) * 10
	if total < sliceEnd {
		sliceEnd = total - 1
	}
	res = res[page*10 : sliceEnd]
	return []int64(res), total, e
}

func SearchOkved(okved string, page int) (res []*Org, t int, e error) {
	ids, total, e := searchOkved(okved, page)
	if e != nil {
		return
	}
	t = total
	res, e = GetOrgs(ids)
	return
}
