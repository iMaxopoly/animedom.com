package gogoanimeAssistant

import "animedom.com/projectModels"

//var mp4Order chan string

// Helpers
type episodelist struct {
	episodeID  string
	episodeURL string
	subdub     string
}

type googleMirror struct {
	Name         string
	Quality      string
	URL          string
	MirrorOrigin string
}

// MonitorRecentEpisodes
type recentAnime struct {
	malID        string
	episodeIndex int
}

type titleAndID struct {
	title string
	id    string
}

// mainscrape helper
type job struct {
	whatKind  string
	allAnimes []projectModels.StructureNameAndLink
	// MonitorRecentEpisodes
	animeTitle     string
	animeEpisodeID string
	// RefreshAnimes
	start int
	end   int
}

type jobResult struct {
	// MonitorRecentEpisodes
	fileWrite   projectModels.FileWrite
	recentAnime recentAnime
	error       error
}
