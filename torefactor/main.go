package torefactor

import (
	"fmt"
	"github.com/Unknwon/com"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/go-xorm/xorm"
	"github.com/sisteamnik/sitemap"
	"github.com/ulrf/ulrf/models"
	"github.com/ulrf/ulrf/modules/setting"
	"gopkg.in/macaron.v1"
	"html/template"
	"strings"
	"sync"
	"time"
)

func RunMacaron(ctx *cli.Context) {
	setting.NewContext(ctx.String("mode"))
	L.Trace("Run with mode %s on host %s.", ctx.String("mode"), setting.Domain)
	initDB(ctx.String("mode"))
	m := macaron.New()
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Layout: "layout",
		Funcs:  []template.FuncMap{{"since": func(t time.Time) time.Duration { return time.Since(t) }}},
	}))

	m.Use(macaron.Recovery())
	m.Use(macaron.Static("static"))
	m.Use(statMid)

	m.Get("/", func(ctx *macaron.Context) {
		ctx.Data["okveds"] = OKVEDAPI
		ctx.HTML(200, "index")
	})

	m.Get("/search", func(ctx *macaron.Context) {
		var (
			orgs []models.Org
			e    error
			q    = ctx.Query("q")
			page = com.StrTo(ctx.Query("p")).MustInt()
		)
		//q := strings.ToUpper(ctx.Query("q"))

		if page < 1 {
			page = 1
		}

		q = q

		var total int
		if ctx.Query("city") != "" {
			orgs, total, e = searchAny(ctx.Query("city"), page, indexCityName)
			if e != nil {
				color.Red("%s", e)
			}
			//c := strings.ToUpper(ctx.Query("city"))
			//e = eng.Where("city = ?", c).Limit(10, (page-1)*10).Find(&orgs)
			//if e != nil {
			//	color.Red("%s", e)
			//}
		} else if ctx.Query("okved") != "" {
			//if !setting.BleveEnabled {
			//e = eng.Where("okved = ?", strings.ToUpper(ctx.Query("okved"))).Limit(10, (page-1)*10).Find(&orgs)
			//if e != nil {
			//	color.Red("%s", e)
			//}
			//} else {
			orgs, total, e = searchAny(ctx.Query("okved"), page, indexEkvdName)
			if e != nil {
				color.Red("%s", e)
			}
			//}
		} else {
			//if !setting.BleveEnabled {
			//e = eng.OrderBy("id").Where("full_name like ?", "%"+strings.ToUpper(ctx.Query("q"))+"%").Limit(10, (page-1)*10).Find(&orgs)
			//if e != nil {
			//	color.Red("%s", e)
			//}
			//} else {
			orgs, total, e = searchAny(q, page, indexTitleName)
			if e != nil {
				color.Red("%s", e)
			}
			//}
		}

		ctx.Data["Title"] = "Поиск: " + q
		ctx.Data["nextPage"] = page + 1
		if page*10 > total && total != 0 {
			ctx.Data["nextPage"] = 0
		}
		ctx.Data["prevPage"] = page - 1
		ctx.Data["currPage"] = page
		ctx.Data["total"] = total
		ctx.Data["searchQuery"] = strings.ToLower(q)
		ctx.Data["searchCity"] = ctx.Query("city")
		ctx.Data["searchOkved"] = ctx.Query("okved")
		ctx.Data["pagination"] = makePagination(int(total), page)
		ctx.Data["orgs"] = orgs
		ctx.HTML(200, "list")
	})

	m.Get("/regions", func(ctx *macaron.Context) {
		ctx.Data["regions"] = cities
		ctx.Data["Title"] = "Регионы"
		ctx.HTML(200, "regions")
	})
	m.Get("/okveds", func(ctx *macaron.Context) {
		ctx.Data["okveds"] = OKVEDAPI
		ctx.Data["Title"] = "ОКВЭД"
		ctx.HTML(200, "okveds")
	})

	m.Get("/okveds/:cat", func(ctx *macaron.Context) {
		ctx.Data["okveds"] = OKVEDAPI
		ctx.Data["okvedsCat"] = com.StrTo(ctx.Params(":cat")).MustInt()
		ctx.Data["Title"] = "ОКВЭД: Категории"
		ctx.HTML(200, "okvedscat")
	})

	m.Get("/stat", func(c *macaron.Context) {
		c.Data["stat"] = struct {
			DocCount        int64
			IndexSpeed      float64
			IndexTitleRatio float64
			Remaining       time.Duration
		}{
			DocumentsCount,
			StatIndexSpeed,
			indexTitleRatio * 100,
			time.Second * time.Duration((100.0-indexTitleRatio)*StatIndexSpeed),
		}
		c.HTML(200, "stat")
	})

	m.Get("/:id", func(ctx *macaron.Context) {
		var (
			id = ctx.ParamsInt64(":id")
		)

		o, e := models.GetOrg(id)
		if e != nil {
			color.Red("%s", e)
		}

		s, e := models.GetSvul(o.OGRN, o.DocLocation, o.DocId)
		if e != nil {
			color.Red("%s", e)
		}

		var orgs []models.Org
		e = eng.Where("id > ?", ctx.ParamsInt(":id")).Limit(3).Find(&orgs)
		if e != nil {
			color.Red("%s", e)
		}

		if len(orgs) != 3 {
			orgs = orgs[:0]
			e = eng.Where("id < ?", com.StrTo(ctx.Params(":id")).MustInt()).Limit(3).Find(&orgs)
			if e != nil {
				color.Red("%s", e)
			}
		}

		// todo tmp
		//search.IndexCity(o.OGRN, strings.ToLower(o.City))

		ctx.Data["okveds"] = OKVEDAPI
		ctx.Data["org"] = o
		ctx.Data["orgs"] = orgs
		ctx.Data["svul"] = s
		ctx.HTML(200, "get")
	})

	m.Get("/map", func(ctx *macaron.Context) {
		page := com.StrTo(ctx.Query("p")).MustInt()
		if page == 0 {
			page = 1
		}

		var orgs []models.Org
		sess := eng.OrderBy("id")
		if page > 100 {
			sess.Where("id > ?", (page-1)*10).Limit(10)
		} else {
			sess.Limit(10, (page-1)*10)
		}
		e := sess.Find(&orgs)
		if e != nil {
			color.Red("%s", e)
		}

		ctx.Data["nextPage"] = page + 1
		ctx.Data["prevPage"] = page - 1
		ctx.Data["lastPage"] = DocumentsCount / 10
		ctx.Data["currPage"] = page
		ctx.Data["orgs"] = orgs
		ctx.Data["DocumentsCount"] = DocumentsCount
		ctx.Data["pagination"] = makePagination(int(DocumentsCount), page)
		ctx.HTML(200, "map")
	})

	numInSitemap := 50
	for i := 1; i < int(DocumentsCount); i++ {
		m.Get("/sitemap."+fmt.Sprint(i)+".xml", func(c *macaron.Context) {
			color.Yellow("%d", i)
			c.Resp.Header().Set("Content-Type", "text/xml; charset=utf-8")
			c.WriteHeader(200)
			var i int
			arr := strings.Split(c.Req.RequestURI, ".")
			if len(arr) != 3 {
				return
			}
			i = com.StrTo(arr[1]).MustInt()

			wr, e := sitemap.NewWriter(c.Resp)
			if e != nil {
				color.Red("%s", e)
			}

			var orgs []models.Org
			e = eng.Cols("id").OrderBy("id").Where("id > ?", i).Limit(numInSitemap, i).Find(&orgs)
			if e != nil {
				color.Red("%s", e)
			}

			var it sitemap.Item
			for _, v := range orgs {
				it.Loc = "http://" + setting.Domain + "/" + fmt.Sprint(v.Id)
				it.ChangeFreq = "weekly"
				it.LastMod = time.Now()
				it.Priority = 0.8
				e = wr.Put(it)
				if e != nil {
					color.Red("%s", e)
				}
			}
			e = wr.Release()
			if e != nil {
				color.Red("%s", e)
			}

		})
		i += numInSitemap - 1
	}
	m.Get("/sitemap.xml", func(c *macaron.Context) {
		c.Resp.Header().Set("Content-Type", "text/xml; charset=utf-8")
		c.WriteHeader(200)

		wr, e := sitemap.NewIndexWriter(c.Resp)
		if e != nil {
			color.Red("%s", e)
		}

		for i := 1; i < int(DocumentsCount); i++ {
			e = wr.Put(sitemap.NewIndexItem("http://"+setting.Domain+"/sitemap."+fmt.Sprint(i)+".xml",
				time.Now()))
			if e != nil {
				color.Red("%s", e)
			}
			i += numInSitemap - 1
		}
		e = wr.Release()
		if e != nil {
			color.Red("%s", e)
		}
	})

	m.Run()
}

var (
	StatLongest  time.Duration
	StatCount    int64
	StatDuration time.Duration
	StatMu       sync.Mutex
)

func statMid(ctx *macaron.Context) {
	s := time.Now()
	ctx.Data["StatLongest"] = StatLongest
	ctx.Data["StatCount"] = StatCount
	ctx.Data["StatAverage"] = fmt.Sprintf("%.3f", (float64(StatDuration.Seconds()) / float64(StatCount)))
	ctx.Data["StatStart"] = s
	ctx.Data["StatSince"] = func(t time.Time) time.Duration {
		return time.Since(t)
	}

	/*
		other middlewares
	*/
	ctx.Data["url"] = ctx.Req.URL
	ctx.Data["domain"] = setting.Domain

	ctx.Next()
	dur := time.Since(s)
	fmt.Println(dur)
	StatMu.Lock()
	if StatLongest < dur {
		StatLongest = dur
	}
	StatMu.Unlock()
	StatCount++
	StatDuration += dur
}

func makePagination(ctn, cur int) []int {
	var (
		start = cur - 5
		l     = 10
		end   = start + l
		res   []int
	)
	if ctn == 0 || cur == 0 {
		return res
	}
	if ctn < 10 {
		ctn = 10
	}

	if start < 0 {
		end += -start
		start = 0
	}
	if end*10 > ctn {
		end = ctn / 10
	}

	for i := start + 1; i < end; i++ {
		res = append(res, i)
	}

	color.Yellow("%d %d, %v", start, end, res)
	return res
}

func getOrg(eng *xorm.Engine, id int64) (*models.Org, error) {
	o := new(models.Org)
	o.Id = id
	_, e := eng.Get(o)
	return nil, e
}

func GetOrg(id int64) (*models.Org, error) {
	return getOrg(eng, id)
}
