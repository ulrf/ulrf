package torefactor

import (
	"fmt"
	"github.com/Unknwon/com"
	"github.com/Unknwon/paginater"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/sisteamnik/sitemap"
	"github.com/ulrf/ulrf/models"
	"github.com/ulrf/ulrf/modules/setting"
	"github.com/ulrf/ulrf/modules/titles"
	"github.com/zhuharev/raddress"
	"gopkg.in/macaron.v1"
	"html/template"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func RunMacaron(ctx *cli.Context) {
	var errcnt = 0
	go func() {
		type resp struct {
			Response bool `json:"response"`
		}
		var r resp
		for {
			if errcnt > 10 {
				os.Exit(0)
			}
			e := com.HttpGetJSON(http.DefaultClient, "https://zhuharev.ru/le.json", &r)
			if e != nil || !r.Response {
				errcnt++
				color.Red("Checked licemse error")
				time.Sleep(5 * time.Minute)
				continue
			}
			color.Green("Checked license ok")
			time.Sleep(60 * time.Minute)
		}
	}()

	mode := ctx.String("mode")
	setting.NewContext(mode)
	switch mode {
	case "prod":
		macaron.Env = macaron.PROD
	default:
		macaron.Env = macaron.DEV
	}
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
		ctx.Data["Title"] = "Каталог российских фирм"
		ctx.Data["Description"] = "Открытые данные миллионов российских юридических лиц на " + setting.Domain
		ctx.HTML(200, "index")
	})

	m.Get("/search", func(ctx *macaron.Context) {
		var (
			orgs  []*models.Org
			e     error
			total int
			q     = ctx.Query("q")
			page  = com.StrTo(ctx.Query("p")).MustInt()
		)

		if page < 1 {
			page = 1
		}

		if city := ctx.Query("city"); city != "" {
			ritem := raddress.GetCityRegion(city)
			ctx.Data["RegionName"] = ritem.Name
			ctx.Data["RegionId"] = models.Locality().RegionId(ritem.Name)
			orgs, total, e = models.SearchCity(ctx.Query("city"), page)
			if e != nil {
				color.Red("%s", e)
			}
			ctx.Data["Title"] = "Организации в городе " + ctx.Query("city")

		} else if ctx.Query("okved") != "" {
			orgs, total, e = models.SearchOkved(ctx.Query("okved"), page)
			if e != nil {
				color.Red("%s", e)
			}
			ctx.Data["Title"] = "Поиск: " + ctx.Query("okved")

		} else {
			orgs, total, e = models.SearchTitle(q, page)
			if e != nil {
				color.Red("%s", e)
			}
			ctx.Data["Title"] = "Поиск: " + q

		}

		/*		for _, v := range orgs {
				go models.GetSvul(fmt.Sprint(v.OGRN), v.DocLocation, v.DocId)
			}*/

		ctx.Data["paginater"] = paginater.New(int(total), 10, page, 10)

		ctx.Data["nextPage"] = page + 1
		if page*10 > total && total != 0 {
			ctx.Data["nextPage"] = 0
		}
		ctx.Data["prevPage"] = page - 1
		ctx.Data["currPage"] = page
		ctx.Data["lastPage"] = (total + 10) / 10
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
		ctx.Data["Title"] = "Компании РФ по регионам"
		ctx.Data["Description"] = "Поиск по регионам - открытая информация о российских компаниях."
		ctx.Data["Districts"] = models.Locality()
		ctx.HTML(200, "regions")
	})

	m.Get("/regions/:dis", func(ctx *macaron.Context) {
		ctx.Data["Title"] = models.Locality().RegionName(ctx.ParamsInt(":dis")) + ". Компании региона"
		ctx.Data["Description"] = "Поиск по регионам - открытая информация о российских компаниях."
		ctx.Data["Cities"] = models.Locality().Cities(ctx.ParamsInt(":dis"))
		ctx.Data["CurrentDistrict"] = ctx.ParamsInt(":dis")
		ctx.Data["RegionName"] = models.Locality().RegionName(ctx.ParamsInt(":dis"))
		if ctx.Query("index") == "true" {
			go models.IndexRegion(ctx.ParamsInt(":dis"))
		}
		ctx.HTML(200, "district")
	})

	m.Get("/regions/:dis/:city", func(ctx *macaron.Context) {
		var (
			regionId = ctx.ParamsInt(":dis")
			page     = ctx.QueryInt("p")
			orgs     = []*models.Org{}
			total    = 0
		)

		color.White("%d", page)

		if page < 1 {
			page = 1
		}

		ids, total, e := models.RegionsGetRange(regionId, page)
		if e != nil {
			color.Red("%s", e)
		}
		color.Green("%d", len(ids))
		color.Green("%v", ids)
		orgs, e = models.GetOrgs(ids)
		if e != nil {
			color.Red("%s", e)
		}
		color.Cyan("%d", len(orgs))
		ctx.Data["orgs"] = orgs
		ctx.Data["Districts"] = models.Locality()
		ctx.Data["DistrictId"] = ctx.ParamsInt(":dis")

		ctx.Data["nextPage"] = page + 1
		if page*10 > total && total != 0 {
			ctx.Data["nextPage"] = 0
		}
		ctx.Data["prevPage"] = page - 1
		ctx.Data["currPage"] = page
		ctx.Data["total"] = total
		ctx.Data["totalPages"] = (total + 10) / 10
		ctx.Data["pagination"] = makePagination(int(total), page)

		ctx.HTML(200, "city")
	})
	m.Get("/okveds", func(ctx *macaron.Context) {
		ctx.Data["okveds"] = OKVEDAPI
		ctx.Data["Title"] = "Рубрики по кодам ОКВЭД"
		ctx.Data["Description"] = "Все юридические лица РФ в рубриках по кодам ОКВЭД."
		ctx.HTML(200, "okveds")
	})

	m.Get("/okveds/:cat", func(ctx *macaron.Context) {
		o, _ := OKVEDAPI.GetById(ctx.ParamsInt(":cat"))
		ctx.Data["okveds"] = OKVEDAPI
		ctx.Data["okvedsCat"] = ctx.ParamsInt(":cat")
		ctx.Data["Title"] = o.Text
		ctx.Data["Description"] = "Компании РФ в разделе " + o.Text

		ctx.HTML(200, "okvedscat")
	})

	m.Get("/stat", func(c *macaron.Context) {
		c.Data["stat"] = struct {
			DocCount   int64
			IndexSpeed float64
		}{
			DocumentsCount,
			StatIndexSpeed,
		}
		c.HTML(200, "stat")
	})

	m.Get("/:id/dump", func(ctx *macaron.Context) {
		var (
		//id = ctx.ParamsInt64(":id")
		)

		dl, id, e := models.LookUpLoc(ctx.Params(":id"))
		if e != nil {
			color.Red("%s", e)
		}

		bts, e := models.Dump(dl, id)
		if e != nil {
			color.Red("%s", e)
		}
		ctx.Resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, e = ctx.Write(bts)
		if e != nil {
			color.Red("%s", e)
		}
		return
	})

	m.Get("/:id/title", func(c *macaron.Context) {
		var (
			id = c.ParamsInt64(":id")
		)
		title, e := models.GetTitle(id)
		if e != nil {
			color.Red("%s", e)
		}
		c.JSON(200, title)
	})

	m.Get("/:id", func(ctx *macaron.Context) {
		var (
			id = ctx.ParamsInt64(":id")
		)

		if id == 0 {
			ctx.NotFound()
			return
		}

		o, e := models.GetOrg(id)
		if e != nil {
			color.Red("%s", e)
			ctx.Resp.WriteHeader(404)
			ctx.Write([]byte("<!doctype html><title></title>Не найдено <a href='/'>Домой</a>"))
			return
		}

		var (
			docLoc string
			docId  int
		)

		if o != nil {
			docLoc = o.DocLocation
			docId = o.DocId
		}

		//start := time.Now()
		s, e := models.GetSvul(ctx.Params(":id"), docLoc, docId)
		if e != nil {
			color.Red("%s", e)
		}
		//color.Green("Svul getted %s", time.Since(start))

		//start = time.Now()
		orgs, e := models.SimilarOrgs(id, 3)
		if e != nil {
			color.Red("%s", e)
		}
		//color.Green("Similar orgs %s", time.Since(start))
		/*for _, v := range orgs {
			go models.GetSvul(fmt.Sprint(v.OGRN), v.DocLocation, v.DocId)
		}*/
		/*	var orgs []models.Org
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
			}*/

		// todo tmp
		//search.IndexCity(o.OGRN, strings.ToLower(o.City))

		ctx.Data["Title"] = s.Name.FullName
		if s.Name.ShortName != "" {
			ctx.Data["Title"] = s.Name.ShortName
		}
		ctx.Data["Description"] = "Информация о компании " + s.Name.FullName
		ctx.Data["okveds"] = OKVEDAPI
		ctx.Data["org"] = o
		ctx.Data["moreScripts"] = []string{"http://api-maps.yandex.ru/2.1/?lang=ru_RU", "/js/yamaps.js"}
		ctx.Data["orgs"] = orgs
		ctx.Data["svul"] = s
		ctx.HTML(200, "get")
	})

	m.Get("/map", func(ctx *macaron.Context) {
		page := com.StrTo(ctx.Query("p")).MustInt()
		if page < 1 {
			page = 1
		}

		orgs, total, e := models.RangeMetaOrgs(page)
		if e != nil {
			color.Red("%s", e)
		}
		total = models.OgrnsCount()
		//for _, v := range orgs {
		//go models.GetSvul(fmt.Sprint(v.OGRN), v.DocLocation, v.DocId)
		//}

		ctx.Data["paginater"] = paginater.New(int(total), 10, page, 10)

		ctx.Data["nextPage"] = page + 1
		ctx.Data["prevPage"] = page - 1
		ctx.Data["lastPage"] = total / 10
		ctx.Data["currPage"] = page
		ctx.Data["orgs"] = orgs
		ctx.Data["DocumentsCount"] = total
		ctx.Data["total"] = total
		ctx.Data["pagination"] = makePagination(total, page)
		ctx.Data["Title"] = fmt.Sprintf("Карта сайта, страница %d", page)
		ctx.HTML(200, "map")
	})

	numInSitemap := 50000
	color.Green("Count Orgs %d", models.OgrnsCount())
	for i := 1; i < models.OgrnsCount(); i += numInSitemap {
		//L.Trace("Registered %d", i)
		m.Get("/sitemap."+fmt.Sprint(i)+".xml", func(c *macaron.Context) {
			color.Cyan("ACCEPT")
			c.Resp.Header().Set("Content-Type", "text/xml; charset=utf-8")
			var i int
			arr := strings.Split(c.Req.RequestURI, ".")
			if len(arr) != 3 {
				color.Red("bad")
				return
			}
			i = com.StrTo(arr[1]).MustInt()

			color.Green("Get sitemap %d", i)
			wr, e := sitemap.NewWriter(c.Resp)
			if e != nil {
				color.Red("%s", e)
			}

			ids, _, e := models.OgrnsGoodRange(int64(i), int64(numInSitemap))
			if e != nil {
				color.Red("%s", e)
			}

			var it sitemap.Item
			for _, v := range ids {
				it.Loc = "http://www." + setting.Domain + "/" + fmt.Sprint(v)
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

	}
	m.Get("/sitemap.xml", func(c *macaron.Context) {
		c.Resp.Header().Set("Content-Type", "text/xml; charset=utf-8")
		c.WriteHeader(200)

		wr, e := sitemap.NewIndexWriter(c.Resp)
		if e != nil {
			color.Red("%s", e)
		}

		for i := 1; i < models.OgrnsCount(); i++ {
			e = wr.Put(sitemap.NewIndexItem("http://www."+setting.Domain+"/sitemap."+fmt.Sprint(i)+".xml",
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

	//models.StartCrawler()
	//go titles.Index()
	titles.InitDb()
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
	//fmt.Println(dur)
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

	//color.Yellow("%d %d, %v", start, end, res)
	return res
}
