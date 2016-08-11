package gogoanimeAssistant

import (
	"errors"
	"fmt"
	"strings"

	"animedom.com/common"
	"animedom.com/projectModels"

	"github.com/PuerkitoBio/goquery"
	"github.com/Sirupsen/logrus"
)

//func CheckBadRedirLinks() {
//	sm := scrapeAnimelistPage()
//	for _, s := range sm {
//		body := common.GetPageResponse(s.Link)
//		if !common.CheckIfCFBypassed(body) || strings.TrimSpace(body) == "" {
//			common.WriteToFile("badRedirLinks", fmt.Sprintf("%s\t%s", s.Name, s.Link))
//		}
//	}
//}

func scrapeAnimelistPage() []projectModels.StructureNameAndLink {
	var animelistpage []projectModels.StructureNameAndLink
	for i := 1; i < 27; i++ {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(common.GetPageResponse(fmt.Sprintf(gogoAnimeAnimelistURL, i))))
		if err != nil {
			panic(err)
		}
		listing := doc.Find(".listing").Find("li")

		for node := range listing.Nodes {
			name := strings.TrimSpace(listing.Eq(node).Text())
			link, exists := listing.Eq(node).Find("a").Attr("href")
			if !exists {
				panic(errors.New("Link not found"))
			}
			link = strings.TrimSpace(link)
			if link == "http://gogoanime.io/category/dragon-ball-super" {
				link = "http://gogoanime.io/category/dragon-ball-super-2016"
			}
			//logrus.Debugln("Adding to list", name)
			animelistpage = append(animelistpage, projectModels.StructureNameAndLink{Name: name, Link: link})
		}
	}
	logrus.Debugln("scrapeAnimelistPage()", len(animelistpage), "animes from animelist page")
	return animelistpage
}

func scrapeAnimeID(url string) (string, string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(common.GetPageResponse(url)))
	if err != nil {
		panic(err)
	}

	animeId, exists := doc.Find("#movie_id").Attr("value")
	if !exists {
		panic(errors.New("#movie_id Value not found"))
	}
	defaultEp, exists := doc.Find("#default_ep").Attr("value")
	if !exists {
		panic(errors.New("#default_ep Value not found"))
	}
	return animeId, defaultEp
}

func scrapeEpisodeLists(id, defaultEp string) []episodelist {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(common.GetPageResponse("http://gogoanime.io/site/loadEpisode?ep_start=-1000&ep_end=2000" + "&id=" + id + "&default_ep=" + defaultEp)))
	if err != nil {
		panic(err)
	}

	episodes := doc.Find("#episode_related").Find("li")

	yepisodes := []episodelist{}

	for node := range episodes.Nodes {
		episodeLink, exists := episodes.Eq(node).Find("a").Attr("href")
		if !exists {
			panic(errors.New("Link not found"))
		}
		episodeID := strings.TrimSpace(episodes.Eq(node).Find(".name").Text())
		subdub := strings.TrimSpace(episodes.Eq(node).Find(".cate").Text())
		episodeLink = strings.TrimSpace(episodeLink)

		for i, v := range episodeID {
			if v == ' ' {
				episodeID = episodeID[i+1:]
				break
			}
		}

		//	log.Println("Got:", episodeID, subdub, episodeLink)
		yepisodes = append(yepisodes, episodelist{episodeID: episodeID, episodeURL: episodeLink, subdub: subdub})
	}

	// Straighten up that episode list
	for i, j := 0, len(yepisodes)-1; i < j; i, j = i+1, j-1 {
		yepisodes[i], yepisodes[j] = yepisodes[j], yepisodes[i]
	}

	return yepisodes
}

//func scrapeVideos(url string) ([]googleVid, []string, []string, error) {
//	doc, err := goquery.NewDocumentFromReader(strings.NewReader(common.GetPageResponse(url)))
//	if err != nil {
//		panic(err)
//	}
//	//videos := doc.Find(".anime_video_body_watch_items")
//
//	// Checking if Google based HD videos exist
//	//m18 is 360, m22 is 720, m37 is 1080
//	googledata := []googleVid{}
//	_, exists := doc.Find("option").Attr("value")
//	if exists {
//		googleVids := doc.Find("option")
//		for node := range googleVids.Nodes {
//			videoAddress, exists := googleVids.Eq(node).Attr("value")
//			if !exists {
//				continue
//			}
//			desc := googleVids.Eq(node).Text()
//			googledata = append(googledata, googleVid{address: videoAddress, quality: desc})
//		}
//	}
//
//	// Checking and storing if vidstream exists
//	var vidstreamdata []string
//	divScanner := doc.Find("div")
//	for node := range divScanner.Nodes {
//		element := divScanner.Eq(node)
//		styleValue, _ := element.Attr("style")
//		if styleValue == "background-color:#000" {
//			iframe, exists := element.Find("iframe").Attr("src")
//			if exists {
//				for i := 0; i < len(iframe); i++ {
//					if iframe[i] == '&' {
//						iframe = iframe[0:i]
//						break
//					}
//				}
//				vidstreamdata = append(vidstreamdata, iframe)
//			}
//		}
//	}
//
//	var mp4data []string
//	mp4uploadScanner := doc.Find(".anime_video_body_watch_items")
//	for node := range mp4uploadScanner.Nodes {
//		mp4Id, exists := mp4uploadScanner.Eq(node).Find(".ads_iframe").Attr("link-watch")
//		if exists {
//			mp4data = append(mp4data, mp4Id)
//		}
//	}
//
//	if len(mp4data) == 0 && len(vidstreamdata) == 0 && len(googledata) == 0 {
//		return []googleVid{}, []string{}, []string{}, errors.New("No Videos")
//	}
//
//	return googledata, vidstreamdata, mp4data, nil
//}
