package models

import (
	"encoding/json"
	"fmt"
	ttt "github.com/ulrf/ulrf/modules/titles"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/fatih/color"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	metaDb     *leveldb.DB
	metaBoltDb *mDb

	metaBuckName = []byte("meta")
)

type mDb struct {
	bdb *bolt.DB
}

func (db *mDb) Get(key []byte) ([]byte, error) {
	var (
		res []byte
	)
	e := db.bdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(metaBuckName)
		res = b.Get(key)
		return nil
	})
	return res, e
}

func (db *mDb) Put(key, val []byte) error {
	e := db.bdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(metaBuckName)
		return b.Put(key, val)
	})
	return e
}

func NewMetaDb() {
	var e error
	mBoltDb, e := bolt.Open("data/meta.bolt", 0777, nil)
	if e != nil {
		panic(e)
	}
	mBoltDb.NoSync = true

	e = mBoltDb.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists(metaBuckName)
		return e
	})
	if e != nil {
		panic(e)
	}
	metaBoltDb = &mDb{bdb: mBoltDb}
	/*o := &opt.Options{}
	o.BlockCacheCapacity = 512 * opt.MiB
	o.WriteBuffer = 4 * opt.MiB
	o.CompactionTableSize = 16 * opt.MiB
	o.OpenFilesCacheCapacity = 50
	var e error
	metaDb, e = leveldb.OpenFile("data/meta.level", o)
	if e != nil {
		panic(e)
	}*/
}

func GetOrg(ogrn int64) (o *Org, e error) {

	if ogrn == 0 {
		return nil, fmt.Errorf("Ogrn is 0")
	}

	v, _ := metaBoltDb.Get(i2b(ogrn))
	if v == nil {
		//color.White("Not found with meta, try svul %d", ogrn)
		s, e := GetSvul(fmt.Sprint(ogrn), "", 0)
		if e != nil {
			return nil, e
		}
		oo := s.ToOrg(0, "")
		bts, e := json.Marshal(oo)
		if e != nil {
			color.Red("%s", e)
		} else {
			e = metaBoltDb.Put(i2b(ogrn), bts)
			if e != nil {
				color.Red("%s", e)
			}
		}
		return &oo, nil
	}

	return UnmarshalOrg(v)
}

func UnmarshalOrg(bts []byte) (o *Org, e error) {
	o = new(Org)
	e = json.Unmarshal(bts, o)
	return
}

func GetMetaOrgs(ogrns []int64) (orgs []*Org, e error) {
	start := time.Now()
	titles, e := ttt.GetTitles(ogrns)
	if e != nil {
		return nil, e
	}
	for i, v := range titles {
		if len(ogrns) > i && v != "" {
			orgs = append(orgs, &Org{OGRN: ogrns[i], FullName: v})
		} else {
			color.Cyan("Error, try get classic")
			return getOrgsByIds(ogrns)
		}
	}
	printNeedSince(start, time.Millisecond*450, "Get Meta with titley")
	return
	return getOrgsByIds(ogrns)
	color.Green("%d, ", ogrns)
	for _, ogrn := range ogrns {
		v, _ := metaBoltDb.Get(i2b(ogrn))
		if v == nil {
			t, _ := GetTitle(ogrn)
			o := new(Org)
			o.OGRN = ogrn
			o.FullName = t
			orgs = append(orgs, o)
		} else {
			o, _ := UnmarshalOrg(v)
			if o != nil {
				orgs = append(orgs, o)
			}
		}
	}
	return
}

func getOrgsByIds(ogrns []int64) (orgs []*Org, e error) {
	start := time.Now()
	metaBoltDb.bdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(metaBuckName)

		for _, ogrn := range ogrns {
			v := b.Get(i2b(ogrn))
			if v == nil {
				//color.White("Not found with meta, try svul %d", ogrn)
				o, e := getTitleAndSave(ogrn)
				if e != nil {
					color.Red("%s", e)
					continue
				}
				orgs = append(orgs, o)
			} else {
				o, e := UnmarshalOrg(v)
				if e != nil {
					color.Red("%s", e)
				}

				if strings.TrimSpace(o.FullName) == "" {
					o, e = getTitleAndSave(ogrn)
					if e != nil {
						color.Red("%s", e)
						continue
					}
				}
				orgs = append(orgs, o)
			}
		}
		return nil
	})
	printNeedSince(start, time.Millisecond*350, "GetMeta")

	return
}

func getTitleAndSave(ogrn int64) (o *Org, e error) {
	var (
		title string
	)
	title, e = GetTitle(ogrn)
	if e != nil {
		return
	}
	if title == "" {
		color.Red("Title is nile!!!! %d", ogrn)
	}
	o = new(Org)
	o.OGRN = ogrn
	o.FullName = title

	go func(o *Org) {
		start := time.Now()
		bts, _ := json.Marshal(o)
		metaBoltDb.Put(i2b(ogrn), bts)
		if since := time.Since(start); since > 200*time.Millisecond {
			color.Yellow("[%s],Put single meta in boltDb", since)
		}
	}(o)
	return
}

func GetOrgs(ogrns []int64) (orgs []*Org, e error) {
	color.Green("%d, ", ogrns)
	for _, ogrn := range ogrns {
		o, _ := GetOrg(ogrn)
		orgs = append(orgs, o)
	}
	return
}

func RangeMetaOrgs(page int) (orgs []*Org, total int, e error) {
	ids, t, e := OgrnsRange(page)
	if e != nil {
		return
	}
	total = t
	orgs, e = GetMetaOrgs(ids)
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

	metaBoltDb.bdb.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket(metaBuckName)

		//skip := rand.Intn(100)

		c := b.Cursor()

		c.Seek(i2b(orgId))

		for k, v := c.Next(); k != nil; k, v = c.Next() {
			o, e := UnmarshalOrg(v)
			if e != nil {
				continue
			}
			orgs = append(orgs, o)
			if len(orgs) == l {
				break
			}
		}

		return nil
	})

	//ok := it.Seek(i2b(orgId))
	//if !ok {
	//orgs, _, e = RangeOrgs(rand.Intn(2 << 20))
	//if len(orgs) > l {
	//	return orgs[:l], e
	//} else {
	//	return orgs, e
	//}

	//}
	//var lim = 1000
	//for len(orgs) < l && lim > 0 {
	//	if it.Next() {
	//		o, e := UnmarshalOrg(it.Value())
	//		if e != nil {
	//			continue
	//		}
	//		orgs = append(orgs, o)
	//	}
	//}
	return
}
