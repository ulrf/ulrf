package setting

import (
	"gopkg.in/ini.v1"
)

var (
	RunMode string

	Domain   string
	LogLevel string

	Db struct {
		User     string
		Pass     string
		Database string
	}

	XMLDBZIP struct {
		Path string
	}

	BleveEnabled bool
)

func NewContext(mode string) {
	f, e := ini.Load("conf/app.ini")
	if e != nil {
		panic(e)
	}

	RunMode = mode

	Domain = f.Section("app").Key("domain").String()

	sec := f.Section(mode)
	Db.User = sec.Key("db.user").String()
	Db.Pass = sec.Key("db.pass").String()
	Db.Database = sec.Key("db.name").String()
	XMLDBZIP.Path = sec.Key("xmldb.path").String()

	BleveEnabled = sec.Key("bleve.enabled").MustBool(false)

}
