package models

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/Unknwon/com"
	"github.com/clbanning/x2j"
	"github.com/fatih/color"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/ulrf/ulrf/modules/setting"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func GetSvul(ogrn string, docLoc string, id int) (s Svul, e error) {
	key := fmt.Sprintf("%d", ogrn)
	if iface, found := cch.Get(key); !found {
		s, e = getSvul(ogrn, docLoc, id)
		go cch.Set(key, s, time.Minute)
		return
	} else {
		return iface.(Svul), nil
	}
}

func getSvul(ogrn string, docLoc string, id int) (s Svul, e error) {
	s, e = GetSvulFromLevelDb(ogrn)
	if e == nil {
		color.Green("found with ogrn")
		return s, nil
	}
	s, e = GetSvulFromLevelDb(docLoc + fmt.Sprint(id))
	if e == nil {
		go SetSvultoLevelDb(&s)
		color.Green("found with docloc")
		return s, nil
	}
	if docLoc == "" {
		docLoc, id, e = LookUpLoc(ogrn)
		if e != nil {
			return s, e
		}
	}
	s, e = GetFromZipDb(docLoc, id)
	if e != nil {
		color.Red("%s", e)
		return s, nil
	}
	color.Green("found in xmlzip")

	e = RemoveFromLevelDb(docLoc + fmt.Sprint(id))
	if e != nil {
		color.Red("%s", e)
		return s, nil
	}
	e = SetSvultoLevelDb(&s)
	if e != nil {
		color.Red("%s", e)
		return s, nil
	}
	return s, e
}

func RemoveFromLevelDb(key string) error {
	return ldb.Delete([]byte(key), nil)
}

func GetSvulFromLevelDb(ogrn string) (s Svul, e error) {
	var (
		res Svul
		bts []byte
	)
	bts, e = ldb.Get([]byte(ogrn), nil)
	if e != nil {
		return
	}
	var (
		b = bytes.NewReader(bts)
		r *gzip.Reader
	)
	r, e = gzip.NewReader(b)
	if e != nil {
		return
	}
	dec := ffjson.NewDecoder()
	e = dec.DecodeReader(r, &res)
	if e != nil {
		return
	}
	s = res
	return
}

var (
	zMu sync.Mutex
)

func GetFromZipDb(docLoc string, id int) (s Svul, e error) {
	var (
		start = time.Now()
		ss    []Svul
	)

	ss, e = UnmarshalAll(docLoc)
	if e != nil {
		return
	}
	for i, v := range ss {
		if i == id {
			s = v
		}
		e = SetSvultoLevelDb(&v)
		if e != nil {
			color.Red("%s", e)
		}
	}
	color.Green("[xml unmarshaled] %s", time.Since(start))

	return
}

func getRaw(in []byte, id int) (bts []byte, e error) {

	vals, e := x2j.ValuesFromTagPath(string(in), "EGRUL.СвЮЛ")
	fmt.Println(len(vals))
	bts, e = json.MarshalIndent(vals[id], " ", " ")
	return
}

func Dump(docLoc string, id int) (bts []byte, e error) {
	key := fmt.Sprintf("d%s", docLoc)
	if iface, found := cch.Get(key); !found {
		bts, e = dump(docLoc, id)
		go cch.Set(key, bts, 6*time.Hour)
		return
	} else {
		return iface.([]byte), nil
	}
}

func dump(docLoc string, id int) (bts []byte, e error) {
	bts, e = getFixed(docLoc, id)
	if e != nil {
		return
	}
	return getRaw(bts, id)
}

func reader(docLoc string, id int) (io.Reader, error) {
	return nil, nil
}

func getFixed(docLoc string, id int) (bts []byte, e error) {
	color.Yellow("%s %d", docLoc, id)
	start := time.Now()
	dname := setting.XMLDBZIP.Path

	var rc *zip.ReadCloser
	dl := getXmlLoc(docLoc)
	if dl == "" {
		e = fmt.Errorf("%s empty doc with id %d", dl, id)
		return
	}
	rc, e = zip.OpenReader(dname + "/" + dl)
	if e != nil {
		return
	}
	color.Green("[reader opened] %s", time.Since(start))
	for _, v := range rc.File {
		if v.Name == docLoc {
			color.Green("[file founded] %s", time.Since(start))
			var (
				xm io.ReadCloser
			)
			xm, e = v.Open()
			if e != nil {
				return
			}
			bts, e = Fixed(xm)
			if e != nil {
				return
			}
			color.Green("[file fixed] %s", time.Since(start))
			e = xm.Close()
			if e != nil {
				return
			}

			return
		}
	}
	rc.Close()
	return
}

func SetSvultoLevelDb(s *Svul) error {
	bts, e := ffjson.Marshal(s)
	if e != nil {
		return e
	}
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, e = w.Write(bts)
	if e != nil {
		return e
	}
	w.Close()
	return ldb.Put([]byte(s.OGRN), b.Bytes(), nil)
}

func Fixed(in io.Reader) ([]byte, error) {
	//r := bytes.NewReader(in)
	tr := transform.NewReader(in, charmap.Windows1251.NewDecoder())
	buf, e := ioutil.ReadAll(tr)
	if e != e {
		return nil, e
	}
	buf = bytes.Replace(buf, []byte("windows-1251"), []byte("utf-8"), 1)
	return buf, e
}

func charsetReader(_ string, r io.Reader) (io.Reader, error) {
	return transform.NewReader(r, charmap.Windows1251.NewDecoder()), nil
}

func unmarshal(docLoc string, id int) (s Svul, e error) {

	var (
		start = time.Now()
		dname = setting.XMLDBZIP.Path
		in    io.Reader
		rc    *zip.ReadCloser
		dl    = getXmlLoc(docLoc)
	)

	if dl == "" {
		e = fmt.Errorf("%s empty doc with id %d", dl, id)
		return
	}
	rc, e = zip.OpenReader(dname + "/" + dl)
	if e != nil {
		return
	}
	color.Green("[reader opened] %s", time.Since(start))
	for _, v := range rc.File {
		if v.Name == docLoc {
			in, e = v.Open()
			if e != nil {
				return
			}
			dec := xml.NewDecoder(in)
			dec.CharsetReader = charsetReader
			i := 0
			for {
				t, _ := dec.Token()
				if t == nil {
					break
				}
				switch se := t.(type) {
				case xml.StartElement:
					if se.Name.Local == "СвЮЛ" {
						if i != id {
							i++
							continue
						}

						e = dec.DecodeElement(&s, &se)
						return
					}
				}
			}
			return
		}
	}
	return
}

func UnmarshalAll(docLoc string) (s []Svul, e error) {

	var (
		start = time.Now()
		dname = setting.XMLDBZIP.Path
		in    io.Reader
		rc    *zip.ReadCloser
		dl    = getXmlLoc(docLoc)
	)

	if dl == "" {
		e = fmt.Errorf("%s empty doc with id %d", dl)
		return
	}
	rc, e = zip.OpenReader(dname + "/" + dl)
	if e != nil {
		return
	}
	color.Green("[reader opened] %s", time.Since(start))
	for _, v := range rc.File {
		if v.Name == docLoc {
			in, e = v.Open()
			if e != nil {
				return
			}
			dec := xml.NewDecoder(in)
			dec.CharsetReader = charsetReader
			for {
				sv := Svul{}
				t, _ := dec.Token()
				if t == nil {
					break
				}
				switch se := t.(type) {
				case xml.StartElement:
					if se.Name.Local == "СвЮЛ" {
						e = dec.DecodeElement(&sv, &se)
						s = append(s, sv)
						continue
					}
				}
			}
		}
	}
	return
}

var (
	xmlLocCacheInited = false
	xmlLocCache       = make(map[string]string)
)

func fillCacheNames() {
	dname := setting.XMLDBZIP.Path
	var e error

	w := func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".zip") {
			var rc *zip.ReadCloser
			rc, e = zip.OpenReader(path)
			defer rc.Close()
			if e != nil {
				return e
			}
			for _, v := range rc.File {
				color.Green("%s %s", v.Name, info.Name())
				xmlLocCache[v.Name] = strings.TrimPrefix(path, dname)
			}
		}
		return nil
	}
	e = filepath.Walk(dname, w)
	if e != nil {
		color.Red("%s", e)
	}

	xmlLocCacheInited = true
}

func getXmlLoc(name string) string {
	zMu.Lock()
	defer zMu.Unlock()
	if !xmlLocCacheInited {
		fillCacheNames()
	}
	if fname, ok := xmlLocCache[name]; ok {
		return fname
	}
	return ""
}

var (
	lookdb *leveldb.DB
)

func initlDb(path string) (e error) {
	if path == "" {
		path = "data/opa"
	}
	// level
	o := &opt.Options{}
	o.Compression = opt.NoCompression
	o.BlockSize = opt.KiB * 32
	o.WriteBuffer = 64 * opt.KiB
	o.BlockCacheCapacity = 64 * opt.MiB
	lookdb, e = leveldb.OpenFile(path, o)
	if e != nil {
		return
	}
	return
}

func init() {
	e := initlDb("")
	if e != nil {
		panic(e)
	}
}

func LookUpLoc(ogrn string) (docLoc string, id int, e error) {

	bts, e := lookdb.Get([]byte(ogrn), nil)
	if e != nil {
		color.Green("%s", e)
	}

	arr := strings.Split(string(bts), " ")
	if len(arr) != 2 {
		e = fmt.Errorf("Len arr not 2, %d (%s)", len(arr), bts)
		return
	}
	docLoc = arr[1]
	id = com.StrTo(arr[0]).MustInt()

	color.Green("LOOKUPED %s %s", ogrn, docLoc)
	return
}
