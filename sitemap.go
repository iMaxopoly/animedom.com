package main

import (
	"log"

	"animedom.com/common"
	"animedom.com/dbOperations"

	"fmt"

	"github.com/ikeikeikeike/go-sitemap-generator/stm"
)

/* Sitemap generator */
func generateSitemap() {
	defer func() {
		// Handle error
		if err := recover(); err != nil {
			log.Println(err)
		}
		log.Println("generateSitemap Job done")
	}()

	animelist, err := dbOperations.DBGetAllAnime()
	if err != nil {
		panic(err)
	}

	for _, domainName := range whiteListedDomains {
		sm := stm.NewSitemap()
		sm.SetDefaultHost(fmt.Sprintf("http://%s", domainName))
		sm.SetPublicPath(fmt.Sprintf("sm/%s", domainName))
		sm.SetCompress(true)
		sm.SetVerbose(true)

		sm.Create()

		sm.Add(stm.URL{
			"loc": "/", "changefreq": "always",
			"priority": "1.00",
		})

		sm.Add(stm.URL{
			"loc": "/popular-anime", "changefreq": "monthly",
			"priority": "0.80",
		})

		sm.Add(stm.URL{
			"loc": "/top-rating", "changefreq": "monthly",
			"priority": "0.80",
		})

		sm.Add(stm.URL{
			"loc": "/classic-anime", "changefreq": "monthly",
			"priority": "0.80",
		})

		sm.Add(stm.URL{
			"loc": "/inspiring-anime", "changefreq": "monthly",
			"priority": "0.80",
		})

		sm.Add(stm.URL{
			"loc": "/thrilling-anime", "changefreq": "monthly",
			"priority": "0.80",
		})

		sm.Add(stm.URL{
			"loc": "/emotional-anime", "changefreq": "monthly",
			"priority": "0.80",
		})

		sm.Add(stm.URL{
			"loc": "/ecchi-anime", "changefreq": "monthly",
			"priority": "0.80",
		})

		sm.Add(stm.URL{
			"loc": "/action-anime", "changefreq": "monthly",
			"priority": "0.80",
		})

		sm.Add(stm.URL{
			"loc": "/latest-episodes", "changefreq": "daily",
			"priority": "0.80",
		})

		sm.Add(stm.URL{
			"loc": "/ongoing-series", "changefreq": "weekly",
			"priority": "0.80",
		})

		sm.Add(stm.URL{
			"loc": "/movies", "changefreq": "monthly",
			"priority": "0.80",
		})

		sm.Add(stm.URL{
			"loc": "/blogs", "changefreq": "daily",
			"priority": "0.80",
		})

		for _, anime := range animelist {
			sm.Add(stm.URL{
				"loc":        "/wiki/" + anime.MALID + "/" + anime.WikiHash,
				"changefreq": "weekly",
				"priority":   "0.70",
			})

			for epIndex := 0; epIndex < len(anime.SubbedEpisodeList); epIndex++ {
				for mirIndex := 0; mirIndex < len(anime.SubbedEpisodeList[epIndex].Mirrors); mirIndex++ {
					sm.Add(stm.URL{"loc": fmt.Sprintf("/watch/%s/%s/%d/m/%d",
						anime.MALID,
						common.GetMirrorHash("sub", anime, epIndex, mirIndex),
						epIndex,
						mirIndex,
					),
						"changefreq": "weekly",
						"priority":   "0.90",
					})
				}
			}
		}

		sm.Finalize().PingSearchEngines()
	}
}
