package models

import (
	"github.com/cznic/mathutil"
	"github.com/syndtr/goleveldb/leveldb"
	"strings"
)

var (
	cities *leveldb.DB
)

func NewCities() {
	var (
		e error
	)
	cities, e = NewLevel("data/cities.level")
	if e != nil {
		panic(e)
	}
}

func searchCity(city string, page int) ([]int64, int, error) {
	city = strings.ToUpper(city)
	page--
	bts, e := cities.Get([]byte(city), nil)
	if e != nil {
		return nil, 0, e
	}

	res, e := DecodeIntArr(bts)

	sliceStart := page * 10
	if sliceStart > len(res) {
		sliceStart = mathutil.Max(0, len(res)-10)
	}

	total := len(res)
	sliceEnd := (page + 1) * 10
	if total < sliceEnd {
		sliceEnd = total - 1
	}

	res = res[sliceStart:sliceEnd]
	return []int64(res), total, e
}

func SearchCity(city string, page int) (res []*Org, t int, e error) {
	ids, total, e := searchCity(city, page)
	if e != nil {
		return
	}
	t = total
	res, e = GetOrgs(ids)
	return
}
