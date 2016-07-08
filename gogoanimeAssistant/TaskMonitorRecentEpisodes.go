package gogoanimeAssistant

import (
	"log"
	"strings"
	"os"
	"fmt"
	"strconv"
	"errors"

	"animedom.com/projectModels"
	"animedom.com/dbOperations"
	"animedom.com/myanimelistAssistant"

	"github.com/PuerkitoBio/goquery"
	"github.com/tv42/slug"
	"sort"
)

//func handleError(err error) {
//	defer func() {
//		if err := recover(); err != nil {
//			log.Println(err)
//		}
//	}()
//	if err != nil {
//		panic(err)
//	}
//}

func writeToFile(filename, content string) {
	file, err := os.OpenFile(filename, os.O_APPEND | os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	_, err = file.WriteString(content + "\n")
	if err != nil {
		panic(err)
	}
}

func testSliceEq(a, b []string) bool {
	if a == nil && b == nil {
		return true;
	}

	if a == nil || b == nil {
		return false;
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func setMirrors(googleVids []googleVid, vidstreamVids, mp4uploadVids []string, episodeIndex int,episodes []episodelist) ([]projectModels.StructureMirror) {

	mirrorList := []projectModels.StructureMirror{}

	// Storing Google Based vids
	if len(googleVids) > 0 {
		for _, j := range googleVids {
			mirror := projectModels.StructureMirror{}
			mirror.Name = fmt.Sprintf("HD %s", j.quality)
			mirror.SubDub = episodes[episodeIndex].subdub
			mirror.Iframe = j.address
			//				log.Println("Adding mirror", mirror.Name, "to", anime.MALTitle)
			mirrorList = append(mirrorList, mirror)
		}
	}

	// Storing vidstreamVids
	if len(vidstreamVids) > 0 {
		for i, j := range vidstreamVids {
			mirror := projectModels.StructureMirror{}
			mirror.Name = fmt.Sprintf("VidStream %d", i)
			mirror.SubDub = episodes[episodeIndex].subdub
			mirror.Iframe = j
			//				log.Println("Adding mirror", mirror.Name, "to", anime.MALTitle)
			mirrorList = append(mirrorList, mirror)
		}
	}

	// Storing mp4uploadVids
	if len(mp4uploadVids) > 0 {
		for i, j := range mp4uploadVids {
			mirror := projectModels.StructureMirror{}
			mirror.Name = fmt.Sprintf("Mp4Upload %d", i)
			mirror.SubDub = episodes[episodeIndex].subdub
			mirror.Iframe = j
			//				log.Println("Adding mirror", mirror.Name, "to", anime.MALTitle)
			mirrorList = append(mirrorList, mirror)
		}
	}
	return mirrorList
}

func MonitorRecentEpisodes() {
	// Handle error
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	// Get entire anime listing in advance

	// Truncate Table
	err := dbOperations.DBTruncateTable("recent_animes")
	if err != nil {
		panic(err)
	}

	allAnimes := scrapeAnimelistPage()
	count := 0
	for page := 1; page <= 3; page++ {

		doc, err := goquery.NewDocument(fmt.Sprintf("http://gogoanime.io/?page=%d", page))
		if err != nil {
			panic(err)
		}

		latestEpisodeItems := doc.Find(".last_episodes_items")

		for node := range latestEpisodeItems.Nodes {
			animeTitle, exists := latestEpisodeItems.Eq(node).Find("a").Attr("title")
			if !exists {
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
					animeEpisodeID = animeEpisodeID[i + 1:]
					break
				}
			}

			// Checking if anime exists in our Fix list, replacing the name with the fixed one therefore
			fixedAnimeName := myanimelistAssistant.CorrectNameForMALLookup(animeTitle)

			// If our fixed list doesnt contain it, we replace fixedAnimeName with original name
			// to later check against MAL database
			if fixedAnimeName == "" {
				fixedAnimeName = animeTitle
			}
			log.Println("Checking Recent Anime", fixedAnimeName)
			// Get info about anime from MAL
			malAnimeData, err := myanimelistAssistant.MalAPIFetchData(fixedAnimeName)
			if err != nil {
				// If we cannot find this anime in MAL, we log this name in excludedAnimes.txt
				// for later review
				if strings.Contains(err.Error(), "EOF") {
					writeToFile(fileExcludedAnimes, fixedAnimeName)
					log.Println("Couldn't find in MAL database", fixedAnimeName)
					// We move on to next anime since this anime is not found in MAL
					continue
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
				log.Println("Couldn't find in MAL xml map", fixedAnimeName)
				writeToFile(fileMALNotFoundInXML, fixedAnimeName)
				continue
			}

			// locate node anime in allAnimes
			animeIndex := func() (int) {
				for i, v := range allAnimes {
					if fixedAnimeName == func() (string) {
						animeName := myanimelistAssistant.CorrectNameForMALLookup(v.Name)
						if animeName == "" {
							animeName = v.Name
						}
						return animeName
					}() {
						return i
					}
				}
				return -1
			}()
			if animeIndex == -1 {
				panic(errors.New("Couldn't locate animeIndex in allAnimes"))
			}

			err = dbOperations.DBCheckExistsAnimesByID(malAnimeData.Entries[malIDIndex].ID)
			if err == nil {
				// We have this anime in our database, but we MIGHT be behind
				// we we will count total eps we have VS total eps gogoanime has

				// capture this anime in a variable from db
				anime, err := dbOperations.DBFetchAnimeByCol("id", malAnimeData.Entries[malIDIndex].ID)
				if err != nil {
					panic(err)
				}

				// Necessary to get a full listing of all episodes and episode links so we can extract videos
				animeId, defaultEp := scrapeAnimeID(allAnimes[animeIndex].Link)

				// Next we scrape episode lists for given anime
				episodes := scrapeEpisodeLists(animeId, defaultEp)

				log.Println(malAnimeData.Entries[malIDIndex].Title, "has", len(episodes))
				log.Println(anime.MALTitle, "has", len(anime.EpisodeList))


				// Now we count if total number of episodes is more than ours

				difference := len(episodes) - len(anime.EpisodeList)
				mirrorsSame := func() (bool) {
					// We fake check last episode URL to see if they fucking changed
					if len(episodes) <= 0 {
						return true
					}
					googleVids, vidstreamVids, mp4uploadVids, err := scrapeVideos(episodes[len(episodes) - 1].episodeURL)
					if err != nil {
						// We check if the episodes are available or if it's a fake entry
						if err.Error() == "No Videos" {
							return true
						}
					}

					var gogoMirrors []string
					var animedomMirrors []string

					// Collect remote mirror links
					for _, v := range googleVids {
						gogoMirrors = append(gogoMirrors, v.address)
					}

					for _, v := range vidstreamVids {
						gogoMirrors = append(gogoMirrors, v)
					}

					for _, v := range mp4uploadVids {
						gogoMirrors = append(gogoMirrors, v)
					}

					// Collect local mirror links
					for _, v := range anime.EpisodeList[len(anime.EpisodeList) - 1].Mirrors {
						if strings.Contains(v.Name, "HD") {
							animedomMirrors = append(animedomMirrors, v.Iframe)
						} else if strings.Contains(v.Name, "VidStream") {
							animedomMirrors = append(animedomMirrors, v.Iframe)
						} else if strings.Contains(v.Name, "Mp4Upload") {
							animedomMirrors = append(animedomMirrors, v.Iframe)
						}
					}

					sort.Strings(gogoMirrors)
					sort.Strings(animedomMirrors)

					// Test if they are same
					return testSliceEq(gogoMirrors, animedomMirrors)
				}()

				if difference <= 0 && mirrorsSame {
					log.Println(malAnimeData.Entries[malIDIndex].Title, "is already updated")
					// So this anime is updated anyway

					// We push this to recent animes table
					count++
					dbOperations.DBPushRecentAnimes(count, anime.MALID, animeEpisodeID)
					continue
				}

				// At this point anime isn't updated, we need to fetch the remnant episodes
				log.Println(malAnimeData.Entries[malIDIndex].Title, "needs to be updated")
				// We get all episode names from MAL to use the new names if there is
				malEpisodeNames := myanimelistAssistant.MalFetchEpisodelist(malAnimeData.Entries[malIDIndex].ID,
					malAnimeData.Entries[malIDIndex].Title)

				episodeList := []projectModels.StructureEpisode{}

				// episodeUnavailable for checking if episode has no vids
				episodeUnavailable := false
				for episodeIndex := 0; episodeIndex < len(episodes); episodeIndex++ {
					// Saving videos of each category in respective variables
					log.Println("Fetching for", anime.MALTitle, "episode", episodes[episodeIndex].episodeID)
					googleVids, vidstreamVids, mp4uploadVids, err := scrapeVideos(episodes[episodeIndex].episodeURL)

					if err != nil {
						// We check if the episodes are available or if it's a fake entry
						if err.Error() == "No Videos" {
							episodeUnavailable = true
							log.Println((fmt.Sprintf("No Videos for this anime? %d %s %s", episodeIndex,
								episodes[episodeIndex].episodeID, fixedAnimeName)))
							writeToFile(fileFakeEntry, fixedAnimeName)
							break
						}
					}

					thisEpisode := projectModels.StructureEpisode{}

					// Append mirrors to this episode's mirror model
					thisEpisode.Mirrors = setMirrors(googleVids, vidstreamVids, mp4uploadVids, episodeIndex, episodes)

					// Put episode Name
					thisEpisode.Name = func() (string) {
						if episodeIndex >= len(malEpisodeNames) || episodeIndex < 0 {
							return ""
						}
						return malEpisodeNames[episodeIndex]
					}()

					// Put episode ID
					thisEpisode.EpisodeID = episodes[episodeIndex].episodeID

					// Append to episode list for this anime
					//		log.Println("Inserting episode", thisEpisode.EpisodeID, "to", anime.MALTitle)
					episodeList = append(episodeList, thisEpisode)
				}
				// If certain episode vids unavailable we skip this anime
				if episodeUnavailable == true {
					log.Println("Skipping for no videos on", anime.MALTitle)
					continue
				}

				// Now if all good we update episode listing for this anime
				anime.EpisodeList = episodeList

				log.Println("Updating", anime.MALTitle)
				err = dbOperations.DBUpdateEpisodelist(anime.EpisodeList, anime.MALID)
				if err != nil {
					panic(err)
				}

				// We push this to recent animes table
				count++
				dbOperations.DBPushRecentAnimes(count, anime.MALID, animeEpisodeID)
				continue
			}
			// Finally if err != nil, which means either legit error or we don't have this anime
			if err.Error() != "Empty Result" {
				if err != nil {
					panic(err)
				}
			}

			// We will fetch the new anime

			log.Println("Gotta add the new anime", malAnimeData.Entries[malIDIndex].Title)
			// Fetching MAL details
			anime := projectModels.StructureAnime{}
			anime.MALID = malAnimeData.Entries[malIDIndex].ID
			anime.MALTitle = malAnimeData.Entries[malIDIndex].Title
			anime.MALEnglish = malAnimeData.Entries[malIDIndex].EnglishName
			anime.Genre = scrapeGenre(allAnimes[animeIndex].Link)
			anime.MALDescription = myanimelistAssistant.CleanMALDescription(malAnimeData.Entries[malIDIndex].Synopsis)
			anime.Score = func() float64 {
				i, err := strconv.ParseFloat(malAnimeData.Entries[malIDIndex].Score, 64)
				if err == nil {
					return i
				}
				return 0.0
			}()
			anime.Status = malAnimeData.Entries[malIDIndex].Status
			anime.Type = malAnimeData.Entries[malIDIndex].Type
			anime.SynonymNames = malAnimeData.Entries[malIDIndex].SynonymNames
			anime.Year = malAnimeData.Entries[malIDIndex].EndDate
			anime.Image = strings.Replace(malAnimeData.Entries[malIDIndex].Image, ".jpg", "l.jpg", 1)
			anime.Trailer = myanimelistAssistant.MalFetchTrailer(anime.MALID)
			malEpisodeNames := myanimelistAssistant.MalFetchEpisodelist(malAnimeData.Entries[malIDIndex].ID, malAnimeData.Entries[malIDIndex].Title)
			anime.Slug = slug.Slug(anime.MALTitle)

			// Necessary to get a full listing of all episodes and episode links so we can extract videos
			animeId, defaultEp := scrapeAnimeID(allAnimes[animeIndex].Link)

			// Next we scrape episode lists for given anime
			episodes := scrapeEpisodeLists(animeId, defaultEp)

			// Next we visit each episode of this anime and fetch videos
			episodeUnavailable := false
			episodeList := []projectModels.StructureEpisode{}
			for episodeIndex, v := range episodes {
				// Saving videos of each category in respective variables
				googleVids, vidstreamVids, mp4uploadVids, err := scrapeVideos(v.episodeURL)

				if err != nil {
					// We check if the episodes are available or if it's a fake entry
					if err.Error() == "No Videos" {
						episodeUnavailable = true
						log.Println((fmt.Sprintf("No Videos for this anime? %d %s %s", episodeIndex, v.episodeID, fixedAnimeName)))
						writeToFile(fileFakeEntry, fixedAnimeName)
						break
					}
				}

				thisEpisode := projectModels.StructureEpisode{}

				// Append mirrors to this episode's mirror model
				thisEpisode.Mirrors = setMirrors(googleVids, vidstreamVids, mp4uploadVids, episodeIndex, episodes)

				// Put episode Name
				thisEpisode.Name = func() (string) {
					if episodeIndex >= len(malEpisodeNames) || episodeIndex < 0 {
						return ""
					}
					return malEpisodeNames[episodeIndex]
				}()

				// Put episode ID
				thisEpisode.EpisodeID = v.episodeID

				// Append to episode list for this anime
				//		log.Println("Inserting episode", thisEpisode.EpisodeID, "to", anime.MALTitle)
				episodeList = append(episodeList, thisEpisode)
			}

			// If certain episode vids unavailable we skip this anime
			if episodeUnavailable == true {
				log.Println("Skipping for no videos on", anime.MALTitle)
				continue
			}

			// Attach episodeList to anime
			anime.EpisodeList = episodeList
			log.Println("Episode list for", anime.MALTitle, "is", anime.EpisodeList)

			// Finally fetch the anime image from MAL
			log.Println("Fetching image for", anime.MALTitle)
			myanimelistAssistant.MalFetchImage(anime.Image, anime.MALID)

			// Insert anime to rethinkdb
			err = dbOperations.DBInsertNewAnime(anime)
			if err != nil {
				panic(err)
			}

			// We push this to recent animes table
			count++
			dbOperations.DBPushRecentAnimes(count, anime.MALID, animeEpisodeID)
			log.Println("Inserted New anime successfully", anime.MALTitle)
		}
	}
	err = dbOperations.DBCopyTableTo("recent_animes", "cache_recent_animes")
	if err != nil {
		panic(err)
	}
}