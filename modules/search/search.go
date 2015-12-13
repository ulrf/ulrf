package search

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/store/boltdb"
	"github.com/blevesearch/blevex/lang/ru"
	"github.com/fatih/color"
	"github.com/ungerik/go-dry"
)

var (
	indexTitleName = "title.search"
	indexCityName  = "city.search"
	indexEkvdName  = "okved.search"
	indexes        = map[string]bleve.Index{
		indexTitleName: nil,
		indexCityName:  nil,
		indexEkvdName:  nil,
	}
)

func NewContext() {
	var (
		e error
	)
	for key := range indexes {
		fp := key
		var cnf = map[string]interface{}{
			"nosync": true,
		}
		if dry.FileExists(fp) {
			indexes[key], e = bleve.OpenUsing(fp, cnf)
			if e != nil {
				panic(e)
			}
		} else {
			im := bleve.NewIndexMapping()
			im.DefaultAnalyzer = ru.AnalyzerName

			indexes[key], e = bleve.NewUsing(fp, im, bleve.Config.DefaultIndexType,
				boltdb.Name, cnf) // bleve.New(fp, im)
			if e != nil {
				panic(e)
			}
		}

		dc, e := indexes[key].DocCount()
		if e != nil {
			color.Red("%s", e)
		}
		color.Green("%s doc count %d", key, dc)
	}
}

// todo remove it
func SetIndexes(i map[string]bleve.Index) {
	indexes = i
}

func SearchTitle(q string, page int) ([]string, int, error) {
	return search(q, page, indexTitleName)
}

func SearchCity(q string, page int) ([]string, int, error) {
	return search(q, page, indexCityName)
}

func SearchOkved(q string, page int) ([]string, int, error) {
	return search(q, page, indexEkvdName)
}

func search(q string, page int, t string) (ids []string, total int, e error) {
	color.Cyan("Search %s: %s", t, q)
	query := bleve.NewQueryStringQuery(q)
	search := bleve.NewSearchRequest(query)

	var (
		ix  bleve.Index
		has bool
	)

	if ix, has = indexes[t]; !has {
		e = fmt.Errorf("index %s not exists", t)
		return
	}

	/*ix.Index("o", struct{ Title string }{"ооо аквалайн"})
	r, _ := ix.Search(bleve.NewSearchRequest(bleve.NewQueryStringQuery("аквалайн")))
	color.Red("%v %s", r.Hits, r.Total)*/

	search.From = (page - 1) * 10
	search.Size = 10
	var searchResults *bleve.SearchResult
	searchResults, e = ix.Search(search)
	if e != nil {
		return
	}
	color.Green("Search %s found %d", q, searchResults.Total)
	total = int(searchResults.Total)

	for _, v := range searchResults.Hits {
		fmt.Println(v.ID)
		ids = append(ids, v.ID)
	}
	if len(ids) == 0 {
		e = fmt.Errorf("empty result")
		return
	}
	return
}
