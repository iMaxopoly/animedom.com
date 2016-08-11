package main

import (
	"fmt"
	"log"
	"time"

	"animedom.com/common"
	"animedom.com/myanimelistAssistant"
	"animedom.com/projectModels"
	"animedom.com/routes"
	"animedom.com/templates"

	// 3rd Party libraries
	"animedom.com/gogoanimeAssistant"
	"github.com/iris-contrib/middleware/logger"
	"github.com/iris-contrib/middleware/recovery"
	"github.com/kataras/iris"
	"github.com/kataras/iris/config"
)

var whiteListedDomains []string

func main() {

	if !common.Production {
		whiteListedDomains = []string{"localhost:1993", "animedom.to", "animedom.com", "animedom.tv", "animedom.net", "animedom.org"}
	} else {
		whiteListedDomains = []string{"animedom.to", "animedom.com", "animedom.tv", "animedom.net", "animedom.org"}
	}

	// Init Iris
	animedom := iris.New()

	/* Declare and Initialize Middlewares */
	animedomLogger := logger.New(iris.Logger, logger.Config{
		// Status displays status code (bool)
		Status: true,
		// IP displays request's remote address (bool)
		IP: true,
		// Method displays the http method (bool)
		Method: true,
		// Path displays the request path (bool)
		Path: true,
		// EnableColors defaults to false
		EnableColors: true,
	})
	animedom.Use(animedomLogger)

	verifyDomains := func(c *iris.Context) {
		if !common.StringInSlice(c.HostString(), whiteListedDomains) {
			c.EmitError(403)
			return
		}
		c.Next()
	}
	animedom.UseFunc(verifyDomains)
	animedom.Use(recovery.New(iris.Logger))

	/* Declare static files directory */
	if !common.Production {
		animedom.Static("/assets", "./assets", 1)
	} else {
		animedom.Static("/assets", "expose/assets", 1)
	}

	/*  SEO */
	animedom.Get("/robots.txt", routes.Robots)
	animedom.Get("/sitemaps/:domain/:filename", routes.Sitemap)

	/* Define view routes */
	animedom.Get("/", routes.RouteIndex)

	animedom.Get("/coming-soon", routes.RouteComingSoon)

	animedom.Get("/disclaimer", routes.RouteDisclaimer)

	animedom.Get("/watch/:malID/:hash/:episodeNum/m/:mirrorNum", routes.RouteWatchVideo)

	animedom.Post("/goto-episode", routes.RouteGotoEpisode)

	animedom.Get("/wiki/:malID/:hash", routes.RouteAnimeDetails)

	animedom.Get("/blogs", func(c *iris.Context) {
		c.Redirect("blogs/1")
	})
	animedom.Get("/blogs/:page_num", routes.RouteBlogList)

	animedom.Get("/blog/:title", routes.RouteBlog)

	animedom.Get("/inspiring-anime", routes.RouteMoodInspireMe)
	animedom.Get("/thrilling-anime", routes.RouteMoodThriller)
	animedom.Get("/emotional-anime", routes.RouteMoodTearJerker)
	animedom.Get("/ecchi-anime", routes.RouteMoodEcchi)
	animedom.Get("/action-anime", routes.RouteMoodAction)
	animedom.Get("/classic-anime", routes.RouteClassicAnimes)

	animedom.Get("/popular-anime", func(c *iris.Context) {
		c.Redirect("/popular-anime/1")
	})
	animedom.Get("/popular-anime/:page_num", routes.RoutePopularAnimes)

	animedom.Get("/a-z-list", func(c *iris.Context) {
		c.Redirect("/a-z-list/A/1")
	})
	animedom.Get("/a-z-list/:alphabet", func(c *iris.Context) {
		alphabet := c.Param("alphabet")
		c.Redirect(fmt.Sprintf("/a-z-list/%s/1", alphabet))
	})
	animedom.Get("/a-z-list/:alphabet/:page_num", routes.RouteAZList)

	//animedom.Get("/register", routes.RouteRegisterView)
	//animedom.Post("/register", routes.RouteRegisterPost)
	//
	//animedom.Get("/login", routes.RouteLoginView)
	//animedom.Post("/login", routes.RouteLoginPost)

	animedom.Post("/genre/filter/:page_num", routes.RouteGenreFilter)

	animedom.Post("/search/animelist", routes.RouteSearchMini)

	animedom.Get("/search", routes.RouteSearchGrid)

	animedom.Get("/latest-episodes", func(c *iris.Context) {
		c.Redirect("latest-episodes/1")
	})
	animedom.Get("/latest-episodes/:page_num", routes.RouteLatestEpisodes)

	animedom.Get("/ongoing-series", func(c *iris.Context) {
		c.Redirect("/ongoing-series/1")
	})
	animedom.Get("/ongoing-series/:page_num", routes.RouteOngoingSeries)

	animedom.Get("/top-rating", func(c *iris.Context) {
		c.Redirect("top-rating/1")
	})
	animedom.Get("/top-rating/:page_num", routes.RouteTopRating)

	animedom.Get("/movies", func(c *iris.Context) {
		c.Redirect("movies/1")
	})
	animedom.Get("/movies/:page_num", routes.RouteMovies)

	animedom.Get("/feeling-lucky", routes.RouteFeelingLucky)

	/* Custom 404 handler */
	animedom.OnError(404, func(c *iris.Context) {
		c.Response.Header.SetContentType("text/html; charset=utf-8")
		templates.WritePageLayout(c.Response.BodyWriter(), projectModels.StructurePage{
			Selection: "ErrorMessage",
			ErrorMsg:  "The page you were looking for could not be found... (┛◉Д◉)┛彡┻━┻",
			BaseURL:   fmt.Sprintf("http://%s", c.HostString()),
		})
	})

	/* Run the goroutines */
	log.Println("Initiating CRON jobs...")

	// Sitemap generator
	//go func() {
	//	for {
	//		generateSitemap()
	//		time.Sleep(2 * time.Hour)
	//	}
	//}()

	// Read new articles into db
	//go func() {
	//	for {
	//		readNewBlogs()
	//		time.Sleep(6 * time.Hour)
	//	}
	//}()

	// Get a list of popular animes from MAL every 24 hours
	go func() {
		for {
			myanimelistAssistant.MonitorPopularAnimes()
			time.Sleep(24 * time.Hour)
		}
	}()

	// Get a list of seasonal animes every 6 hours
	go func() {
		for {
			myanimelistAssistant.MonitorSeasonalAnimes()
			time.Sleep(6 * time.Hour)
		}
	}()

	// Recent episodes/anime check on gogoanime.io every 7 minutes
	//go func() {
	//	for {
	//		runTask(gogoanimeAssistant.MonitorRecentEpisodes)
	//		time.Sleep(5 * time.Minute)
	//	}
	//}()

	// Refresh all animes from gogoanime.io every 48 hours
	go func() {
		for {
			runTask(gogoanimeAssistant.RefreshAnimes)
			time.Sleep(48 * time.Hour)
		}
	}()

	// Fetch Latest MALInfo weekly
	//go func() {
	//	for {
	//		runTask(myanimelistAssistant.RefreshAllProperties)
	//		time.Sleep(24 * 7 * time.Hour)
	//	}
	//}()

	/* Launch the server */
	conf := config.Server{
		Name:          "animedomHTTP v1.1",
		ListeningAddr: "localhost:1993",
	}
	animedom.ListenTo(conf)
}

func runTask(task func()) {
	for {
		if common.GetWorkOngoing() {
			time.Sleep(10 * time.Second)
		} else {
			common.SetWorkOngoing(true)
			task()
			common.SetWorkOngoing(false)
			return
		}
	}
}
