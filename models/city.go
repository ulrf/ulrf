package models

import (
	"fmt"
	"github.com/cznic/mathutil"
	"github.com/fatih/color"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/zhuharev/raddress"
	"strings"
	"sync"
	"time"
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

	c, e := getAllCities()
	if e != nil {
		panic(e)
	}

	if len(c) == 0 {
		panic("nil")
	}

	raddress.AllowCities(c)
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

func getAllCities() (res []string, e error) {
	iter := cities.NewIterator(nil, nil)

	for iter.Next() {
		res = append(res, string(iter.Key()))
	}
	iter.Release()

	return
}

func SearchCity(city string, page int) (res []*Org, t int, e error) {
	start := time.Now()
	ids, total, e := searchCity(city, page)
	if e != nil {
		return
	}
	t = total
	printNeedSince(start, time.Millisecond*100, "Searched")
	res, e = GetMetaOrgs(ids)
	return
}

func CityAddToIndex(city string, ids ...int64) error {
	bts, e := cities.Get([]byte(city), nil)
	if e != nil {
		return e
	}

	res, e := DecodeIntArr(bts)
	if e != nil {
		return e
	}

	for _, v := range ids {
		if !res.Has(v) {
			res = append(res, v)
		}
	}
	bts, e = res.GobEncode()
	if e != nil {
		return e
	}

	return cities.Put([]byte(city), bts, nil)
}

func CityTruncateIndex(city string) error {
	return cities.Delete([]byte(city), nil)
}

var (
	imu sync.Mutex
)

func IndexRegion(id int) error {
	imu.Lock()
	defer imu.Unlock()
	bid := fmt.Sprint(id)
	if len(bid) == 1 {
		bid = "0" + bid
	}
	if len(bid) > 2 {
		bid = bid[1:]
	}
	iter := lookdb.NewIterator(nil, nil)
	for iter.Next() {
		k := iter.Key()
		//dup[com.StrTo(string([]byte{k[3], k[4]})).MustInt()] = struct{}{}

		if len(k) > 4 && k[3] == bid[0] && k[4] == bid[1] {
			color.Green("index %s", iter.Key())
			//n++
			s, e := GetSvul(string(k), "", 0)
			if e != nil {
				color.Red("%s", e)
			}
			o := s.ToOrg(0, "")
			e = CityAddToIndex(o.City, o.OGRN)
			if e != nil {
				color.Red("%s", e)
			}
		}
	}
	return nil
}
