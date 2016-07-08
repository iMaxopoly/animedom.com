package dbOperations

import "encoding/xml"

// Search related
type structureSearchJSON struct {
	MALID         string `json:"url" gorethink:"id"`
	AnimeShowName string `json:"title" gorethink:"AnimeShowName"`
}