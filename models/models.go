package models

import (
	"github.com/fatih/color"
	"github.com/syndtr/goleveldb/leveldb/opt"

	//"github.com/go-xorm/xorm"
	"github.com/syndtr/goleveldb/leveldb"
	//"github.com/ulrf/ulrf/modules/setting"
)

var (
	//x   *xorm.Engine
	ldb *logdb
)

type logdb struct {
	*leveldb.DB
}

func (l *logdb) Get(key []byte, o []byte) ([]byte, error) {
	color.Cyan("[leveldb] GET %s", key)
	return l.DB.Get(key, nil)
}

func (l *logdb) Put(key, value, o []byte) error {
	return l.DB.Put(key, value, nil)
}

func NewEngine() {

}

/*func SetEngine(e *xorm.Engine) {
	x = e
}*/

func SetLevelDb(db *leveldb.DB) {
	l := new(logdb)
	l.DB = db
	ldb = l
}

func NewLevel(path string) (*leveldb.DB, error) {
	o := &opt.Options{}
	o.BlockCacheCapacity = 128 * opt.MiB
	o.WriteBuffer = 64 * opt.MiB
	o.CompactionTableSize = 16 * opt.MiB
	o.OpenFilesCacheCapacity = 500
	return leveldb.OpenFile(path, o)
}
