package models

import (
	"github.com/fatih/color"
	"github.com/syndtr/goleveldb/leveldb"
	"strings"
)

var (
	titles *leveldb.DB
)

func NewTitles() {
	var e error
	titles, e = NewLevel("data/titles.level")
	if e != nil {
		panic(e)
	}
}

func searchTitle(okved string, page int) ([]int64, int, error) {
	color.Green("Search %s", okved)
	page--
	bts, e := titles.Get([]byte(okved), nil)
	if e != nil {
		return nil, 0, e
	}
	color.Green("found %d", len(bts)/8)
	res, e := DecodeIntArr(bts)
	total := len(res)
	sliceEnd := (page + 1) * 10
	if total < sliceEnd {
		sliceEnd = total - 1
	}
	res = res[page*10 : sliceEnd]
	return []int64(res), total, e
}

func SearchTitle(okved string, page int) (res []*Org, t int, e error) {
	okved = strings.ToLower(okved)
	ids, total, e := searchTitle(okved, page)
	if e != nil {
		return
	}
	t = total
	res, e = GetMetaOrgs(ids)
	return
}
