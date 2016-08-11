package gogoanimeAssistant

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"animedom.com/common"
	"animedom.com/dbOperations"
	"animedom.com/projectModels"

	"github.com/PuerkitoBio/goquery"
	"github.com/Sirupsen/logrus"
)

func MonitorRecentEpisodes() {
	// Handle error
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorln(err)
		}
		logrus.Infoln("gogoanimeAssistant.MonitorRecentEpisodes Job done")
	}()

	logrus.Infoln("gogoanimeAssistant.MonitorRecentEpisodes Job started")
	// Get entire anime listing in advance
	allAnimes, titleAndIDs := fetchRecentAnimeListing()
	count := 0
	for _, v := range titleAndIDs {
		if strings.TrimSpace(v.title) == "" || strings.TrimSpace(v.id) == "" {
			logrus.Errorln("for-top - MonitorRecentEpisodes() Spotted empty", v.title, v.id)
			continue
		}
		jobRes := mainscrape(job{whatKind: "MonitorRecentEpisodes", allAnimes: allAnimes,
			animeTitle: v.title, animeEpisodeID: v.id})
		if jobRes.error != nil {
			logrus.Warningln(jobRes.error.Error())
			common.WriteToFile(jobRes.fileWrite.Filename, jobRes.fileWrite.Message)
			continue
		}
		if strings.TrimSpace(jobRes.recentAnime.malID) == "" || jobRes.recentAnime.episodeIndex <= 0 {
			logrus.Errorln("for-bottom - MonitorRecentEpisodes() Spotted empty", v.title, v.id, jobRes.recentAnime.malID, jobRes.recentAnime.episodeIndex)
			continue
		}
		count++
		dbOperations.DBPushRecentAnimes(count, jobRes.recentAnime.malID, jobRes.recentAnime.episodeIndex)
	}

	nRecentAnimes, err := dbOperations.DBCountSize("recent_animes")
	if err != nil {
		panic(err)
	}
	if nRecentAnimes <= 0 {
		panic(errors.New("Very few recent animes?"))
	}

	err = dbOperations.DBCopyTableTo("recent_animes", "cache_recent_animes")
	if err != nil {
		panic(err)
	}
}

func fetchRecentAnimeListing() ([]projectModels.StructureNameAndLink, []titleAndID) {
	allAnimes := scrapeAnimelistPage()

	// Truncate Table
	err := dbOperations.DBTruncateTable("recent_animes")
	if err != nil {
		panic(err)
	}

	titleAndIDs := []titleAndID{}
	for page := 1; page <= 3; page++ {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(common.GetPageResponse(fmt.Sprintf("http://gogoanime.io/?page=%d", page))))
		if err != nil {
			panic(err)
		}

		latestEpisodeItems := doc.Find(".last_episodes_items")
		for node := range latestEpisodeItems.Nodes {
			animeTitle := strings.TrimSpace(latestEpisodeItems.Eq(node).Find(".name").Find("a").Text())
			if animeTitle == "" {
				panic(errors.New("Issue with reading anime title on recent animes page 1"))
			}
			animeTitle = strings.TrimSpace(animeTitle)

			animeEpisodeID := strings.TrimSpace(latestEpisodeItems.Eq(node).Find(".time").Text())
			if animeEpisodeID == "" {
				log.Panicln(errors.New("Empty Episode ID"))
			}
			// We remove the Episode from "Episode 6" for instance so we get 6
			for i, v := range animeEpisodeID {
				if v == ' ' {
					animeEpisodeID = animeEpisodeID[i+1:]
					break
				}
			}
			stitleAndID := titleAndID{title: animeTitle, id: animeEpisodeID}
			titleAndIDs = append(titleAndIDs, stitleAndID)
		}
	}
	return allAnimes, titleAndIDs
}
