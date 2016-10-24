package models

import (
	"bufio"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fatih/color"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/ungerik/go-dry"
	"os"
	"time"
)

func StartCrawler() {
	go func() {
		ht()
	}()
}

func ht() {
	skip := 2
	f, e := os.Open("range.test")
	if e != nil {
		panic(e)
	}
	br := bufio.NewReader(f)
	for line, _, e := br.ReadLine(); e == nil; line, _, e = br.ReadLine() {
		if skip > 0 {
			skip--
			continue
		}
		//_ = line
		if len(line) == 8 {
			start := time.Now()
			u := "http://localhost:4000/" + fmt.Sprintf("%d", b2i(line))
			_, _ = dry.FileGetBytes(u)
			since := time.Since(start)
			if since > time.Second {
				color.Red("%d %s", b2i(line), time.Since(start))
			}
			//time.Sleep(10 * time.Second)
		}

	}
}

func frommeta() {
	color.Green("ALE")
	time.Sleep(2 * time.Second)
	metaBoltDb.bdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(metaBuckName)
		c := b.Cursor()
		color.Green("Start RANGE")
		i := 0
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			color.Green("write %d", i)
			i++
			var (
				bts []byte
			)
			bts = append(bts, k...)
			bts = append(bts, '\n')
			dry.FileAppendBytes("range.test", bts)
		}
		return nil
	})
}

func NewLevel(path string) (*leveldb.DB, error) {
	o := &opt.Options{}
	o.BlockCacheCapacity = 128 * opt.MiB
	o.WriteBuffer = 64 * opt.MiB
	o.CompactionTableSize = 16 * opt.MiB
	o.OpenFilesCacheCapacity = 50
	return leveldb.OpenFile(path, o)
}
