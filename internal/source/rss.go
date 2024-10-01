package source

import "telbot/internal/model"

type RSSSource struct {
	URL        string
	SourceID   int64
	SourceName string
}

func NewRSSSourceFromModel(m model.Source) RSSSource {
	return RSSSource{
		URL:        m.FeedURL,
		SourceID:   m.ID,
		SourceName: m.Name,
	}
}
