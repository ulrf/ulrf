package models

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"github.com/fatih/color"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/ulrf/ulrf/modules/setting"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func GetSvul(ogrn string, docLoc string, id int) (s Svul, e error) {
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

func GetFromZipDb(docLoc string, id int) (s Svul, e error) {
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
				xm  io.ReadCloser
				bts []byte
			)
			xm, e = v.Open()
			if e != nil {
				return
			}
			bts, e = fixed(xm)
			if e != nil {
				return
			}
			color.Green("[file fixed] %s", time.Since(start))
			e = xm.Close()
			if e != nil {
				return
			}

			s, e = unmarshal(bts, id)
			if e != nil {
				return
			}
			color.Green("[xml unmarshaled] %s", time.Since(start))
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

func fixed(in io.Reader) ([]byte, error) {
	//r := bytes.NewReader(in)
	tr := transform.NewReader(in, charmap.Windows1251.NewDecoder())
	buf, e := ioutil.ReadAll(tr)
	if e != e {
		return nil, e
	}
	buf = bytes.Replace(buf, []byte("windows-1251"), []byte("utf-8"), 1)
	return buf, e
}

func unmarshal(in []byte, id int) (s Svul, e error) {
	f := bytes.NewReader(in)
	dec := xml.NewDecoder(f)
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

var (
	xmlLocCacheInited = false
	xmlLocCache       = map[string]string{}
)

func fillCacheNames() {
	dname := setting.XMLDBZIP.Path
	var e error

	w := func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".zip") {
			var rc *zip.ReadCloser
			rc, e = zip.OpenReader(path)
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
		panic(e)
	}

	xmlLocCacheInited = true
}

func getXmlLoc(name string) string {
	if !xmlLocCacheInited {
		fillCacheNames()
	}
	if fname, ok := xmlLocCache[name]; ok {
		return fname
	}
	return ""
}
