package main

import (
	"animedom.com/templates"
	"strings"
	"log"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/xml"
	"strconv"
	"os"
	"io"

	"github.com/PuerkitoBio/goquery"
)

/*
 malMonitorPopularAnimes() is the function where we automatically sync the popular animes from myanimelist.net sortd in
 order of popularity.
*/
func malMonitorPopularAnimes() {
	doc, err := goquery.NewDocument("http://myanimelist.net/anime/season")
	if err != nil {
		panic(err)
	}
	animelist := doc.Find(".link-image")

	// Truncate Table
	err = dbTruncateTable("popular_animes")
	if err != nil {
		panic(err)
	}

	// Using count to index(for sorting) the IDs in animedom.popular_animes.
	count := 0
	for node := range animelist.Nodes {
		res, exists := animelist.Eq(node).Attr("href")
		if !exists {
			continue
		}

		// Parsing the URL to get MAL ID.
		res = strings.Replace(res, "http://myanimelist.net/anime/", "", 1)
		for i, v := range res {
			if v == '/' {
				res = res[0:i]
				break
			}
		}

		// If it exists in our database we check if it exists in animedom.animes table.
		if err := dbCheckExistsAnimesByID(res); err != nil {
			if err.Error() != "Empty Result" {
				// There was some unexpected error
				panic(err)
			}
			continue
		}
		// Insert it to animedom.popular_animes.
		dbPushPopularAnimes(count, res)
		count++
	}
	err = dbCopyTableTo("popular_animes","cache_popular_animes")
	if err != nil{
		panic(err)
	}
}

/*
malMonitorRecentEpisodes() is the function where we automatically sync the recent animes from animeshow.tv, thereby
including any new animes if they have listed which may not exist in our database. Considering the anime exists in our db
we simply include the episode to our anime episodes listing for given anime.
*/
func malMonitorRecentEpisodes() {
	doc, err := goquery.NewDocument("http://animeshow.tv/")
	if err != nil {
		panic(err)
	}

	// Truncate Table
	err = dbTruncateTable("recent_animes")
	if err != nil {
		panic(err)
	}

	animeshowAnimelist := animeshowAnimelistPage()
	episodelist := doc.Find(".latest_episode_title")

	// Using count to index(for sorting) the IDs in animedom.recent_animes.
	count := 0
	for node := range episodelist.Nodes {
		animeName := episodelist.Eq(node).Text()


		animeFromDb, err := dbFetchAnimeByName(animeName)
		if err != nil {
			panic(err)
		}
		dbPushRecentAnimes(count, animeFromDb.MALID)
		count++

		err = dbCheckExistsAnimesByName(animeName);
		if err != nil {
			if err.Error() != "Empty Result" {
				// There was some unexpected error.
				panic(err)
			}
			// Anime does not exist in DB, so we will need to attempt to add this anime to animedom.animes table.
			log.Println("Adding", animeName, "to db.")
			// Sanity check if anime is listed on animeshow.tv list
			if _, ok := animeshowAnimelist[animeName]; !ok {
				panic("Anime episode in recent episodes but not listed?" + animeName)
			}

			// Fetching episodes list from animeshow.tv
			metaData := animeshowEpisodelistPage(animeshowAnimelist[animeName])

			// Fetching myanimelist.net data for anime
			malData, err := malAPIFetchData(animeName)
			if err != nil {
				panic(err)
			}

			entryNum := entryNumFixMAL(animeName)
			malTrailer := malFetchTrailer(malData.Entries[entryNum].ID)
			malEpisodeNames := malFetchEpisodelist(malData.Entries[entryNum].ID, malData.Entries[entryNum].Title)

			newAnime := templates.StructureAnime{}
			// Fetching score & ID & image from MAL
			newAnime.Score = func() float64 {
				i, err := strconv.ParseFloat(malData.Entries[entryNum].Score, 64)
				if err == nil {
					return i
				}
				return 0.0
			}()
			newAnime.MALID = malData.Entries[entryNum].ID
			newAnime.Image = strings.Replace(malData.Entries[entryNum].Image, ".jpg", "l.jpg", 1)
			newAnime.AnimeShowName = animeName
			newAnime.SynonymNames = malData.Entries[entryNum].SynonymNames
			newAnime.AltName = metaData.Altname
			newAnime.Genre = metaData.Genre
			newAnime.AnimeShowDescription = metaData.Description
			newAnime.Status = metaData.Status
			newAnime.Type = metaData.Type
			newAnime.Year = metaData.Year
			newAnime.Trailer = malTrailer
			newAnime.MALEnglish = malData.Entries[entryNum].EnglishName
			newAnime.MALTitle = malData.Entries[entryNum].Title
			newAnime.MALDescription = malData.Entries[entryNum].Synopsis

			tEpisodelist := []templates.StructureEpisode{}
			for tEpisodeNumber, tEpisodeLink := range metaData.Episodes {
				/* Scrape per episode link in episode listing page */
				tMirrorlistToExtract := []structureMirrorToExtract{}
				tMirrorlist := []templates.StructureMirror{}

				tempMirrorlistLinksToScrape, boolean := animeshowMirrorlistLinksToScrape(tEpisodeLink)
				if !boolean {
					continue
				}

				for _, tMirror := range tempMirrorlistLinksToScrape {
					/* Store pages to extract mirrors from */
					tMirrorlistToExtract = append(tMirrorlistToExtract, tMirror)
				}

				for i, tLinkToExtractIframe := range tMirrorlistToExtract {
					/* Iterate over stored mirror list links to fetch iframe */
					t_iframe := animeshowMirrorlistIframe(tLinkToExtractIframe.Link)
					tMirrorlist = append(tMirrorlist, templates.StructureMirror{
						Name:   tMirrorlistToExtract[i].Name,
						Iframe: t_iframe,
						SubDub: tLinkToExtractIframe.SubDub,
					})
				}

				/* Assigning episode name */
				tEpisodelist = append(tEpisodelist, templates.StructureEpisode{
					Name: func() string {
						if tEpisodeNumber >= len(malEpisodeNames) || tEpisodeNumber < 0 {
							return newAnime.AnimeShowName + " Episode " + strconv.Itoa(tEpisodeNumber + 1)
						}
						return malEpisodeNames[tEpisodeNumber]
					}(), //anime.Name + " Episode " + strconv.Itoa(t_episode_number + 1),
					Mirrors: tMirrorlist,
				})
			}
			/* Download Anime Image */
			malFetchImage(newAnime.Image, newAnime.MALID)

			newAnime.EpisodeList = tEpisodelist
			err = dbInsertNewAnime(newAnime)
			if err != nil {
				panic(err)
			}
			continue
		}

		// Now if there was no error, means we have the anime in our db, we need to include this episode
		// after checking if it is already included in animedom.animes table.

		// Get episode count, if episode count greater than db episode count we fetch the number of episodes that we are
		// short of.

		// Sanity check if anime is listed on animeshow.tv list
		if _, ok := animeshowAnimelist[animeName]; !ok {
			panic("Anime episode in recent episodes but not listed?" + animeName)
		}
		episodeslistPageData := animeshowEpisodelistPage(animeshowAnimelist[animeName])

		// Getting a list of all episode names
		// Fetching myanimelist.net data for anime
		malData, err := malAPIFetchData(animeName)
		if err != nil {
			panic(err)
		}

		entryNum := entryNumFixMAL(animeName)
		malEpisodeNames := malFetchEpisodelist(malData.Entries[entryNum].ID, malData.Entries[entryNum].Title)

		difference := len(episodeslistPageData.Episodes) - len(animeFromDb.EpisodeList)

		if difference > 0 {
			// Count on animeshow.tv is greater so we need to add any extra episodes to db
			for i := 1; i <= difference; i++ {
				// Got the link of episode on animeshow.tv starting from bottom-most

				// Helper structs for mirrors
				tMirrorlistToExtract := []structureMirrorToExtract{}
				tMirrorlist := []templates.StructureMirror{}

				tEpisodeNumber := len(animeFromDb.EpisodeList) + i - 1
				newEpisode := episodeslistPageData.Episodes[tEpisodeNumber]
				tempMirrorlistLinksToScrape, boolean := animeshowMirrorlistLinksToScrape(newEpisode)
				if !boolean {
					continue
				}
				for _, tMirror := range tempMirrorlistLinksToScrape {
					// Store pages to extract mirrors from
					tMirrorlistToExtract = append(tMirrorlistToExtract, tMirror)
				}
				for i, tLinkToExtractIframe := range tMirrorlistToExtract {
					// Iterate over stored mirror list links to fetch iframe
					tIframe := animeshowMirrorlistIframe(tLinkToExtractIframe.Link)
					tMirrorlist = append(tMirrorlist, templates.StructureMirror{
						Name:   tMirrorlistToExtract[i].Name,
						Iframe: tIframe,
						SubDub: tLinkToExtractIframe.SubDub,
					})
				}
				// Assigning episode name
				anime, err := dbFetchAnimeByName(animeName)
				if err != nil {
					panic(err)
				}
				tEpisode := templates.StructureEpisode{
					Name: func() string {
						if tEpisodeNumber >= len(malEpisodeNames) || tEpisodeNumber < 0 {
							return animeName + " Episode " + strconv.Itoa(tEpisodeNumber + 1)
						}
						return malEpisodeNames[tEpisodeNumber]
					}(),
					Mirrors: tMirrorlist,
				}
				anime.EpisodeList = append(anime.EpisodeList, tEpisode)
				// Insert Episode to animdom.animes table
				err = dbUpdateEpisodelist(anime.EpisodeList, animeName)
				if err != nil {
					panic(err)
				}
			}
		}
	}
	err = dbCopyTableTo("recent_animes","cache_recent_animes")
	if err != nil{
		panic(err)
	}
	log.Println("Recent Animes Task Finished")
}

// Borrowed snippets from my animeshow.tv_scraper project.
func animeshowAnimelistPage() map[string]string {
	doc, err := goquery.NewDocument(urlAnimeshowAnimelist)
	if err != nil {
		panic(err)
	}

	animelistMap := make(map[string]string)

	animelistDoc := doc.Find(".anime_list_result")

	for node := range animelistDoc.Nodes {
		container_doc := animelistDoc.Eq(node).Find("ul").Find("li")

		for subnode := range container_doc.Nodes {
			link, _ := container_doc.Eq(subnode).Find("a").Attr("href")
			name := container_doc.Eq(subnode).Find("a").Text()

			animelistMap[name] = link
		}
	}
	return animelistMap
}

func animeshowEpisodelistPage(url string) structureAnimeshowEpisodelist {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		panic(err)
	}

	baseDoc := doc.Find(".container.main").Find(".row").Find(".col-lg-9.col-md-9.col-sm-8.col-xs-12").Find("#main").
		Find("#anime")

	// Fetch Type
	metaAnimeDoc := baseDoc.Find(".row").Find(".col-lg-6.col-md-6.col-sm-12.col-xs-12.anime_info").
		Find(".col-lg-9.col-md-9.col-sm-9.col-xs-9")
	metaAnimeType := metaAnimeDoc.Eq(0).Text()

	// Fetch Year
	metaAnimeYear := metaAnimeDoc.Eq(1).Text()

	// Fetch Status
	metaAnimeStatus := metaAnimeDoc.Eq(2).Text()

	// Fetch Genre
	metaAnimeGenre := strings.Split(metaAnimeDoc.Eq(3).Text(), ", ")

	// Fetch Alternative Title
	metaAnimeAltname := baseDoc.Find(".row").Find(".col-lg-6.col-md-6.col-sm-12.col-xs-12.anime_info").
		Find(".alternative_titles").Find("ul").Find("li").Text()

	// Fetching description
	metaAnimeDescription, err := baseDoc.Find(".anime_discription").Html()
	if err != nil {
		panic(err)
	}

	// Fetching episode links
	var metaAnimeEpisodes []string
	episodelistDoc := baseDoc.Closest("#main").Find("#episodes_list").
		Find(".episodes_list_result")

	for episode := len(episodelistDoc.Nodes) - 1; episode >= 0; episode-- {
		link, _ := episodelistDoc.Eq(episode).Find("a").Attr("href")

		metaAnimeEpisodes = append(metaAnimeEpisodes, link)
	}

	data := structureAnimeshowEpisodelist{
		Type:        metaAnimeType,
		Year:        metaAnimeYear,
		Status:      metaAnimeStatus,
		Genre:       metaAnimeGenre,
		Altname:     metaAnimeAltname,
		Description: metaAnimeDescription,
		Episodes:    metaAnimeEpisodes,
	}

	return data
}

func animeshowMirrorlistLinksToScrape(url string) ([]structureMirrorToExtract, bool) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		if strings.Contains(err.Error(), "no Host in request URL") {
			file, err := os.OpenFile("noaccess_mirrorlist_toscrape.txt", os.O_APPEND | os.O_WRONLY, 0600)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			if _, err = file.WriteString(url + "\n"); err != nil {
				panic(err)
			}
			return []structureMirrorToExtract{}, false
		} else {
			panic(err)
		}
	}

	mirrorlistDoc := doc.Find(".col-lg-6.col-md-6.col-sm-12.col-xs-12")

	mirrorToExtractStruct := []structureMirrorToExtract{}
	defaultMirror := mirrorlistDoc.Find(".episode_mirrors_wraper.episode_mirrors_wraper_focus").Find(".episode_mirrors_name").Text()
	defaultSubdub := mirrorlistDoc.Find(".episode_mirrors_wraper.episode_mirrors_wraper_focus").Find(".episode_mirrors_type_sub").Text()
	mirrorToExtractStruct = append(mirrorToExtractStruct, structureMirrorToExtract{
		Name: defaultMirror, Link: url, SubDub: defaultSubdub,
	})

	for mirror := range mirrorlistDoc.Nodes {
		linkPath := mirrorlistDoc.Eq(mirror).Find("a")
		link, exists := linkPath.Attr("href")
		if exists {
			mirrorName := linkPath.Find(".episode_mirrors_wraper").Find(".episode_mirrors_name").Text()
			subdub := linkPath.Find(".episode_mirrors_wraper").Find(".episode_mirrors_type_sub").Text()
			mirrorToExtractStruct = append(mirrorToExtractStruct, structureMirrorToExtract{
				Name: mirrorName, Link: link, SubDub: subdub,
			})
		}
	}
	return mirrorToExtractStruct, true
}

func animeshowMirrorlistIframe(url string) string {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		panic(err)
	}

	iframe_object, err := doc.Find(".embed.embeded").First().Html()

	if err != nil {
		panic(err)
	}

	iframe_object = strings.TrimPrefix(iframe_object, " <nil>")
	return iframe_object
}

/* myanimelist.net specific helpers */

func malAPIFetchData(animeName string) (structureMALApiAnime, error) {
	client := &http.Client{}

	animeName = strings.Replace(animeNameFixMAL(animeName), " ", "+", -1)

	req, err := http.NewRequest("GET", fmt.Sprintf("http://myanimelist.net/api/anime/search.json?q=%s", animeName), nil)
	req.SetBasicAuth(malUsername, malPassword)
	resp, err := client.Do(req)
	if err != nil {
		return structureMALApiAnime{}, err
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return structureMALApiAnime{}, err
	}

	xobj := structureMALApiAnime{}
	err = xml.Unmarshal(bodyText, &xobj)
	if err != nil {
		return structureMALApiAnime{}, err
	}
	return xobj, nil
}

func malFetchEpisodelist(malID string, malTitle string) []string {
	malTitle = strings.Replace(malTitle, "%", "", -1)
	doc, err := goquery.NewDocument(fmt.Sprintf("http://myanimelist.net/anime/%s/%s/episode", malID, malTitle))
	if err != nil {
		panic(err)
	}

	var result []string
	episodeList := doc.Find(".fl-l.fw-b")

	for node := range episodeList.Nodes {
		episode := episodeList.Eq(node)
		result = append(result, episode.Text())
	}
	return result
}

func malFetchTrailer(malID string) string {
	doc, err := goquery.NewDocument(fmt.Sprintf("http://myanimelist.net/anime/%s/", malID))
	if err != nil {
		panic(err)
	}

	trailer, exists := doc.Find(".iframe.js-fancybox-video.video-unit.promotion").Attr("href")

	if !exists {
		return "Nil"
	}
	return trailer
}

func malFetchImage(path string, name string) {
	// don't worry about errors
	response, e := http.Get(path)
	if e != nil {
		panic(e)
	}

	defer response.Body.Close()

	// open a file for writing
	file, err := os.Create(fmt.Sprintf("assets/img/anime/%s.jpg", name))
	if err != nil {
		panic(err)
	}
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, response.Body)
	if err != nil {
		panic(err)
	}
	file.Close()
}

/* Random fixes for scraping/fetching via mal api */
func entryNumFixMAL(animeName string) (int) {
	var entry_num int
	if animeName == "Jinsei" {
		entry_num = 1
	} else if animeName == "Beelzebub" {
		entry_num = 1
	} else if animeName == "Million Doll" {
		entry_num = 1
	} else if animeName == "Ajin" {
		entry_num = 25
	} else if animeName == "Sengoku Musou" {
		entry_num = 1
	} else if animeName == "Amnesia" {
		entry_num = 3
	} else if animeName == "Hundred" {
		entry_num = 9
	} else if animeName == "Itoshi no Muco" {
		entry_num = 3
	} else if animeName == "Another" {
		entry_num = 20
	} else if animeName == "Onigiri" {
		entry_num = 1
	} else if animeName == "Charlotte" {
		entry_num = 1
	} else {
		entry_num = 0
	}
	return entry_num
}

func animeNameFixMAL(animeName string) (string) {
	if animeName == "Chaos Head" {
		animeName = "Chaos;Head"
	} else if animeName == "JoJo's Bizarre Adventure: Stardust Crusaders Season 2" {
		animeName = "JoJo no Kimyou na Bouken: Stardust Crusaders 2nd Season"
	} else if animeName == "Cardfight!! Vanguard G Stride Gate-hen" {
		animeName = "Cardfight!! Vanguard Third Season"
	} else if animeName == "Uta no Prince-sama Maji Love 2000%" {
		animeName = "Uta no Prince Sama 2"
	} else if animeName == "Osomatsu-san Year-End Special" {
		animeName = "Osomatsu-san Special"
	} else if animeName == "Kaitou Joker Season 2" {
		animeName = "Kaitou Joker 2nd Season"
	} else if animeName == "Saki: Nationals" {
		animeName = "Saki: The Nationals"
	} else if animeName == "Silver Spoon Season 2" {
		animeName = "Silver Spoon 2nd Season"
	} else if animeName == "Aldnoah Zero" {
		animeName = "Aldnoah.Zero"
	} else if animeName == "Aldnoah Zero Season 2" {
		animeName = "Aldnoah.Zero 2nd Season"
	} else if animeName == "Fate Zero Second Season" {
		animeName = "Fate/Zero 2nd Season"
	} else if animeName == "Re Hamatora" {
		animeName = "Re: Hamatora: Season 2"
	} else if animeName == "Futsuu no Joshikousei ga Locodol Yatte Mita Special" {
		animeName = "Futsuu no Joshikousei ga [Locodol] Yattemita.: Nagarekawa, Annai Shitemita."
	} else if animeName == "Kaitou Joker Season 3" {
		animeName = "Kaitou Joker 3rd Season"
	} else if animeName == "Selector Infected WIXOSS Specials" {
		animeName = "Selector Infected WIXOSS: Midoriko-san to Piruruku-tan"
	} else if animeName == "Gintama 2015" {
		animeName = "Gintama' (2015)"
	} else if animeName == "Ore, Twintails ni Narimasu." {
		animeName = "Gonna be the Twin-Tail!!"
	} else if animeName == "Fate/stay night: Unlimited Blade Works (TV) Season 2" {
		animeName = "Fate/stay night: Unlimited Blade Works 2nd Season"
	} else if animeName == "Futsuu no Joshikousei ga Locodol Yatte Mita" {
		animeName = "Futsuu no Joshikousei ga [Locodol] Yattemita."
	} else if animeName == "Active Raid: Kidou Kyoushuushitsu Dai Hakkei" {
		animeName = "Active Raid: Kidou Kyoushuushitsu Dai Hachi Gakari"
	} else if animeName == "Fate/kaleid liner Prisma Illya 2wei!" {
		animeName = "Fate/kaleid liner Prisma☆Illya 2wei!"
	} else if animeName == "Sailor Moon: Crystal Season III" {
		animeName = "Bishoujo Senshi Sailor Moon Crystal Season III"
	} else if animeName == "Ace of Diamond Season 2" {
		animeName = "Diamond no Ace: Second Season"
	} else if animeName == "Norn9: Norn+Nonet" {
		animeName = "Norn9"
	} else if animeName == "Ai Tenchi Muyo!" {
		animeName = "Ai Tenchi Muyou!"
	} else if animeName == "Shingeki no Bahamut - Genesis" {
		animeName = "Shingeki no Bahamut: Genesis"
	} else if animeName == "Futsuu no Joshikousei ga Locodol Yatte Mita OVA" {
		animeName = "Futsuu no Joshikousei ga [Locodol] Yattemita. OVA"
	} else if animeName == "Fate Zero" {
		animeName = "Fate/Zero"
	} else if animeName == "Hunter X Hunter 2011" {
		animeName = "Hunter x Hunter (2011)"
	} else if animeName == "Haikyu!! Second Season" {
		animeName = "Haikyuu!! Second Season"
	} else if animeName == "Sailor Moon: Crystal" {
		animeName = "Bishoujo Senshi Sailor Moon Crystal"
	} else if animeName == "Fate/stay night: Unlimited Blade Works (TV)" {
		animeName = "Fate/stay night: Unlimited Blade Works"
	} else if animeName == "Sabagebu! - Survival Game Club" {
		animeName = "Sabagebu!"
	} else if animeName == "Nisekoi Season 2" {
		animeName = "Nisekoi: False Love"
	} else if animeName == "Uta no Prince-sama: Maji Love Revolutions" {
		animeName = "Uta no Prince Sama Revolutions"
	} else if animeName == "Fairy Tail 2014" {
		animeName = "Fairy Tail (2014)"
	} else if animeName == "Oregairu Season 2" {
		animeName = "My Teen Romantic Comedy SNAFU TOO!"
	} else if animeName == "Rozen Maiden 2013" {
		animeName = "Rozen Maiden (2013)"
	} else if animeName == "Sengoku Basara Judge End" {
		animeName = "Sengoku Basara: Judge End"
	} else if animeName == "Baby Steps Season 2" {
		animeName = "Baby Steps 2nd Season"
	} else if animeName == "Tokyo Ghoul Season 2" {
		animeName = "Tokyo Ghoul 2nd Season"
	} else if animeName == "Luck & Logic" {
		animeName = "Luck and Logic"
	} else if animeName == "Kore wa Zombie Desu ka? of the Dead" {
		animeName = "Is this A Zombie? of the Dead"
	} else if animeName == "Toriko" {
		animeName = "Toriko (2011)"
	} else if animeName == "OreImo Season 2" {
		animeName = "Oreimo 2"
	} else if animeName == "Shin Strange+" {
		animeName = "Strange Plus Second Season"
	} else if animeName == "Ai Mai Mi" {
		animeName = "Choboraunyopomi Gekijou Ai Mai Mii"
	} else if animeName == "Himegoto" {
		animeName = "Secret Princess Himegoto"
	} else if animeName == "Maken-Ki! Two" {
		animeName = "Maken-Ki! Second Season"
	} else if animeName == "Highschool DxD New" {
		animeName = "High School DxD New"
	} else if animeName == "Strange+" {
		animeName = "Strange Plus"
	} else if animeName == "Shoujo-tachi wa Kouya wo Mezasu" {
		animeName = "Girls Beyond the Wasteland"
	} else if animeName == "Shounen Maid" {
		animeName = "Boy Maid"
	} else if animeName == "Hunter x Hunter: The Last Mission" {
		animeName = "Hunter x Hunter Movie: The Last Mission"
	}
	return animeName
}