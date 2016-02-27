package models

import (
	"github.com/zhuharev/flagdb"
)

var (
	ogrns *flagdb.SliceDB
)

func NewSliceDb() {
	var e error
	ogrns, e = flagdb.ReadSliceDb("data/ogrns.flag")
	if e != nil {
		panic(e)
	}
}

func OgrnsRange(page int) ([]int64, int, error) {
	page--
	var (
		countInPage int64 = 10
	)
	arr, e := ogrns.Limit(countInPage, int64(page)*countInPage*8)
	if e != nil {
		return nil, 0, e
	}
	return uint2int(arr), len(arr), nil
}

// todo change name
func OgrnsGoodRange(offset, limit int64) ([]int64, int, error) {
	arr, e := ogrns.Limit(limit, offset)
	if e != nil {
		return nil, 0, e
	}
	return uint2int(arr), len(arr), nil
}

func OgrnsCount() int {
	return ogrns.Len()
}

func uint2int(arr []uint64) (res []int64) {
	for _, v := range arr {
		res = append(res, int64(v))
	}
	return
}
