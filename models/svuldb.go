package models

import (
	"bytes"
	"compress/gzip"
	"github.com/Unknwon/com"
	"github.com/boltdb/bolt"
	"github.com/fatih/color"
	"github.com/pquerna/ffjson/ffjson"
	"sync"
)

var (
	svulDb *mDb
)

func NewSvulDb() {
	var e error
	mBoltDb, e := bolt.Open("data/svuls.bolt", 0777, nil)
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
	svulDb = &mDb{bdb: mBoltDb}
}

func GetSvulFromBoltDb(ogrn string) (s *Svul, e error) {
	var (
		res *Svul
		bts []byte

		key = i2b(com.StrTo(ogrn).MustInt64())
	)
	bts, e = svulDb.Get(key)
	if e != nil {
		return
	}
	var (
		b = bytes.NewReader(bts)
		r *gzip.Reader
	)
	r, e = gzip.NewReader(b)
	if e != nil {
		return
	}
	dec := ffjson.NewDecoder()
	res = new(Svul)
	e = dec.DecodeReader(r, res)
	if e != nil {
		return
	}
	s = res
	return
}

func SetSvultoBoltDb(s *Svul) error {
	var (
		key = i2b(com.StrTo(s.OGRN).MustInt64())
	)
	bts, e := ffjson.Marshal(s)
	if e != nil {
		return e
	}
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, e = w.Write(bts)
	if e != nil {
		if w != nil {
			w.Close()
		}
		return e
	}
	w.Close()
	return svulDb.Put(key, b.Bytes())
}

func SetSvulstoBoltDb(ss []Svul) error {
	return svulDb.bdb.Update(func(tx *bolt.Tx) error {
		for _, s := range ss {
			var (
				key = i2b(com.StrTo(s.OGRN).MustInt64())
				b   = tx.Bucket(metaBuckName)
			)
			bts, e := ffjson.Marshal(s)
			if e != nil {
				return e
			}
			var buf bytes.Buffer
			w := gzip.NewWriter(&buf)
			_, e = w.Write(bts)
			if e != nil {
				if w != nil {
					w.Close()
				}
				return e
			}
			w.Close()
			e = b.Put(key, buf.Bytes())
			if e != nil {
				return e
			}
		}
		return nil
	})

}

type NamedMu struct {
	Name string
	sync.Mutex
}

type Mutexes []NamedMu

func (m *Mutexes) Lock(name string) {
	color.Cyan("Lock %s", name)
	var (
		found = false
	)
	for _, v := range *m {
		if v.Name == name {
			found = true
			v.Lock()
		}
	}
	if !found {
		nm := NamedMu{}
		nm.Name = name
		nm.Lock()
		ar := append(*m, nm)
		m = &ar
	}
}

func (m Mutexes) Unlock(name string) {
	color.Cyan("Unlock %s", name)

	for _, v := range m {
		if v.Name == name {
			v.Unlock()
		}
	}
}

var (
	smusi = new(Mutexes)
)

func getSvul(ogrn string, docLoc string, id int) (s *Svul, e error) {
	//start := time.Now()

	s, e = GetSvulFromBoltDb(ogrn)
	if s != nil {
		//color.Green("found with ogrn (%s)", time.Since(start))
		return s, nil
	}

	if docLoc == "" {
		color.Green("Start Lookup")
		docLoc, id, e = LookUpLoc(ogrn)
		if e != nil {
			return s, e
		}
	}

	s, e = GetOneFromZipDb(docLoc, id)
	if e != nil {
		color.Red("%s", e)
		return s, nil
	}
	//color.Green("found in xmlzip (%s)", time.Since(start))

	return s, e
}
