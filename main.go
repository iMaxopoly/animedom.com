package main

import (
	"os"
	//"time"
	"log"

	// 3rd Party libraries
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recovery"
	"github.com/kataras/iris/middleware/logger"
)

func main() {
	/* Declare and Initialize Middlewares */
	animedom := iris.New()
	animedom.UseFunc(logger.Default())
	animedom.Use(recovery.New(os.Stderr))

	/* Declare static files directory */
	animedom.Static("/assets", "./assets", 1)

	/* Define view routes */
	// Home Page, this page will contain a thumbnailed grid layout of all ongoing animes
	// followed by another row of popular animes, can either precede "ongoing anime" list
	// or be put later on. Can also sport popular genres.
	animedom.Get("/", routeIndex)

	// Page to watch the video, the page will have a link to the main anime page which
	// will contain an episode list, anime score and trailer
	// Page will also contain a link to next episode or previous episode if the episodes exist.
	// eg. animedom.com/naruto/{insert episode name}
	animedom.Get("/watch/:animeName/:episodeName", routeWatchVideo)

	// Page to list filtered animes based on Genre.
	// Just a convenience feature to list popular genre animes together for people with
	// specific taste.
	// eg. animedom.com/genre/filter [POST Request]
	animedom.Get("/genre/filter", routeGenreFilter)

	// Page to show anime details, along with episode listing.
	// Basically the motherhub for any given anime. Sporting score for the anime which will
	// link people to myanimelist.net for further reading. Page will also contain trailer for
	// said anime.
	// eg. animedom.com/narutoshippuden
	animedom.Get("/:animeName", routeAnimeDetails)

	// Page to show latest episodes
	animedom.Get("/latest-episodes", func(c *iris.Context) {
		c.Redirect("latest-episodes/1")
	})
	animedom.Get("/latest-episodes/:page_num", routeLatestEpisodes)

	// Page to show new anime releases
	animedom.Get("/ongoing-series", func(c *iris.Context) {
		c.Redirect("/ongoing-series/1")
	})
	animedom.Get("/ongoing-series/:page_num", routeOngoingSeries)

	// Page to show top rated animes
	animedom.Get("/top-rating", func(c *iris.Context) {
		c.Redirect("top-rating/1")
	})
	animedom.Get("/top-rating/:page_num", routeTopRating)

	/* Define API routes */
	// API to list filtered animes based on Genre.
	// Just a convenience feature to list popular genre animes together for people with
	// specific taste.
	// eg. animedom.com/genre/filter [POST Request]
	animedom.Post("/genre/filter", apiGenreFilter)

	/* Run the goroutines */
	// Running Popular anime check every 24 hours
	log.Println("Running tasks please wait...")
	//malMonitorPopularAnimes()
	//go func() {
	//	for {
	//		time.Sleep(24 * time.Hour)
	//		malMonitorPopularAnimes()
	//	}
	//}()
	//// Running Recent episodes/anime check every 5 minutes
	//malMonitorRecentEpisodes()
	//go func(){
	//	for {
	//		time.Sleep(5 * time.Minute)
	//		malMonitorRecentEpisodes()
	//	}
	//}()

	/* Launch the server */
	log.Println("Server is running at :1234")
	animedom.Listen(":1234")
}
