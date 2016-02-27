package models

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"math/rand"
)

var (
	metaDb *leveldb.DB
)

func NewMetaDb() {
	o := &opt.Options{}
	o.BlockCacheCapacity = 512 * opt.MiB
	o.WriteBuffer = 4 * opt.MiB
	o.CompactionTableSize = 16 * opt.MiB
	o.OpenFilesCacheCapacity = 400
	var e error
	metaDb, e = leveldb.OpenFile("data/meta.level", o)
	if e != nil {
		panic(e)
	}
}

func GetOrg(ogrn int64) (o *Org, e error) {

	v, e := metaDb.Get(i2b(ogrn), nil)
	if e != nil {
		if e == leveldb.ErrNotFound {
			s, e := GetSvul(fmt.Sprint(ogrn), "", 0)
			if e != nil {
				return nil, e
			}
			oo := s.ToOrg(0, "")
			return &oo, nil
		}
		return nil, e
	}

	/*	{
		r := bytes.NewReader(v)
		gr, e := gzip.NewReader(r)
		if e != nil {
			return nil, e
		}
		bts, e := ioutil.ReadAll(gr)
		if e != nil {
			return nil, e
		}
		v = bts
	}*/

	return UnmarshalOrg(v)
}

func UnmarshalOrg(bts []byte) (o *Org, e error) {
	o = new(Org)
	e = json.Unmarshal(bts, o)
	return
}

func GetOrgs(ogrns []int64) (orgs []*Org, e error) {
	color.Green("%d, ", ogrns)
	for _, ogrn := range ogrns {
		o, e := GetOrg(ogrn)
		if e != nil {
			if e != leveldb.ErrNotFound {
				return nil, e
			} else {
				continue
			}
		}
		orgs = append(orgs, o)
	}
	return
}

func RangeOrgs(page int) (orgs []*Org, total int, e error) {
	ids, t, e := OgrnsRange(page)
	if e != nil {
		return
	}
	total = t
	orgs, e = GetOrgs(ids)
	return
}

func SimilarOrgs(orgId int64, l int) (orgs []*Org, e error) {
	it := metaDb.NewIterator(nil, nil)
	ok := it.Seek(i2b(orgId))
	if !ok {
		orgs, _, e = RangeOrgs(rand.Intn(2 << 20))
		if len(orgs) > l {
			return orgs[:l], e
		} else {
			return orgs, e
		}
	}
	var lim = 1000
	for len(orgs) < l && lim > 0 {
		if it.Next() {
			o, e := UnmarshalOrg(it.Value())
			if e != nil {
				continue
			}
			orgs = append(orgs, o)
		}
		lim--
	}
	return
}
