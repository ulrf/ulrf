package models

import (
	"bytes"
	"encoding/binary"
	"github.com/boltdb/bolt"
	"github.com/fatih/color"
	"github.com/zhuharev/raddress"
	"io"
)

var (
	regionsDb         *bolt.DB
	regionsBucketName = []byte("regions")
)

func Locality() raddress.Locality {
	return raddress.Loc
}

func NewRegionsDb() {
	var e error
	regionsDb, e = bolt.Open("data/regions.bolt", 0777, nil)
	if e != nil {
		panic(e)
	}
}

func RegionsGetRange(regionId int, page int) (res []int64, l int, e error) {
	var (
		bts []byte
		bid = make([]byte, 8)
	)
	page = page - 1
	regionsDb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(regionsBucketName)
		bts = b.Get([]byte{byte(regionId)})
		l = len(bts) / 8
		return nil
	})
	r := bytes.NewReader(bts)
	skip := page * 10
	n := 0
	for _, e = r.Read(bid); e == nil; _, e = r.Read(bid) {
		n++
		if n < skip {
			continue
		}
		if len(res) >= 10 {
			continue
		}
		ui := binary.BigEndian.Uint64(bid)
		if ui == 0 {
			color.Red("[regions.go] ogrn is 0")
		}
		res = append(res, int64(ui))
		bid = make([]byte, 8)
	}
	if e == io.EOF {
		e = nil
	}
	color.Green("[page] %d", page*10)
	return
}
