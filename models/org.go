package models

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/Unknwon/com"
	"io"
	"strings"
)

type Org struct {
	Id          int64
	DocLocation string `json:"doc_location" xorm:"TEXT"`
	DocId       int    `json:"doc_id"`

	FullName  string `json:"full_name" xorm:"TEXT"`
	ShortName string `json:"short_name" xorm:"TEXT"`
	RegionId  int    `json:"region_id" xorm:"index"`
	City      string `xorm:"index"`

	INN    int64    `json:"inn" xorm:"inn"`
	OGRN   int64    `json:"ogrn" xorm:"unique index 'ogrn'"`
	OKVED  string   `json:"okved" xorm:"index 'okved'"`
	OKVEDS []string `xorm:"'okveds'"`
	OPF    int      `json:"opf" xorm:"'opf'"`
	KPP    int64    `json:"kpp" xorm:"'kpp'"`
}

func NewOrgFromCsv(in []string) (o Org, e error) {
	if len(in) != 13 {
		e = fmt.Errorf("error len != %d, got %d (%v)", 13, len(in), in)
		return
	}
	o.Id = com.StrTo(in[0]).MustInt64()
	o.DocLocation = in[1]
	o.DocId = com.StrTo(in[2]).MustInt()

	o.FullName = in[3]
	o.ShortName = in[4]
	o.RegionId = com.StrTo(in[5]).MustInt()
	o.City = in[6]
	o.INN = com.StrTo(in[7]).MustInt64()
	o.OGRN = com.StrTo(in[8]).MustInt64()
	o.OKVED = in[9]
	var okveds []string
	e = json.Unmarshal([]byte(strings.Replace(in[10], `""`, `"`, -1)), &okveds)
	if e != nil {
		return
	}
	o.OKVEDS = okveds
	o.OPF = com.StrTo(in[11]).MustInt()
	o.KPP = com.StrTo(in[12]).MustInt64()
	return
}

// type ForIndex string
type ForIndex struct {
	FullName  string //`json:"full_name" xorm:"TEXT"`
	ShortName string //`json:"short_name" xorm:"TEXT"`
}

//type ForOKVEDIndex string

type ForOKVEDIndex struct {
	OKVED  string
	OKVEDS []string
}

//type ForCityIndex string

type ForCityIndex struct {
	City string
}

func (o Org) ForIndex() ForIndex {
	return ForIndex{
		FullName:  o.FullName,
		ShortName: o.ShortName,
	}
	//return o.ShortName + " " + o.FullName
}

func (o Org) ForCityIndex() ForCityIndex {
	return ForCityIndex{
		City: o.City,
	}
	//return o.City
}

func (o Org) ForOKVEDIndex() ForOKVEDIndex {
	return ForOKVEDIndex{
		o.OKVED,
		o.OKVEDS,
	}
	//return o.OKVED + " " + strings.Join(o.OKVEDS, " ")
}

func (o Org) GobEncode() ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	// Id
	e := wInt32(buf, o.Id)

	e = wString(buf, o.DocLocation)
	e = wInt16(buf, uint16(o.DocId))

	e = wString(buf, o.FullName)
	e = wString(buf, o.ShortName)
	_, e = buf.Write([]byte{byte(o.RegionId)})
	e = wString(buf, o.City)
	e = wInt64(buf, o.INN)
	e = wInt64(buf, o.OGRN)
	e = wString(buf, o.OKVED)
	e = wOkveds(buf, o.OKVEDS)
	e = wInt16(buf, uint16(o.OPF))
	e = wInt64(buf, o.KPP)

	return buf.Bytes(), e

}

func (o *Org) GobDecode(b []byte) (e error) {
	if len(b) == 0 {
		return fmt.Errorf("[org.go] Decode is nil")
	}
	buf := bytes.NewReader(b)

	// Id
	o.Id = int64(rInt(buf))
	o.DocLocation = rString(buf)
	o.DocId = rInt16(buf)
	o.FullName = rString(buf)
	o.ShortName = rString(buf)
	o.RegionId = rByte(buf)
	o.City = rString(buf)
	o.INN = rInt64(buf)
	o.OGRN = rInt64(buf)
	o.OKVED = rString(buf)
	o.OKVEDS = rOkveds(buf)
	o.OPF = rInt16(buf)
	o.KPP = rInt64(buf)

	return e

}

func rInt64(r io.Reader) int64 {
	bid := make([]byte, 8)
	r.Read(bid)
	return int64(b2i(bid))
}

func rByte(r io.Reader) int {
	bid := make([]byte, 1)
	r.Read(bid)
	return int(bid[0])
}

func rInt(r io.Reader) int {
	bid := make([]byte, 4)
	r.Read(bid)
	return int(b2i32(bid))
}

func rInt16(r io.Reader) int {
	bid := make([]byte, 2)
	r.Read(bid)
	return int(b2i16(bid))
}

func rString(r io.Reader) string {
	bid := make([]byte, 2)
	r.Read(bid)
	l := b2i16(bid)
	bid = make([]byte, l)
	r.Read(bid)
	return string(bid)
}

func rOkveds(r io.Reader) (o []string) {
	l := rInt16(r)
	for i := 0; i < int(l); i++ {
		o = append(o, rString(r))
	}
	return
}

func wString(w io.Writer, s string) (e error) {
	l := uint16(len(s))
	e = wInt16(w, l)
	if e != nil {
		return
	}
	_, e = w.Write([]byte(s))
	return
}

func i2b(i int64) []byte {
	var bid = make([]byte, 8)
	binary.BigEndian.PutUint64(bid, uint64(i))
	return bid
}

func b2i16(b []byte) int64 {
	return int64(binary.BigEndian.Uint16(b))
}

func b2i32(b []byte) int64 {
	return int64(binary.BigEndian.Uint32(b))
}

func b2i(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func wInt16(w io.Writer, i uint16) (e error) {
	bid := i2b(int64(i))
	_, e = w.Write(bid[6:])
	return
}

func wInt32(w io.Writer, i int64) (e error) {
	bid := i2b(int64(i))
	_, e = w.Write(bid[4:])
	return
}

func wInt64(w io.Writer, i int64) (e error) {
	bid := i2b(i)
	_, e = w.Write(bid)
	return
}

func wOkveds(w io.Writer, okveds []string) (e error) {
	l := len(okveds)
	e = wInt16(w, uint16(l))
	if e != nil {
		return
	}
	for _, v := range okveds {
		e = wString(w, v)
	}
	return
}
