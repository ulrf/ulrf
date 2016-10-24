package titles

import (
	"archive/zip"
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/Unknwon/com"
	"github.com/boltdb/bolt"
	"github.com/fatih/color"
	"github.com/ulrf/ulrf/modules/setting"
	"github.com/ungerik/go-dry"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"html"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	dname   = "" // "/home/god/e/EGRUL"
	e       error
	gstart  = time.Now()
	targets []string
	ii      = 0

	saveChan = make(chan map[int64]string, 1000)

	cache   = map[int64][]byte{}
	cacheMu sync.Mutex
)

func Index() {
	InitDb()
	dname = setting.XMLDBZIP.Path
	//view()
	w := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			//color.Red("%s | %s", err, path)
			return nil
		}
		//color.Green("%s", path)

		if strings.HasSuffix(info.Name(), ".zip") {
			targets = append(targets, strings.TrimPrefix(path, dname))
		}
		return nil
	}
	//color.White("Start walk %s", dname)
	e = filepath.Walk(dname, w)
	if e != nil {
		color.Red("%s, %s", e, dname)
	}
	color.Green("Files: %d", len(targets))

	needIstr, _ := dry.FileGetString("data/titlescounter")
	needI := com.StrTo(needIstr).MustInt()

	color.Green("Skip %d", needI)

	nameChan := make(chan string)

	for i := 0; i < runtime.NumCPU(); i++ {
		go parse(nameChan)
	}

	go save(saveChan)

	for i := len(targets) - 1; i > 0; i-- {
		dry.FileSetString("data/titlescounter", fmt.Sprint(i))
		nameChan <- dname + "/" + targets[i]
	}

	/*for i, v := range targets {
		//if i < needI {
		//	continue
		//}
		dry.FileSetString("data/titlescounter", fmt.Sprint(i))
		nameChan <- dname + "/" + v
	}*/
	wg.Wait()
}

var (
	wg sync.WaitGroup
)

func parse(in chan string) {
	for fname := range in {
		wg.Add(1)
		ii++
		var (
			r *zip.ReadCloser
			e error
		)

		r, e = zip.OpenReader(fname)
		if e != nil {
			color.Red("%s %s", e, fname)
			return
		}

		for _, fl := range r.File {
			//s := time.Now()
			//color.Cyan("%d/%d %s", j, len(r.File), fl.Name)
			rdr, e := fl.Open()
			if e != nil {
				color.Red("%s", e)
				continue
			}

			m, e := titles(rdr)
			if e != nil {
				color.Red("%s", e)
			}
			//color.White("%d", len(m))
			saveChan <- m
			e = rdr.Close()
			if e != nil {
				color.Red("%s", e)
			}

			//color.Cyan("File Parsed %s", time.Since(s))
		}
		since := time.Since(gstart)
		//cur := float64(i) / float64(len(targets))
		total := float64(since) * float64(len(targets)) / float64(ii+1)
		rem := time.Duration(total) - since

		color.Green("Start %d. Remaining: %s", ii, rem)
		wg.Done()
	}
}

func InitDb() {
	var (
		e error
	)
	bdb, e = bolt.Open("data/titla.bolt", 0777, nil)
	if e != nil {
		panic(e)
	}
	e = bdb.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists(buck)
		return nil
	})
	if e != nil {
		panic(e)
	}
	go FillCache()
}

var (
	bdb *bolt.DB

	buck = []byte{'t'}
)

func save(a chan map[int64]string) {
	for in := range a {
		start := time.Now()
		e := bdb.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(buck)
			for k, v := range in {
				e := b.Put(i2b(k), []byte(v))
				if e != nil {
					return e
				}
			}
			return nil
		})
		if e != nil {
			color.Red("%s", e)
		}
		if since := time.Since(start); since > time.Millisecond*450 {
			color.Yellow("File saved %s %d", since, len(in))
		}
	}
}

func i2b(i int64) []byte {
	var bid = make([]byte, 8)
	binary.BigEndian.PutUint64(bid, uint64(i))
	return bid
}

func b2i(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func titles(rdr io.Reader) (res map[int64]string, e error) {

	var (
		needEnd  = "НаимЮЛПолн=\""
		needOgrn = " ОГРН=\""

		curOgrn []byte
	)
	s := time.Now()
	res = map[int64]string{}

	//tmp

	tr := transform.NewReader(rdr, charmap.Windows1251.NewDecoder())
	br := bufio.NewReader(tr)
	for _, e := br.ReadBytes('<'); e == nil; _, e = br.ReadBytes('<') {
		needTag, _ := br.ReadBytes('>')
		if strings.HasPrefix(strings.TrimSpace(string(needTag)), "СвЮЛ") {
			//color.Green("OGRN %s", curOgrn)
			if ix := strings.Index(string(needTag), needOgrn); ix > -1 {
				//endVal := strings.Index(, "\"")
				sta := ix + len([]byte(needOgrn))
				curOgrn = needTag[sta : sta+13]
				//color.Green("OGRN %s", curOgrn)
			}
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(string(needTag)), "СвНаимЮЛ") {
			/*bts, e = br.ReadBytes('>')*/
			if ix := strings.Index(string(needTag), needEnd); ix > -1 {
				//color.Green("%d %d", ix, len(needTag))
				//endVal := strings.Index(, "\"")
				arr := strings.SplitN(string(needTag[ix+len([]byte(needEnd)):]), "\"", 2)
				if len(arr) != 2 {
					color.Green("len arr !=2 %d", len(arr))
				}
				res[com.StrTo(string(curOgrn)).MustInt64()] = arr[0]
				//color.Green("%s %s", curOgrn, arr[0])
			}
		}
	}
	if since := time.Since(s); since > time.Millisecond*1000 {
		color.Yellow("File parsed %s %d", time.Since(s), len(res))
	}

	return
}

func view() {
	bdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(buck)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			color.Green("%d %s", b2i(k), v)
		}
		return nil
	})
}

func GetTitle(ogrn int64) (res string) {
	bdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(buck)
		val := b.Get(i2b(ogrn))
		res = string(val)
		return nil
	})
	return
}

func GetTitles(ogrns []int64) (orgs []string, e error) {
	//start := time.Now()
	bdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(buck)

		for _, ogrn := range ogrns {
			if val, ok := get(ogrn); ok {
				orgs = append(orgs, html.UnescapeString(val))
			} else {
				val := html.UnescapeString(string(b.Get(i2b(ogrn))))
				orgs = append(orgs, val)
				set(ogrn, []byte(val))
			}
		}
		return nil
	})
	return
}

func FillCache() {
	bdb.View(func(tx *bolt.Tx) error {
		var (
			b = tx.Bucket(buck)
			c = b.Cursor()
			i = 0
		)

		for k, v := c.First(); k != nil; k, v = c.Next() {
			set(b2i(k), v)
			i++
			if i%10000 == 0 {
				color.Cyan("Cached %d", i)
			}
		}

		return nil
	})
}

func set(o int64, val []byte) {
	cacheMu.Lock()
	cache[o] = val
	cacheMu.Unlock()
}

func get(o int64) (string, bool) {
	s, ok := cache[o]
	return string(s), ok
}
