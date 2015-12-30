package torefactor

/*
import (
	_ "github.com/blevesearch/blevex/lang/ru"
	"github.com/fatih/color"
	"github.com/ulrf/ulrf/models"
	"github.com/ulrf/ulrf/modules/search"
	"strings"
)

func searchAny(q string, page int, t string) (orgs []models.Org, total int, e error) {
	color.Cyan("%s %s", q, t)
	var (
		ids []string
	)
	switch t {
	case indexEkvdName:
		ids, total, e = search.SearchOkved(q, page)
		if e != nil {
			color.Red("%s", e)
		}
		break
	case indexCityName:
		ids, total, e = search.SearchCity(q, page)
		if e != nil {
			color.Red("%s", e)
		}

		break
	case indexTitleName:
		idsi, t, e := search.Search(q, page)
		if e != nil {
			color.Red("%s", e)
		}
		ids = search.IntToStringSlice(idsi)
		color.Green("%s", ids)
		total = t
		break
	}

	e = eng.In("ogrn", ids).Find(&orgs)

	return
}

func searchInDb(q string, t string) (orgs []models.Org, total int, e error) {
	switch t {
	case indexEkvdName:
		e = eng.Where("okved = ?", strings.ToUpper(q)).Limit(10).Find(&orgs)
		if e != nil {
			color.Red("%s", e)
		}
		break
	case indexCityName:
		c := strings.ToUpper(q)
		e = eng.Where("city = ?", c).Limit(10).Find(&orgs)
		if e != nil {
			color.Red("%s", e)
		}
		break
	case indexTitleName:
		e = eng.OrderBy("id").Where("full_name like ?", "%"+strings.ToUpper(q)+"%").Limit(10).Find(&orgs)
		if e != nil {
			color.Red("%s", e)
		}
		break
	}
	return
}*/

/*func indexTitleQuery(q string) error {
	var (
		orgs []models.Org
	)

	e := eng.Cols("full_name", "short_name", "ogrn").Where("full_name like ?", "%"+strings.ToUpper(q)+"%").Limit(1000).Find(&orgs)
	if e != nil {
		color.Red("%s", e)
	}

	return search.IndexTitleBatch(orgs)
}

func indexCityQuery(q string) (e error) {
	var (
		n    = 1000
		city = strings.ToLower(q)
	)
	q = strings.ToUpper(q)

	cnt, e := eng.Where("city = ?", q).Count(new(models.Org))
	if e != nil {
		return e
	}
	L.Trace("Index %d orgs(%s)", cnt, city)

	for i := 0; i < int(cnt); i += n {
		var (
			orgs []models.Org
		)
		e = eng.OrderBy("id").Limit(n, i).Cols("ogrn").Where("city = ?", q).Find(&orgs)
		if e != nil {
			color.Red("%s", e)
		}
		e = search.IndexCityBatch(city, orgs)
		if e != nil {
			color.Red("%s", e)
			return e
		}
	}

	return
}

func indexOkvedQuery(q string) (e error) {
	var (
		orgs   []models.Org
		stop   = true
		offset = 0
		n      = 5000
	)
	q = strings.ToUpper(q)

	for stop {
		e = eng.OrderBy("id").Limit(n, offset).Cols("okveds", "ogrn").Where("okved = ?", q).Find(&orgs)
		if e != nil {
			color.Red("%s", e)
		}
		q = strings.ToLower(q)
		e = search.IndexOkvedBatch(q, orgs)
		if e != nil {
			color.Red("%s", e)
			return e
		}
		offset += n
		if len(orgs) < n {
			stop = false
		}
		orgs = orgs[:0]
	}

	return
}

func indexQuery(q string) (e error) {
	q = strings.ToLower(q)
	var (
		orgs []models.Org
	)
	e = eng.Cols("ogrn", "okved", "okveds").Where("okved = ?", strings.ToUpper(q)).Find(&orgs)
	if e != nil {
		color.Red("%s", e)
	}
	for _, v := range orgs {
		e = search.IndexOkved(v.OGRN, v.ForOKVEDIndex())
		if e != nil {
			color.Red("%s", e)
		}
	}
	orgs = orgs[:0]

	e = eng.Cols("ogrn").Where("city = ?", q).Find(&orgs)
	if e != nil {
		color.Red("%s", e)
	}
	for _, v := range orgs {
		e = search.IndexCity(v.OGRN, models.ForCityIndex{City: q})
		if e != nil {
			color.Red("%s", e)
		}
	}
	orgs = orgs[:0]

	e = eng.Cols("full_name", "short_name", "ogrn").Where("full_name like ?", "%"+strings.ToUpper(q)+"%").Limit(1000).Find(&orgs)
	if e != nil {
		color.Red("%s", e)
	}
	for _, v := range orgs {
		e = search.IndexTitle(v.OGRN, v.ForIndex())
		if e != nil {
			color.Red("%s", e)
		}
	}
	orgs = orgs[:0]

	return nil
}
*/
