package dbModels

// Helper structs for scraping from myanimeshow.net
type StructureAnimeshowEpisodelist struct {
	Type        string
	Year        string
	Status      string
	Genre       []string
	Altname     string
	Description string
	Episodes    []string
}

type StructureMirrorToExtract struct {
	Name   string
	Link   string
	SubDub string
}

