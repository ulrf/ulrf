package torefactor

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
)

var (
	OKVEDAPI OKVEDS
)

type okvedXML struct {
	Okveds OKVEDS `xml:"nsiOKVEDList>nsiOKVED"`
}

type OKVEDS []OKVED

func (o OKVEDS) Get(code string) OKVED {
	for _, v := range o {
		if v.Code == code {
			return v
		}
	}
	return OKVED{}
}

func (o OKVEDS) GetById(id int) (OKVED, bool) {
	for _, v := range o {
		if v.Id == id {
			return v, true
		}
	}
	return OKVED{}, false
}

func (o OKVEDS) Parent(code string) (OKVED, bool) {
	child := o.Get(code)
	if child.Text == "" {
		return OKVED{}, false
	}
	return o.GetById(child.ParentId)
}

func (o OKVEDS) Parents() (p OKVEDS) {
	for _, v := range o {
		if v.ParentId == 0 {
			p = append(p, v)
		}
	}
	return
}

func (o OKVEDS) SubSections(id int) (p OKVEDS) {
	for _, v := range o {
		if v.SubSection != "" && v.ParentId == id {
			p = append(p, v)
		}
	}
	return
}

func (o OKVEDS) Childs(id int) (p OKVEDS) {
	for _, v := range o {
		if v.ParentId == id {
			p = append(p, v)
		}
	}
	return
}

type OKVED struct {
	Id         int    `xml:"id"`
	ParentId   int    `xml:"parentId"`
	Code       string `xml:"code"`
	Text       string `xml:"name"`
	Section    string `xml:"section"`
	SubSection string `xml:"subsection"`
	Actual     bool   `xml:"actual"`
}

func init() {
	bts, e := ioutil.ReadFile("okved.xml")
	if e != nil {
		panic(e)
	}
	bts = bytes.Replace(bts, []byte("oos:"), []byte{}, -1)
	var v okvedXML
	e = xml.Unmarshal(bts, &v)
	if e != nil {
		panic(e)
	}
	OKVEDAPI = v.Okveds
}

//func main() {}
