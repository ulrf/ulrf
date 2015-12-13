package models

import (
	"github.com/fatih/color"
)

type Org struct {
	Id          int64
	DocLocation string `json:"doc_location" xorm:"TEXT"`
	DocId       int    `json:"doc_id"`

	FullName  string `json:"full_name" xorm:"TEXT"`
	ShortName string `json:"short_name" xorm:"TEXT"`
	RegionId  int    `json:"region_id" xorm:"index"`
	City      string `xorm:"index"`

	INN    string   `json:"inn" xorm:"inn"`
	OGRN   string   `json:"ogrn" xorm:"unique index 'ogrn'"`
	OKVED  string   `json:"okved" xorm:"index 'okved'"`
	OKVEDS []string `xorm:"'okveds'"`
	OPF    string   `json:"opf" xorm:"'opf'"`
	KPP    string   `json:"kpp" xorm:"'kpp'"`
}

//type ForIndex string

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

func GetOrg(id int64) (o *Org, e error) {
	o = new(Org)
	_, e = x.Id(id).Get(o)
	color.Cyan("%v", o)
	return
}
