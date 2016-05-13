package main

import (
	"animedom.com/templates"
	"strconv"

	"github.com/kataras/iris"
	"fmt"
	"log"
)

func routeIndex(c *iris.Context) {
	ongoingAnimes, err := dbFetchXOngoingAnime(0, 20)
	if err != nil {
		log.Println(err)
		c.WriteHTML(200, "Nothing more to show.")
		return
	}

	popularAnimes, err := dbFetchXPopularAnime(12)
	if err != nil {
		log.Println(err)
		c.WriteHTML(200, "Nothing more to show.")
		return
	}
	c.Response.Header.SetContentType("text/html; charset=utf-8")
	templates.WritePageLayout(c.Response.BodyWriter(), templates.Page{
		Selection:"Home",
		OngoingAnimes:ongoingAnimes,
		PopularAnimes:popularAnimes,
	})
}

func routeWatchVideo(c *iris.Context) {
	c.Response.Header.SetContentType("text/html; charset=utf-8")
	templates.WritePageLayout(c.Response.BodyWriter(), templates.Page{Selection:"WatchVideo"})
}

func routeGenreFilter(c *iris.Context) {
	c.Render("", nil)
}

func routeAnimeDetails(c *iris.Context) {
	c.Render("", nil)
}

func routeLatestEpisodes(c *iris.Context) {
	c.Response.Header.SetContentType("text/html; charset=utf-8")
	templates.WritePageLayout(c.Response.BodyWriter(), templates.Page{
		Selection:"GridDisplay",
		Title:"Latest Episodes",
	})
}

func routeOngoingSeries(c *iris.Context) {
	var pageNumInt int
	var err error
	if pageNumInt, err = strconv.Atoi(c.Param("page_num")); err != nil || pageNumInt == 0 {
		c.NotFound()
		return
	}

	animelist, err := dbFetchXOngoingAnime((pageNumInt - 1) * 30, 30)
	if err != nil {
		if err.Error() == "Empty Result" {
			c.WriteHTML(200, "Nothing more to show.")
			return
		}
		log.Println(err.Error())
		c.WriteHTML(500, "Sorry we countered an error.")
		return
	}

	c.Response.Header.SetContentType("text/html; charset=utf-8")
	templates.WritePageLayout(c.Response.BodyWriter(), templates.Page{
		Selection:"GridDisplay",
		Title:"Ongoing Series",
		GridAnimes:animelist,
	})
}

func routeTopRating(c *iris.Context) {
	var pageNumInt int
	var err error
	if pageNumInt, err = strconv.Atoi(c.Param("page_num")); err != nil || pageNumInt == 0 {
		c.NotFound()
		return
	}

	animelist, err := dbFetchXTopRatingAnime((pageNumInt - 1) * 30, 30)
	if err != nil {
		if err.Error() == "Empty Result" {
			c.WriteHTML(200, "Nothing more to show.")
			return
		}
		fmt.Println(err.Error())
		c.WriteHTML(500, "Sorry we countered an error.")
		return
	}

	c.Response.Header.SetContentType("text/html; charset=utf-8")
	templates.WritePageLayout(c.Response.BodyWriter(), templates.Page{
		Selection:"GridDisplay",
		Title:"Top Rating",
		GridAnimes:animelist,
	})
}
