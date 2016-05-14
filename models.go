package main

import "encoding/xml"

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

type structureMirrorToExtract struct {
	Name   string
	Link   string
	SubDub string
}