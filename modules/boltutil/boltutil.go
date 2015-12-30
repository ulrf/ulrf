package boltutil

import (
	"bytes"
	"github.com/boltdb/bolt"
	"github.com/fatih/color"
	"github.com/ulrf/ulrf/models"
	"io/ioutil"
	"strings"
)

type Prefix byte

var (
	// search index
	Title  Prefix = 't'
	Okved  Prefix = 'o'
	Region Prefix = 'r'
	Ogrn   Prefix = 'i'
	Data   Prefix = 'd'
)

var (
	ogrnsKeyName = []byte("okveds")
)

func (p Prefix) ToSlice() []byte {
	return []byte{byte(p)}
}

func (p Prefix) Byte() byte {
	return byte(p)
}

type DB struct {
	bdb *bolt.DB
}

func New(path string) (*DB, error) {
	bdb, e := bolt.Open(path, 0777, &bolt.Options{NoGrowSync: true})
	if e != nil {
		return nil, e
	}

	e = bdb.Update(func(tx *bolt.Tx) error {
		arr := []byte{Title.Byte(), Okved.Byte(), Region.Byte(), Ogrn.Byte(), Data.Byte()}
		for _, v := range arr {
			_, e := tx.CreateBucketIfNotExists([]byte{v})
			if e != nil {
				return e
			}
		}
		return nil
	})
	if e != nil {
		return nil, e
	}

	db := new(DB)
	db.bdb = bdb
	return db, e
}

func (db *DB) Insert(org models.Org) error {
	color.Green("Indexing %s", org.ShortName)
	bts, e := org.GobEncode()
	if e != nil {
		return e
	}

	e = db.IndexTitle(org.OGRN, org.ShortName)
	if e != nil {
		return e
	}

	e = db.IndexRegion(org.OGRN, byte(org.RegionId))
	if e != nil {
		return e
	}

	e = db.IndexOkveds(org.OGRN, append(org.OKVEDS, org.OKVED))
	if e != nil {
		return e
	}

	e = db.IndexOgrn(org.OGRN)
	if e != nil {
		return e
	}

	return db.put(Data.ToSlice(), i2b(org.OGRN), bts)
}

func (db *DB) get(bucket, key []byte) (res []byte, e error) {
	e = db.bdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		res = b.Get([]byte(key))
		return nil
	})
	return
}

func (db *DB) GetTitle(title string) (res []byte, e error) {
	return db.get(Title.ToSlice(), []byte(title))
}

func (db *DB) put(bucket, key, value []byte) (e error) {
	e = db.bdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		e = b.Put(key, value)
		return nil
	})
	return
}

// todo optimize to 1 request to base
func (db *DB) append(bucket, key, value []byte) (e error) {
	bts, e := db.get(bucket, key)
	if e != nil {
		return e
	}
	if len(value) == 0 {
		return nil
	}
	color.Green("%d", b2i(value))
	bts = append(bts, value...)
	return db.put(bucket, key, bts)
}

// func (db *DB) hasInt(bucket, key)
func (db *DB) IndexTitle(id int64, title string) (e error) {
	for _, word := range strings.Fields(cleanString(title)) {
		e = db.append(Title.ToSlice(), []byte(word), i2b(id))
		if e != nil {
			return
		}
	}
	return
}

func (db *DB) IndexRegion(id int64, regionId byte) (e error) {
	return db.append(Region.ToSlice(), []byte{regionId}, i2b(id))
}

func (db *DB) IndexOkveds(id int64, okveds []string) (e error) {
	for _, o := range okveds {
		e = db.append(Okved.ToSlice(), []byte(o), i2b(id))
		if e != nil {
			return
		}
	}
	return
}

func (db *DB) IndexOgrn(id int64) error {
	return db.append(Ogrn.ToSlice(), ogrnsKeyName, i2b(id))
}

func (db *DB) PutDate(id int64, data []byte) error {
	return db.put(Data.ToSlice(), i2b(id), data)
}

func (db *DB) LimitOgrn(limit int, offsets ...int) ([]byte, error) {
	offset := 0
	if len(offsets) > 0 {
		offset = offsets[0] * 8
	}
	b, e := db.get(Ogrn.ToSlice(), ogrnsKeyName)
	if e != nil {
		return nil, e
	}
	end := offset*8 + limit*8
	if end > len(b) {
		end = len(b) - 1
	}
	return b[offset:end], nil
}

func (db *DB) LoadXml(path string) error {
	bts, e := ioutil.ReadFile(path)
	if e != nil {
		return e
	}
	buf := bytes.NewReader(bts)
	bts, e = models.Fixed(buf)
	if e != nil {
		return e
	}
	svuls, e := models.UnmarshalAll(bts)
	if e != nil {
		return e
	}
	for i, v := range svuls {
		org := v.ToOrg(i, path)
		e = db.Insert(org)
		if e != nil {
			return e
		}
	}
	return nil
}
