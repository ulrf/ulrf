package models

import (
	"github.com/go-xorm/xorm"
	"github.com/syndtr/goleveldb/leveldb"
	//"github.com/ulrf/ulrf/modules/setting"
)

var (
	x   *xorm.Engine
	ldb *leveldb.DB
)

func NewEngine() {

}

func SetEngine(e *xorm.Engine) {
	x = e
}

func SetLevelDb(db *leveldb.DB) {
	ldb = db
}
