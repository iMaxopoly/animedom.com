package main

import "encoding/xml"

type StructurePopularAnimes struct {
	MALID string  `json:"MAL ID" storm:"id" gorethink:"id"`
}

/* myanimelist.net models */
type structureMALApiAnime struct {
	Anime   xml.Name               `xml:"anime"`
	Entries []structureMALApiEntry `xml:"entry"`
}

type structureMALApiEntry struct {
	ID            string `xml:"id"`
	Title         string `xml:"title"`
	EnglishName   string `xml:"english"`
	SynonymNames  string `xml:"synonyms"` // Semi-colon+space seperated
	EpisodesCount string `xml:"episodes"`
	Score         string `xml:"score"`
	Type          string `xml:"type"`
	Status        string `xml:"status"`
	StartDate     string `xml:"start_date"`
	EndDate       string `xml:"end_date"`
	Synopsis      string `xml:"synopsis"`
	Image         string `xml:"image"`
}

// Helper structs for scraping from myanimeshow.net
type structureAnimeshowEpisodelist struct {
	Type        string
	Year        string
	Status      string
	Genre       []string
	Altname     string
	Description string
	Episodes    []string
}

type structureEpisode struct {
	Name    string            `json:"Episode Name" storm:"index" gorethink:"Name"`
	Mirrors []structureMirror `json:"Episode Mirrors" storm:"index" gorethink:"Mirrors"`
}

type structureMirror struct {
	Name   string `json:"Mirror Name" storm:"index" gorethink:"Name"`
	Iframe string `json:"Embed Code" storm:"index" gorethink:"Iframe"`
	SubDub string `json:"SubDub" storm:"index" gorethink:"SubDub"`
}

type structureMirrorToExtract struct {
	Name   string
	Link   string
	SubDub string
}