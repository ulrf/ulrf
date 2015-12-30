package models

import (
	"encoding/xml"
	"github.com/Unknwon/com"
)

type Svul struct {
	//XMLName xml.Name `xml:"СвЮЛ"`
	DateVip     string `xml:"ДатаВып,attr"`
	OGRN        string `xml:"ОГРН,attr"`
	DateOGRN    string `xml:"ДатаОГРН,attr"`
	KodOpf      string `xml:"КодОПФ,attr"`
	FullNameOpf string `xml:"ПолнНаимОПФ,attr"`
	SprOpf      string `xml:"СпрОПФ,attr"`

	INN string `xml:"ИНН,attr"`
	KPP string `xml:"КПП,attr"`

	Name struct {
		FullName  string `xml:"НаимЮЛПолн,attr"`
		ShortName string `xml:"НаимЮЛСокр,attr"`
	} `xml:"СвНаимЮЛ"`

	OKVED struct {
		Osn struct {
			Code string `xml:"КодОКВЭД,attr"`
			Name string `xml:"НаимОКВЭД,attr"`
			Grn  struct {
				Grn  string `xml:"ГРН,attr"`
				Date string `xml:"ДатаЗаписи,attr"`
			} `xml:"ГРНДата"`
		} `xml:"СвОКВЭДОсн"`
		Dop []struct {
			Code string `xml:"КодОКВЭД,attr"`
			Name string `xml:"НаимОКВЭД,attr"`
		} `xml:"СвОКВЭДДоп"`
	} `xml:"СвОКВЭД"`

	EGRYL []struct {
		Id  string `xml:"ИдЗап,attr"`
		Vid struct {
			CodeSPVZ string `xml:"КодСПВЗ,attr"`
			Name     string `xml:"НаимВидЗап,attr"`
		} `xml:"ВидЗап,attr"`
		RegOrg struct {
			Name string `xml:"НаимНО,attr"`
			Code int    `xml:"КодНО,attr"`
		} `xml:"СвРегОрг"`
		Cert struct {
			Serial string `xml:"Серия,attr"`
			Date   string `xml:"ДатаВыдСвид,attr"`
			Status struct {
				GrnUp []GrnUp
			}
		} `xml:"СвСвид"`

		Doc []struct {
			Name   string `xml:"НаимДок"`
			Date   string `xml:"ДатаДок"`
			NumDoc string `xml:"НомДок"`
		} `xml:"СведПредДок"`
		Grn   Grn
		GrnUp []GrnUp
	} `xml:"СвЗапЕГРЮЛ"`

	Address struct {
		Address struct {
			Index    string `xml:"Индекс,attr"`
			RegionId string `xml:"КодРегион,attr"`
			City     struct {
				Type string `xml:"ТипГород,attr"`
				Name string `xml:"НаимГород,attr"`
			} `xml:"Город"`
			Kladr    string `xml:"КодАдрКладр,attr"`
			House    string `xml:"Дом,attr"`
			Korpus   string `xml:"Корпус,attr"`
			Kvartira string `xml:"Кварт,attr"`
			Region   struct {
				Type string `xml:"ТипРегион,attr"`
				Name string `xml:"НаимРегион,attr"`
			} `xml:"Регион"`
			Street struct {
				Type string `xml:"ТипУлица,attr"`
				Name string `xml:"НаимУлица,attr"`
			} `xml:"Улица"`
		} `xml:"АдресРФ"`
	} `xml:"СвАдресЮЛ"`

	Addresses []struct {
		Index    string `xml:"Индекс,attr"`
		RegionId string `xml:"КодРегион,attr"`
		City     struct {
			Type string `xml:"ТипГород"`
			Name string `xml:"НаимГород"`
		} `xml:"Город"`
		KladrCode string `xml:"КодАдрКладр",attr`
		House     string `xml:"КодАдрКладр,attr"`
		Korpus    string `xml:"Корпус,attr"`
		Kvartira  string `xml:"Кварт,attr"`
		Region    struct {
			Type string `xml:"ТипРегион,attr"`
			Name string `xml:"НаимРегион,attr"`
		} `xml:"Регион"`
		Street struct {
			Type string `xml:"ТипУлица,attr"`
			Name string `xml:"НаимУлица,attr"`
		} `xml:"Улица"`
	} `xml:"АдресРФ"`

	Born struct {
		Code     string `xml:"КодСпОбрЮЛ,attr"`
		Name     string `xml:"НаимСпОбрЮЛ,attr"`
		OGRN     string `xml:"ОГРН,attr"`
		DateOgrn string `xml:"ДатаОГРН,attr"`
		RegNum   string `xml:"РегНом,attr"`
		DateReg  string `xml:"ДатаРег,attr"`
		NameRO   string `xml:"НаимРО,attr"`
		BornYL   struct {
			Code  string `xml:"КодСпОбрЮЛ,attr"`
			Title string `xml:"НаимСпОбрЮЛ,attr"`
		} `xml:"СпОбрЮЛ"`
	} `xml:"СпОбрЮЛ"`

	Dead struct {
		Date   string `xml:"ДатаПрекрЮЛ,attr"`
		Grn    Grn
		RegOrg struct {
			Name    string `xml:"НаимНО,attr"`
			Code    int    `xml:"КодНО,attr"`
			Address string `xml:"АдрРО,attr"`
		} `xml:"СвРегОрг"`
		Why struct {
			Code    string `xml:"КодСпПрекрЮЛ,attr"`
			WhyName string `xml:"НаимСпПрекрЮЛ,attr"`
		} `xml:"СпПрекрЮЛ"`
	} `xml:"СвПрекрЮЛ"`

	RegOrg struct {
		Name    string `xml:"НаимНО,attr"`
		Code    int    `xml:"КодНО,attr"`
		Address string `xml:"АдрРО,attr"`
	} `xml:"СвРегОрг"`

	Accounting struct {
		INN  string `xml:"ИНН,attr"`
		KPP  string `xml:"КПП,attr"`
		Date string `xml:"ДатаПостУч,attr"`

		NalogOrg struct {
			Code string `xml:"КодНО,attr"`
			Name string `xml:"НаимНО,attr"`
		} `xml:"СвНО"`
	} `xml:"СвУчетНО"`

	FSS struct {
		Num  string `xml:"РегНомФСС,attr"`
		Date string `xml:"ДатаРег,attr"`
		Org  struct {
			Code string `xml:"КодФСС,attr"`
			Name string `xml:"НаимФСС,attr"`
		}
	} `xml:"СвРегФСС"`

	Capital struct {
		Name string `xml:"НаимВидКап,attr"`
		Sum  string `xml:"СумКап,attr"`
	} `xml:"СвУстКап"`

	Positions struct {
		FizFace struct {
			LastName   string `xml:"Фамилия,attr"`
			FirstName  string `xml:"Имя,attr"`
			Patronymic string `xml:"Отчество,attr"`
		} `xml:"СвФЛ"`
		Position struct {
			Type     string `xml:"ВидДолжн,attr"`
			NameType string `xml:"НаимВидДолжн,attr"`
			Name     string `xml:"НаимДолжн,attr"`
		} `xml:"СвДолжн"`
	} `xml:"СведДолжнФЛ"`

	Founders struct {
		Founder []struct {
			Name struct {
				OGRN     string `xml:"ОГРН,attr"`
				INN      string `xml:"ИНН,attr"`
				FullName string `xml:"НаимЮЛПолн,attr"`
				Grn      Grn
				GrnUp    GrnUp
			} `xml:"НаимИННЮЛ"`
			Part struct {
				Sum   string `xml:"НоминСтоим,attr"`
				Grn   Grn
				GrnUp GrnUp
			} `xml:"ДоляУстКап"`
		} `xml:"УчрЮЛРос"`
	} `xml:"СвУчредит"`

	SvDerjReestrAO struct {
		DerjReestrAO struct {
			OGRN     string `xml:"ОГРН,attr"`
			INN      string `xml:"ИНН,attr"`
			FullName string `xml:"НаимЮЛПолн,attr"`
			Grn      Grn
		} `xml:"ДержРеестрАО"`
	} `xml:"СвДержРеестрАО"`

	License []struct {
		NumLicense    string   `xml:"НомЛиц,attr"`
		Date          string   `xml:"ДатаЛиц,attr"`
		DateStart     string   `xml:"ДатаНачЛиц,attr"`
		KindActivity  []string `xml:"НаимЛицВидДеят"`
		PlaceActivity []string `xml:"МестоДейстЛиц"`
		Org           string   `xml:"ЛицОргВыдЛиц"`
	} `xml:"СвЛицензия"`

	Unit struct {
		Filial []struct {
			Name struct {
				FullName string `xml:"НаимПолн,attr"`
				Addr     struct {
					Index    string `xml:"Индекс,attr"`
					RegionId string `xml:"КодРегион,attr"`
					City     struct {
						Type string `xml:"ТипГород,attr"`
						Name string `xml:"НаимГород,attr"`
					} `xml:"Город"`
					Kladr    string `xml:"КодАдрКладр,attr"`
					House    string `xml:"Дом,attr"`
					Korpus   string `xml:"Корпус,attr"`
					Kvartira string `xml:"Кварт,attr"`
					Region   struct {
						Type string `xml:"ТипРегион,attr"`
						Name string `xml:"НаимРегион,attr"`
					} `xml:"Регион"`
					Street struct {
						Type string `xml:"ТипУлица,attr"`
						Name string `xml:"НаимУлица,attr"`
					} `xml:"Улица"`
				} `xml:"АдрМНРФ"`
			} `xml:"СвНаим"`
		} `xml:"СвФилиал"`
		Agency []struct {
			Name struct {
				FullName string `xml:"НаимПолн,attr"`
				Addr     struct {
					OKSM        string `xml:"ОКСМ,attr"`
					CountryName string `xml:"НаимСтран,attr"`
					Addr        string `xml:"АдрИн,attr"`
				} `xml:"АдрМНИн"`
			} `xml:"СвНаим"`
		}
	} `xml:"СвПодразд"`

	PF struct {
		Num   string `xml:"РегНомПФ,attr"`
		PFOrg struct {
			Code string `xml:"КодПФ,attr"`
			Name string `xml:"НаимПФ,attr"`
		} `xml:"СвОргПФ"`
	} `xml:"СвРегПФ"`
}

type Grn struct {
	XMLName xml.Name `xml:"ГРНДата"`
	Id      string   `xml:"ИдЗап,attr"`
	Grn     string   `xml:"ГРН,attr"`
	Date    string   `xml:"ДатаЗаписи,attr"`
}
type GrnUp struct {
	XMLName xml.Name `xml:"ГРНДатаИспр"`
	Id      string   `xml:"ИдЗап,attr"`
	Grn     string   `xml:"ГРН,attr"`
	Date    string   `xml:"ДатаЗаписи,attr"`
}

func (s Svul) ToOrg(docId int, loc string) Org {
	o := Org{}
	o.OKVED = s.OKVED.Osn.Code
	o.DocId = docId
	o.DocLocation = loc
	o.FullName = s.Name.FullName
	o.ShortName = s.Name.ShortName
	o.INN = com.StrTo(s.INN).MustInt64()
	o.OGRN = com.StrTo(s.OGRN).MustInt64()
	for _, v := range s.OKVED.Dop {
		o.OKVEDS = append(o.OKVEDS, v.Code)
	}
	if s.Address.Address.RegionId != "" {
		if s.Address.Address.City.Name != "" {
			o.City = s.Address.Address.City.Name
		} else {
			o.City = s.Address.Address.Region.Name
		}
		o.RegionId = com.StrTo(s.Address.Address.RegionId).MustInt()
	} else if len(s.Addresses) > 0 {
		o.RegionId = com.StrTo(s.Addresses[0].RegionId).MustInt()
		if s.Addresses[0].City.Name != "" {
			o.City = s.Addresses[0].City.Name
		} else {
			o.City = s.Addresses[0].Region.Name
		}
	}

	o.OPF = com.StrTo(s.KodOpf).MustInt()
	o.KPP = com.StrTo(s.KPP).MustInt64()
	return o
}
