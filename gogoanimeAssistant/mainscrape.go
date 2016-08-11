package gogoanimeAssistant

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"animedom.com/common"
	"animedom.com/dbOperations"
	"animedom.com/myanimelistAssistant"
	"animedom.com/projectModels"

	"github.com/Sirupsen/logrus"
	"github.com/iris-contrib/errors"
	"github.com/tv42/slug"
)

func getEpisodeIndex(version string, anime projectModels.StructureAnime, id string) int {
	switch version {
	case "sub":
		for e := 0; e < len(anime.SubbedEpisodeList); e++ {
			if anime.SubbedEpisodeList[e].EpisodeID == id {
				return e
			}
		}
	case "dub":
		for e := 0; e < len(anime.EnglishDubbedEpisodeList); e++ {
			if anime.EnglishDubbedEpisodeList[e].EpisodeID == id {
				return e
			}
		}
	case "ger":
	case "fra":
	case "spa":
	}

	panic(errors.New("getEpisodeIndex mixmatch"))
}

func getRightEpisodeList(dubbed bool, anime projectModels.StructureAnime) []projectModels.StructureEpisode {
	switch dubbed {
	case true:
		return anime.EnglishDubbedEpisodeList
	case false:
		return anime.SubbedEpisodeList
	}
	panic(errors.New("getRightEpisodeList HUH"))
}

func checkifDub(title string) bool {
	switch strings.HasSuffix(title, "(Dub)") {
	case true:
		return true
	case false:
		return false
	}
	return false
}

func mainscrape(order job) jobResult {
	// Handle error
	defer func() {
		if err := recover(); err != nil {
			logrus.Errorln(err)
		}
	}()

	var startIndex, endIndex int
	if order.whatKind == orderMonitorRecentEpisodes {
		startIndex = 0
		endIndex = 1
	} else {
		startIndex = order.start
		endIndex = order.end
	}

	for index := startIndex; index < endIndex; index++ {
		var isDub bool

		// Checking if anime exists in our Fix list, replacing the name with the fixed one therefore
		var fixedAnimeName string
		if order.whatKind == orderMonitorRecentEpisodes {
			isDub = checkifDub(order.animeTitle)

			fixedAnimeName = myanimelistAssistant.CorrectNameForMALLookup(order.animeTitle)
			// If our fixed list doesnt contain it, we replace fixedAnimeName with original name
			// to later check against MAL database
			if fixedAnimeName == "" {
				fixedAnimeName = order.animeTitle
			}
		} else {
			logrus.Println("gogoanime|mainscrape|", order.whatKind, startIndex, endIndex)
			isDub = checkifDub(order.allAnimes[index].Name)

			var langName string
			if isDub {
				langName = strings.TrimSpace(strings.TrimSuffix(order.allAnimes[index].Name, "(Dub)"))
			} else {
				langName = strings.TrimSpace(strings.TrimSuffix(order.allAnimes[index].Name, "(Sub)"))
			}
			fixedAnimeName = myanimelistAssistant.CorrectNameForMALLookup(langName)
			// If our fixed list doesnt contain it, we replace fixedAnimeName with original name
			// to later check against MAL database
			if fixedAnimeName == "" {
				fixedAnimeName = langName
			}
		}

		if order.whatKind == orderMonitorRecentEpisodes {
			logrus.Println("gogoanime|mainscrape|", order.whatKind, "Checking Recent Anime", fixedAnimeName)
		} else {
			logrus.Println("gogoanime|mainscrape|", order.whatKind, "Checking Anime", fixedAnimeName)
		}

		// Get info about anime from MAL
		malAnimeData, err := myanimelistAssistant.MalAPIFetchData(fixedAnimeName)
		if err != nil {
			// If we cannot find this anime in MAL, we log this name in excludedAnimes.txt
			// for later review
			if strings.Contains(err.Error(), "EOF") {
				// This anime is not found in MAL
				if order.whatKind == orderMonitorRecentEpisodes {
					return jobResult{
						fileWrite: projectModels.FileWrite{
							Filename: common.FileExcludedAnimes,
							Message:  fixedAnimeName,
						},
						error: fmt.Errorf("Couldn't find in MAL database - %s", fixedAnimeName),
					}
				} else {
					common.WriteToFile(common.FileExcludedAnimes, fixedAnimeName)
					continue
				}

			}
			// If the error is something else we fucking panic
			if err != nil {
				panic(err)
			}
		}

		// Anime is in Myanimelist, , so we perform deep search in xml to get the right ID
		malIDIndex := myanimelistAssistant.FindIDinMALXML(malAnimeData, fixedAnimeName)
		if malIDIndex == -1 {
			// We couldn't find it in XML map
			if order.whatKind == orderMonitorRecentEpisodes {
				return jobResult{
					fileWrite: projectModels.FileWrite{
						Filename: common.FileMALNotFoundInXML,
						Message:  fixedAnimeName,
					},
					error: fmt.Errorf("Couldn't find in MAL xml map - %s", fixedAnimeName),
				}
			} else {
				common.WriteToFile(common.FileMALNotFoundInXML, fixedAnimeName)
				continue
			}
		}

		// locate node anime in allAnimes
		indexInAnimeList := func() int {
			for i, v := range order.allAnimes {
				findIndexInGogolist := func() string {
					animeName := myanimelistAssistant.CorrectNameForMALLookup(v.Name)
					if animeName == "" {
						animeName = v.Name
					}
					return animeName
				}
				switch isDub {
				case true:
					locName := fixedAnimeName + " (Dub)"
					if locName == findIndexInGogolist() {
						return i
					}
				case false:
					locName := fixedAnimeName + " (Sub)"
					if locName == findIndexInGogolist() || fixedAnimeName == findIndexInGogolist() {
						return i
					}
				}

			}
			return -1
		}()
		if indexInAnimeList == -1 {
			if order.whatKind == orderMonitorRecentEpisodes {
				return jobResult{
					fileWrite: projectModels.FileWrite{
						Filename: common.FileNotInFixedList,
						Message:  fixedAnimeName,
					},
					error: fmt.Errorf("Not in fixed list - %s", fixedAnimeName),
				}
			}
			common.WriteToFile(common.FileNotInFixedList, fixedAnimeName)
			continue
		}

		mAnime, err := dbOperations.DBFetchAnimeByCol("id", malAnimeData.Entries[malIDIndex].ID)
		if err == nil {
			// We have this anime in our database, but we MIGHT be behind
			// we we will count total eps we have VS total eps gogoanime has

			// Necessary to get a full listing of all episodes and episode links so we can extract videos
			animeId, defaultEp := scrapeAnimeID(order.allAnimes[indexInAnimeList].Link)

			// Next we scrape episode lists for given anime
			episodes := scrapeEpisodeLists(animeId, defaultEp)

			if order.whatKind == orderMonitorRecentEpisodes {
				logrus.Println("gogoanime|mainscrape|", order.whatKind, malAnimeData.Entries[malIDIndex].Title, "has", len(episodes))
				switch isDub {
				case true:
					logrus.Println("gogoanime|mainscrape|", order.whatKind, mAnime.MALTitle, "has", len(mAnime.EnglishDubbedEpisodeList))
				case false:
					logrus.Println("gogoanime|mainscrape|", order.whatKind, mAnime.MALTitle, "has", len(mAnime.SubbedEpisodeList))
				}

				// Now we count if total number of episodes is more than ours

				difference := len(episodes) - len(getRightEpisodeList(isDub, mAnime))
				mirrorsSame := func() bool {
					// We fake check last episode URL to see if they fucking changed
					if len(episodes) <= 0 {
						return true
					}

					var wpage watchPage
					wpage.url = episodes[len(episodes)-1].episodeURL
					wpage.getBody()
					wpage.getIFrames()
					wpage.getGoogleURL()

					if len(wpage.googleMirrors) <= 0 {
						// We check if the episodes are available or if it's a fake entry
						return true

					}

					var gogoMirrors []string
					var animedomMirrors []string

					// Collect remote mirror links
					for _, v := range wpage.googleMirrors {
						gogoMirrors = append(gogoMirrors, v)
					}

					// Collect local mirror links
					for _, v := range getRightEpisodeList(isDub, mAnime)[len(getRightEpisodeList(isDub, mAnime))-1].Mirrors {
						if strings.Contains(v.Name, "Alpha") || strings.Contains(v.Name, "Beta") {
							animedomMirrors = append(animedomMirrors, v.EmbedCode)
						}
					}

					sort.Strings(gogoMirrors)
					sort.Strings(animedomMirrors)

					// Test if they are same
					return common.TestSliceEq(gogoMirrors, animedomMirrors)
				}()

				if difference <= 0 && mirrorsSame {
					logrus.Println("gogoanime|mainscrape|", order.whatKind, malAnimeData.Entries[malIDIndex].Title, "is already updated")
					// So this anime is updated anyway

					// We push this to recent animes table
					return jobResult{
						fileWrite: projectModels.FileWrite{},
						recentAnime: recentAnime{malID: mAnime.MALID, episodeIndex: getEpisodeIndex(func() string {
							if isDub {
								return "dub"
							} else {
								return "sub"
							}
						}(), mAnime, order.animeEpisodeID)},
						error: nil,
					}
				}
			}

			if order.whatKind == orderMonitorRecentEpisodes {
				// At this point anime isn't updated, we need to fetch the remnant episodes
				logrus.Println("gogoanime|mainscrape|", order.whatKind, malAnimeData.Entries[malIDIndex].Title, "needs to be updated.")
			} else {
				logrus.Println("gogoanime|mainscrape|", order.whatKind, malAnimeData.Entries[malIDIndex].Title, "being refreshed.")
			}

			// Now if all good we update episode listing for this anime
			var fEpisodes []projectModels.StructureEpisode
			fEpisodes, err = setEpisodes(order, malAnimeData, malIDIndex, mAnime, episodes)
			if err != nil {
				if order.whatKind == orderMonitorRecentEpisodes {
					return jobResult{
						fileWrite: projectModels.FileWrite{
							Filename: common.FileFakeEntry,
							Message:  fixedAnimeName,
						},
						error: fmt.Errorf("Skipping for no videos on - %s", fixedAnimeName),
					}
				} else {
					common.WriteToFile(common.FileFakeEntry, fixedAnimeName)
					continue
				}
			}

			if isDub {
				mAnime.EnglishDubbedEpisodeList = fEpisodes
			} else {
				mAnime.SubbedEpisodeList = fEpisodes
			}

			// Set hash for mirrors
			for hashIndex := 0; hashIndex < len(getRightEpisodeList(isDub, mAnime)); hashIndex++ {
				for mhashIndex := 0; mhashIndex < len(getRightEpisodeList(isDub, mAnime)[hashIndex].Mirrors); mhashIndex++ {
					getRightEpisodeList(isDub, mAnime)[hashIndex].Mirrors[mhashIndex].MirrorHash =
						common.GetFNV(fmt.Sprintf("mirror%s%s%d%d%s",
							mAnime.MALID,
							mAnime.Slug,
							hashIndex,
							mhashIndex,
							common.HashMirror))
				}
			}

			logrus.Println("gogoanime|mainscrape|", order.whatKind, "Updating", mAnime.MALTitle)
			err = dbOperations.DBUpdateEpisodelist(getRightEpisodeList(isDub, mAnime), mAnime.MALID)
			if err != nil {
				panic(err)
			}

			// We push this to recent animes table
			if order.whatKind == orderMonitorRecentEpisodes {
				return jobResult{
					fileWrite: projectModels.FileWrite{},
					recentAnime: recentAnime{malID: mAnime.MALID, episodeIndex: getEpisodeIndex(func() string {
						if isDub {
							return "dub"
						} else {
							return "sub"
						}
					}(), mAnime, order.animeEpisodeID)},
					error: nil,
				}
			}
			continue
		}
		// Finally if err != nil, which means either legit error or we don't have this anime
		if err != nil {
			if err.Error() != "Empty Result" {
				panic(err)
			}
		}

		// We will fetch the new anime

		logrus.Println("gogoanime|mainscrape|", order.whatKind, "Gotta add the new anime", malAnimeData.Entries[malIDIndex].Title)
		// Fetching MAL details
		sAnime := projectModels.StructureAnime{}
		sAnime.MALID = malAnimeData.Entries[malIDIndex].ID
		sAnime.MALTitle = malAnimeData.Entries[malIDIndex].Title
		sAnime.MALEnglish = malAnimeData.Entries[malIDIndex].EnglishName
		sAnime.Genre = myanimelistAssistant.MalFetchGenre(sAnime.MALID)
		sAnime.MALDescription = myanimelistAssistant.CleanMALDescription(malAnimeData.Entries[malIDIndex].Synopsis)
		sAnime.Score = func() float64 {
			var i float64
			i, err = strconv.ParseFloat(malAnimeData.Entries[malIDIndex].Score, 64)
			if err == nil {
				return i
			}
			return 0.0
		}()
		sAnime.Status = malAnimeData.Entries[malIDIndex].Status
		sAnime.Type = malAnimeData.Entries[malIDIndex].Type
		sAnime.SynonymNames = func() []string {
			s := strings.Split(malAnimeData.Entries[malIDIndex].SynonymNames, ";")
			if len(s) < 1 {
				return []string{}
			}
			var val []string
			for _, v := range s {
				val = append(val, strings.TrimSpace(v))
			}
			return val
		}()

		sAnime.Year = malAnimeData.Entries[malIDIndex].StartDate
		sAnime.Image = strings.Replace(malAnimeData.Entries[malIDIndex].Image, ".jpg", "l.jpg", 1)
		sAnime.Trailer = myanimelistAssistant.MalFetchTrailer(sAnime.MALID)
		sAnime.Slug = slug.Slug(sAnime.MALTitle)

		// Necessary to get a full listing of all episodes and episode links so we can extract videos
		animeId, defaultEp := scrapeAnimeID(order.allAnimes[indexInAnimeList].Link)

		// Next we scrape episode lists for given anime
		episodes := scrapeEpisodeLists(animeId, defaultEp)

		// Attach episodeList to anime
		fEpisodes, err := setEpisodes(order, malAnimeData, malIDIndex, sAnime, episodes)
		if err != nil {
			if order.whatKind == orderMonitorRecentEpisodes {
				return jobResult{
					fileWrite: projectModels.FileWrite{
						Filename: common.FileFakeEntry,
						Message:  fixedAnimeName,
					},
					error: fmt.Errorf("Skipping for no videos on - %s", fixedAnimeName),
				}
			} else {
				common.WriteToFile(common.FileFakeEntry, fixedAnimeName)
				continue
			}
		}

		if isDub {
			sAnime.EnglishDubbedEpisodeList = fEpisodes
		} else {
			sAnime.SubbedEpisodeList = fEpisodes
		}

		// Set hash for wiki
		sAnime.WikiHash = common.GetFNV("wiki" +
			sAnime.MALID +
			sAnime.Slug +
			common.HashWiki)

		// Set hash for mirrors
		for hashIndex := 0; hashIndex < len(getRightEpisodeList(isDub, sAnime)); hashIndex++ {
			for mhashIndex := 0; mhashIndex < len(getRightEpisodeList(isDub, sAnime)[hashIndex].Mirrors); mhashIndex++ {
				getRightEpisodeList(isDub, sAnime)[hashIndex].Mirrors[mhashIndex].MirrorHash =
					common.GetFNV(fmt.Sprintf("mirror%s%s%d%d%s",
						sAnime.MALID,
						sAnime.Slug,
						hashIndex,
						mhashIndex,
						common.HashMirror))
			}
		}

		logrus.Println("gogoanime|mainscrape|", order.whatKind, "Episode list for", sAnime.MALTitle, "is", getRightEpisodeList(isDub, sAnime))

		// Finally fetch the anime image from MAL
		logrus.Println("gogoanime|mainscrape|", order.whatKind, "Fetching image for", sAnime.MALTitle)
		err = myanimelistAssistant.MalFetchImage(sAnime.Image, sAnime.MALID)
		if err != nil {
			if order.whatKind == orderMonitorRecentEpisodes {
				return jobResult{
					fileWrite: projectModels.FileWrite{
						Filename: common.FileNoImage,
						Message:  fixedAnimeName,
					},
					error: fmt.Errorf("Couldnt find image - %s", fixedAnimeName),
				}
			} else {
				common.WriteToFile(common.FileNoImage, fixedAnimeName)
				continue
			}
		}

		// Insert anime to rethinkdb
		err = dbOperations.DBInsertNewAnime(sAnime)
		if err != nil {
			panic(err)
		}

		// We push this to recent animes table
		logrus.Println("gogoanime|mainscrape|", order.whatKind, "Inserted New anime successfully", sAnime.MALTitle)

		if order.whatKind == orderMonitorRecentEpisodes {
			return jobResult{
				fileWrite: projectModels.FileWrite{},
				recentAnime: recentAnime{malID: sAnime.MALID, episodeIndex: getEpisodeIndex(func() string {
					if isDub {
						return "dub"
					} else {
						return "sub"
					}
				}(), sAnime, order.animeEpisodeID)},
				error: nil,
			}
		}
	}
	fmt.Println("Returning 0")
	return jobResult{}
}

func setMirrors(googleVids []googleMirror, mp4mirrors, moestreaMirrors []string, episodeIndex int, malID, slug string) []projectModels.StructureMirror {

	mirrorList := []projectModels.StructureMirror{}

	// Storing Google Based vids
	if len(googleVids) > 0 {
		for i, j := range googleVids {
			mirror := projectModels.StructureMirror{}
			mirror.Name = j.Name
			mirror.Quality = j.Quality
			mirror.EmbedCode = j.URL
			mirror.MirrorOrigin = j.MirrorOrigin
			mirror.MirrorHash = common.GetFNV(fmt.Sprintf("mirror%s%s%d%d%s",
				malID,
				slug,
				episodeIndex,
				i,
				common.HashMirror))
			mirrorList = append(mirrorList, mirror)
		}
	}

	if len(moestreaMirrors) > 0 {
		for i, j := range moestreaMirrors {
			mirror := projectModels.StructureMirror{}
			mirror.Name = "stream"
			mirror.EmbedCode = j
			mirror.MirrorHash = common.GetFNV(fmt.Sprintf("mirror%s%s%d%d%s",
				malID,
				slug,
				episodeIndex,
				len(googleVids)+i,
				common.HashMirror))
			mirrorList = append(mirrorList, mirror)
		}
	}

	if len(mp4mirrors) > 0 {
		for _, j := range mp4mirrors {
			mirror := projectModels.StructureMirror{}
			mirror.Name = "mp4upload"
			mirror.EmbedCode = j
			mirrorList = append(mirrorList, mirror)
		}
	}

	return mirrorList
}

func setEpisodes(j job, malAnimeData projectModels.StructureMALApiAnime, malIDIndex int, mAnime projectModels.StructureAnime, episodes []episodelist) ([]projectModels.StructureEpisode, error) {
	// We get all episode names from MAL to use the new names if there is
	malEpisodeNames := myanimelistAssistant.MalFetchEpisodelist(malAnimeData.Entries[malIDIndex].ID)

	episodeList := []projectModels.StructureEpisode{}

	// episodeUnavailable for checking if episode has no vids
	episodeUnavailable := false
	for episodeIndex := 0; episodeIndex < len(episodes); episodeIndex++ {
		// Saving videos of each category in respective variables
		logrus.Println("gogoanime|mainscrape|", j.whatKind, "Fetching for", mAnime.MALTitle, "episode", episodes[episodeIndex].episodeID)

		var wpage watchPage
		wpage.url = episodes[episodeIndex].episodeURL
		wpage.getBody()
		wpage.getIFrames()
		wpage.getGoogleURL()
		wpage.mp42Stream(malAnimeData.Entries[malIDIndex].ID + episodes[episodeIndex].episodeID)

		if len(wpage.googleMirrors) <= 0 {
			// We check if the episodes are available or if it's a fake entry
			episodeUnavailable = true
			break

		}

		thisEpisode := projectModels.StructureEpisode{}

		// Append mirrors to this episode's mirror model
		thisEpisode.Mirrors = setMirrors(wpage.googleMirrors, wpage.mp4uploadMirrors, wpage.moestreamMirrors, episodeIndex, mAnime.MALID, mAnime.Slug)

		// Put episode Name
		thisEpisode.Name = func() string {
			if episodeIndex >= len(malEpisodeNames) || episodeIndex < 0 {
				return ""
			}
			return malEpisodeNames[episodeIndex]
		}()

		// Put episode ID
		thisEpisode.EpisodeID = episodes[episodeIndex].episodeID

		// Append to episode list for this anime
		//		logrus.Println("Inserting episode", thisEpisode.EpisodeID, "to", anime.MALTitle)
		episodeList = append(episodeList, thisEpisode)
	}
	// If certain episode vids unavailable we skip this anime
	if episodeUnavailable == true {
		return []projectModels.StructureEpisode{}, fmt.Errorf("Skipping for no videos on - %s", mAnime.MALTitle)
	}

	if len(episodeList) < 1 {
		return []projectModels.StructureEpisode{}, fmt.Errorf("No episodes? %s", mAnime.MALTitle)
	}

	return episodeList, nil
}
