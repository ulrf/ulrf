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
	arr, t, e := ogrns.Limit(countInPage, int64(page)*countInPage)
	if e != nil {
		return nil, 0, e
	}
	return uint2int(arr), t, nil
}

// todo change name
func OgrnsGoodRange(offset, limit int64) ([]int64, int, error) {
	arr, t, e := ogrns.Limit(limit, offset)
	if e != nil {
		return nil, 0, e
	}
	return uint2int(arr), t, nil
}

func OgrnsCount() int {
	_, t, _ := OgrnsGoodRange(0, 0)
	return t
}

func uint2int(arr []uint64) (res []int64) {
	for _, v := range arr {
		res = append(res, int64(v))
	}
	return
}
