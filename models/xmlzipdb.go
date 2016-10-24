package models

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/Unknwon/com"
	"github.com/boltdb/bolt"
	"github.com/clbanning/x2j"
	"github.com/fatih/color"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/ulrf/ulrf/modules/setting"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"html"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	gsl sync.Mutex

	saveChan       = make(chan []Svul, 70)
	jobs     int64 = 0

	dupCache = map[int64]struct{}{}
)

func init() {
	for i := 0; i < runtime.NumCPU(); i++ {
		go ranger(saveChan)
	}
}

func ranger(saveChan chan []Svul) {
	start := time.Now()

	for s := range saveChan {
		color.Yellow("%d", jobs)

		//ogrn := com.StrTo(s.OGRN).MustInt64()

		e := SaveSvuls(s)
		if e != nil {
			color.Red("%s", e)
		}

		color.Cyan("%.2f", (1000.0 / time.Since(start).Seconds()))
		start = time.Now()
		atomic.AddInt64(&jobs, -1)
	}
}

func SaveSvul(s *Svul) (e error) {
	e = SetSvultoBoltDb(s)
	if e != nil {
		return
	}
	dl, id, _ := LookUpLoc(s.OGRN)
	o := s.ToOrg(id, dl)
	var bts []byte
	bts, e = ffjson.Marshal(o)
	if e != nil {
		return
	}
	e = metaBoltDb.Put(i2b(com.StrTo(s.OGRN).MustInt64()), bts)
	if e != nil {
		return
	}
	return
}

func SvulsToOrgs(ss []Svul) (res []Org) {
	for _, v := range ss {
		dl, id, _ := LookUpLoc(v.OGRN)
		o := v.ToOrg(id, dl)
		res = append(res, o)
	}
	return
}

func SaveSvuls(s []Svul) (e error) {
	e = SetSvulstoBoltDb(s)
	if e != nil {
		return e
	}
	e = metaBoltDb.bdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(metaBuckName)
		orgs := SvulsToOrgs(s)
		for _, v := range orgs {
			var bts []byte
			bts, e = ffjson.Marshal(v)
			if e != nil {
				return e
			}
			e = b.Put(i2b(v.OGRN), bts)
			if e != nil {
				return e
			}
		}
		return nil
	})
	if e != nil {
		return e
	}
	return
}

func GetSvul(ogrn string, docLoc string, id int) (s *Svul, e error) {
	start := time.Now()
	//key := fmt.Sprintf("%d", ogrn)
	//if iface, found := cch.Get(key); !found {
	s, e = getSvul(ogrn, docLoc, id)
	printSince(start, "SVUL (i = %d)", id)
	//color.Green("[%s] SVUL", time.Since(start))
	//	go cch.Set(key, s, 5*time.Minute)
	//	return
	//} else {
	//	return iface.(*Svul), nil
	//}
	return
}

/*func getSvul(ogrn string, docLoc string, id int) (s Svul, e error) {
	gsl.Lock()
	defer gsl.Unlock()
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
}*/

/*func RemoveFromLevelDb(key string) error {
	return ldb.Delete([]byte(key), nil)
}*/

/*func GetSvulFromLevelDb(ogrn string) (s Svul, e error) {
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
*/

func GetOneFromZipDb(docLoc string, id int) (s *Svul, e error) {
	var (
	//start = time.Now()
	)
	sv, e := unmarshal(docLoc, id)
	if e != nil {
		return nil, e
	}

	go func(s Svul) {
		e := SaveSvul(&s)
		if e != nil {
			color.Red("%s", e)
		}
	}(sv)
	//color.Green("Unmarshalled %s", time.Since(start))

	return &sv, nil
	//unmarshal
}

func GetFromZipDb(docLoc string, id int) (s *Svul, e error) {
	var (
		start = time.Now()
		ss    []Svul
		//cached = false
	)

	ss, e = UnmarshalAll(docLoc)
	if e != nil {
		return
	}
	//color.Green("Unmarshalled %s", time.Since(start))

	for i, v := range ss {
		if i == id {
			s = &v
			break
		} /*else if !cached && s != nil {
			if _, ok := dupCache[com.StrTo(s.OGRN).MustInt64()]; ok {
				continue
			} else {
				dupCache[com.StrTo(s.OGRN).MustInt64()] = struct{}{}
				saveChan <- v
			}
		}
		*/
	}

	go func(ss []Svul) {
		if len(ss) > 0 {
			if _, ok := dupCache[com.StrTo(ss[0].OGRN).MustInt64()]; !ok {
				dupCache[com.StrTo(ss[0].OGRN).MustInt64()] = struct{}{}
				if jobs < 5 {
					atomic.AddInt64(&jobs, 1)
					saveChan <- ss
				}
			}
		}
	}(ss)

	color.Green("Sended to save chan %s", time.Since(start))

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
	defer rc.Close()
	printSince(start, "reader opened")
	start = time.Now()
	for i, v := range rc.File {
		if v.Name == docLoc {
			printSince(start, "file founded (i = %d)", i)
			start = time.Now()
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
			printSince(start, "file fixed", i)
			e = xm.Close()
			if e != nil {
				return
			}

			return
		}
	}
	return
}

/*func SetSvultoLevelDb(s *Svul) error {
	bts, e := ffjson.Marshal(s)
	if e != nil {
		return e
	}
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, e = w.Write(bts)
	if e != nil {
		if w != nil {
			w.Close()
		}
		return e
	}
	w.Close()
	key := i2b(com.StrTo(s.OGRN).MustInt64())
	return ldb.Put(key, b.Bytes(), nil)
}*/

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

		bts []byte
	)

	if dl == "" {
		e = fmt.Errorf("%s empty doc with id %d", dl, id)
		return
	}
	rc, e = zip.OpenReader(dname + "/" + dl)
	if e != nil {
		return
	}
	defer rc.Close()
	for _, v := range rc.File {
		if v.Name == docLoc {
			if since := time.Since(start); since > time.Second {
				color.Yellow("[file lookuped in zip] %s", since)
			}
			start = time.Now()
			in, e = v.Open()
			if e != nil {
				return
			}
			bts, e = getBytes(in, id)
			if e != nil {
				return
			}
			e = xml.Unmarshal(bts, &s)
			return
		}
	}
	return
}

func getBytes(rdr io.Reader, id int) (res []byte, e error) {

	var (
		i = 0
	)

	res = []byte{'<'}

	//tmp

	tr := transform.NewReader(rdr, charmap.Windows1251.NewDecoder())
	br := bufio.NewReader(tr)
	for bts, e := br.ReadBytes('<'); e == nil; bts, e = br.ReadBytes('<') {
		needTag, _ := br.ReadBytes('>')
		if strings.HasPrefix(strings.TrimSpace(string(needTag)), "СвЮЛ") {
			if i != id {
				i++
				continue
			}

			res = append(res, needTag...)

			for bts, e = br.ReadBytes('>'); e == nil; bts, e = br.ReadBytes('>') {
				res = append(res, bts...)
				if strings.HasSuffix(string(bts), "</СвЮЛ>") {
					return res, nil
				}
			}
			return nil, nil
		}
	}
	return nil, nil
}

func GetTitle(ogrn int64) (string, error) {
	key := "_t" + fmt.Sprint(ogrn)
	if iface, ok := cch.Get(key); ok {
		return string(iface.([]byte)), nil
	}
	var (
		dname = setting.XMLDBZIP.Path
		start = time.Now()
	)
	docloc, _, _ := LookUpLoc(fmt.Sprint(ogrn))
	dl := getXmlLoc(docloc)

	rc, e := zip.OpenReader(dname + "/" + dl)
	if e != nil {
		return "", e
	}
	defer rc.Close()
	printNeedSince(start, 200*time.Millisecond, "Get title zip reader opened")
	start = time.Now()
	for _, v := range rc.File {
		if v.Name == docloc {
			printNeedSince(start, 50*time.Millisecond, "Range docloc found")
			start = time.Now()
			in, e := v.Open()
			if e != nil {
				return "", e
			}
			o, e := searchAttr(in, []byte(fmt.Sprint(ogrn)), []byte{209, 226, 205, 224, 232, 236, 222, 203},
				[]byte{205, 224, 232, 236, 222, 203, 207, 238, 235, 237})
			printNeedSince(start, 150*time.Millisecond, "Attribute found")
			return html.UnescapeString(decode(o)), e
		}
	}
	return "", nil
}

func decode(in []byte) string {
	r := bytes.NewReader(in)
	tr := transform.NewReader(r, charmap.Windows1251.NewDecoder())
	b, _ := ioutil.ReadAll(tr)
	return string(b)
}

func searchAttr(in io.Reader, ogrn []byte, tag, attr []byte) (res []byte, e error) {

	var (
		curOgrn []byte
	)

	br := bufio.NewReader(in)
	for _, e := br.ReadBytes('<'); e == nil; _, e = br.ReadBytes('<') {
		tagVal, e := br.ReadBytes('>')
		if e != nil {
			if e != io.EOF {
				return nil, e
			}
			break
		}

		if tagVal[0] == 209 && tagVal[1] == 226 &&
			tagVal[2] == 222 && tagVal[3] == 203 {

			indx := bytes.Index(tagVal, []byte{206, 195, 208, 205})
			curOgrn = tagVal[indx+6 : indx+13+6]

		}

		val, e := getAttr(tagVal, tag, attr)
		if e != nil {
			color.Red("%s", e)
		}

		if val != nil {
			key := "_t" + string(curOgrn)
			cch.Set(key, []byte(html.UnescapeString(decode(val))), time.Hour)
		}

		if 0 == bytes.Compare(ogrn, curOgrn) && val != nil {
			return val, nil
		}

	}
	return
}

func getAttr(tag []byte, needTagName []byte, need []byte) ([]byte, error) {
	r := bytes.NewReader(tag)
	sk := bufio.NewReader(r)

	//skip tagname
	tagName, e := sk.ReadBytes(' ')
	if e != nil {
		if e == io.EOF {
			return nil, nil
		}
		return nil, e
	}
	if len(tagName) > 0 && tagName[0] == '/' {
		return nil, nil
	}

	if 0 != bytes.Compare(needTagName, tagName[:len(tagName)-1]) {
		return nil, nil
	}

	for attrName, e := sk.ReadBytes('='); e == nil; attrName, e = sk.ReadBytes('=') {
		if 0 != bytes.Compare(attrName[:len(attrName)-1], need) {
			_, e = sk.ReadBytes('"')
			if e != nil && e != io.EOF {
				return nil, e
			}
			_, e = sk.ReadBytes('"')
			if e != nil && e != io.EOF {
				return nil, e
			}
			_, e = sk.ReadBytes(' ')
			if e != nil && e != io.EOF {
				return nil, e
			}
		} else {

			_, e := sk.ReadBytes('"')
			if e != nil && e != io.EOF {
				return nil, e
			}
			tval, e := sk.ReadBytes('"')
			if e != nil && e != io.EOF {
				return nil, e
			}

			res := tval[:len(tval)-2]
			return res, nil
		}
	}
	return nil, nil
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
	defer rc.Close()
	printSince(start, "reader opened")
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
	//color.Green("fiil %s", dname)
	var e error

	w := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			color.Red("%s", err)
			return err
		}

		//color.Green("%s", path)

		if strings.HasSuffix(info.Name(), ".zip") {

			var rc *zip.ReadCloser
			rc, e = zip.OpenReader(path)
			if e != nil {
				return e
			}
			for _, v := range rc.File {
				//if v.Name == "EGRUL_FULL_2015-08-29_30515.XML" {
				//color.Green("%s %s", v.Name, info.Name())

				//}
				xmlLocCache[v.Name] = strings.TrimPrefix(path, dname)
			}
			rc.Close()
		}
		return nil
	}
	e = filepath.Walk(dname, w)
	if e != nil {
		color.Red("%s", e)
	}

	xmlLocCacheInited = true
}

var (
	zMu sync.Mutex
)

func getXmlLoc(name string) string {
	//color.Green("GET %s", name)
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
	//start := time.Now()

	bts, e := lookdb.Get([]byte(ogrn), nil)
	if e != nil {
		color.Red("%s", e)
	}

	arr := strings.Split(string(bts), " ")
	if len(arr) != 2 {
		e = fmt.Errorf("Len arr not 2, %d (%s)", len(arr), bts)
		return
	}
	docLoc = arr[1]
	id = com.StrTo(arr[0]).MustInt()

	//color.Green("[%s LOOKUPED] %s %s", time.Since(start), ogrn, docLoc)
	return
}
