package torefactor

import (
	"fmt"
	"github.com/blevesearch/bleve"
	_ "github.com/blevesearch/bleve/index/store/goleveldb"
	"github.com/fatih/color"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	"github.com/otium/queue"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/ulrf/ulrf/models"
	"github.com/ulrf/ulrf/modules/search"
	"github.com/ulrf/ulrf/modules/setting"
	"os"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	eng *xorm.Engine

	indexTitleName = "title.search"
	indexCityName  = "city.search"
	indexEkvdName  = "okved.search"
	indexes        = map[string]bleve.Index{
		indexTitleName: nil,
		indexCityName:  nil,
		indexEkvdName:  nil,
	}
	indexTitleRatio float64
	indexCityRatio  float64
	indexEkvdRatio  float64

	ldb *leveldb.DB

	DocumentsCount int64
	StatIndexSpeed float64

	cities []string
)

const (
	Dev  = "dev"
	Prod = "prod"
)

var (
	RunMode = Dev
)

func initDB(mode string) {
	L.Trace("Connect to db %s,%s,%s", setting.Db.User, setting.Db.Pass, setting.Db.Database)
	var e error

	// todo move this to models
	eng, e = xorm.NewEngine("postgres", fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable",
		setting.Db.User, setting.Db.Pass, setting.Db.Database))
	if e != nil {
		panic(e)
	}
	models.SetEngine(eng)

	f, e := os.OpenFile("log/sql.log", os.O_CREATE|os.O_WRONLY, 0666)
	if e != nil {
		panic(e)
	}
	eng.ShowDebug = true
	eng.ShowInfo = true
	eng.ShowSQL = true
	eng.ShowErr = true
	eng.ShowWarn = true
	eng.Logger = xorm.NewSimpleLogger(f)

	e = eng.Sync2(new(models.Org))
	if e != nil {
		L.Error("%s", e)
	}

	//go func() {
	cnt, e := eng.Count(new(models.Org))
	if e != nil {
		L.Error("%s", e)
	}
	DocumentsCount = cnt
	L.Trace("Documents count: %d", DocumentsCount)
	//}()

	var orgs []models.Org
	e = eng.Cols("city").Distinct("city").Find(&orgs)
	if e != nil {
		panic(e)
	}

	uF := func(s string) string {
		if s == "" {
			return ""
		}
		r, n := utf8.DecodeRuneInString(s)
		return string(unicode.ToUpper(r)) + s[n:]
	}

	for _, v := range orgs {
		cities = append(cities, uF(strings.ToLower(v.City)))
	}

	search.NewContext()

	o := &opt.Options{}
	o.Compression = opt.NoCompression
	o.BlockSize = opt.KiB * 32
	o.WriteBuffer = 64 * opt.KiB
	o.BlockCacheCapacity = 64 * opt.MiB
	if mode == Prod {
		o.BlockCacheCapacity = 256 * opt.MiB
	}
	o.OpenFilesCacheCapacity = 100
	ldb, e = leveldb.OpenFile("base.leveldb", o)
	if e != nil {
		panic(e)
	}

	models.SetLevelDb(ldb)

}

func bleveIndex() {
	color.Yellow("Start indexing")
	cnt := DocumentsCount
	bic, e := indexes[indexTitleName].DocCount()
	if e != nil {
		color.Red("%s", e)
		return
	}
	s := time.Now()
	n := 100
	size := 0
	ii := int(bic)

	o := new(models.Org)
	eng.Cols("id").Limit(1, int(bic)).Get(o)

	bathcQ := queue.NewQueue(func(val interface{}) {
		orgs := val.([]models.Org)
		e := indexBatch(orgs)
		if e != nil {
			color.Red("%s", e)
			return
		}
		size += n
		StatIndexSpeed = (float64(size) / time.Since(s).Seconds())
		indexCityRatio = (float64(ii) / float64(DocumentsCount))
		indexTitleRatio = (float64(ii) / float64(DocumentsCount))
		indexEkvdRatio = (float64(ii) / float64(DocumentsCount))
	}, 1)

	q := queue.NewQueue(func(_ interface{}) {
		var orgs []models.Org
		e = eng.Cols("id", "city", "okved", "okveds", "full_name", "short_name").Where("id > ?", o.Id).OrderBy("id").Limit(n).Find(&orgs)
		if e != nil {
			color.Red("%s", e)
			return
		}
		o.Id = getMaxId(orgs)
		bathcQ.Push(orgs)
	}, 1)

	for i := int(bic); i <= int(cnt); i += n {
		q.Push(nil)
	}
	q.Wait()
}

func getMaxId(arr []models.Org) (m int64) {
	for _, v := range arr {
		if v.Id > m {
			m = v.Id
		}
	}
	return
}

func indexBatch(orgs []models.Org) (e error) {
	ch := make(chan struct{}, 3)
	for name, ix := range indexes {
		switch name {
		case indexTitleName:
			go func() {
				defer func() { ch <- struct{}{} }()
				b := ix.NewBatch()
				for _, org := range orgs {
					e := b.Index(fmt.Sprint(org.Id), org.ForIndex())
					if e != nil {
						color.Red("%s", e)
						return
					}
				}
				e = ix.Batch(b)
				if e != nil {
					color.Red("%s", e)
				}
			}()
		case indexCityName:
			go func() {
				defer func() { ch <- struct{}{} }()

				b := ix.NewBatch()
				for _, org := range orgs {
					e := b.Index(fmt.Sprint(org.Id), org.ForCityIndex())
					if e != nil {
						color.Red("%s", e)
						return
					}
				}
				e = ix.Batch(b)
				if e != nil {
					color.Red("%s", e)
					return
				}
			}()
		case indexEkvdName:
			go func() {
				defer func() { ch <- struct{}{} }()

				b := ix.NewBatch()
				for _, org := range orgs {
					e := b.Index(fmt.Sprint(org.Id), org.ForOKVEDIndex())
					if e != nil {
						color.Red("%s", e)
						return
					}
				}
				e = ix.Batch(b)
				if e != nil {
					color.Red("%s", e)
					return
				}
			}()
		}
	}
	for range indexes {
		<-ch
	}
	close(ch)
	return nil
}
