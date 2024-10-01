package fetcher

import (
	"context"
	"log"
	"strings"
	"sync"
	set "telbot/external/datatypes"
	"telbot/internal/model"
	rss "telbot/internal/source"
	"time"
)

type ArticleStorage interface {
	Store(ctx context.Context, article model.Article) error
}

type SourceStorage interface {
	Sources(ctx context.Context) ([]model.Source, error)
}

type Source interface {
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]model.Item, error)
}

type Fetcher struct {
	articles       ArticleStorage
	sources        SourceStorage
	fetchInterval  time.Duration
	filterKeywords []string
}

func New(
	articles ArticleStorage,
	sources SourceStorage,
	fetchInterval time.Duration,
	filterKeywords []string) *Fetcher {
	return &Fetcher{
		articles:       articles,
		sources:        sources,
		fetchInterval:  fetchInterval,
		filterKeywords: filterKeywords,
	}
}

func (f *Fetcher) Fetch(ctx context.Context) error {
	sources, err := f.sources.Sources(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, source := range sources {
		wg.Add(1)

		rssSource := rss.NewRSSSourceFromModel(source)

		go func(source Source) {
			defer wg.Done()

			items, err := source.Fetch(ctx)
			if err != nil {
				log.Printf("[ERROR] Fetching items from source %s: %v", source.Name(), err)
				return
			}

			if err := f.processItems(ctx, source, items); err != nil {
				log.Printf("[ERROR] Processing items from source %s: %v", source.Name(), err)
				return
			}
		}(rssSource)
	}

	wg.Wait()

	return nil
}

func (f *Fetcher) processItems(ctx context.Context, source Source, items []model.Item) error {

	for _, item := range items {
		item.Date = item.Date.UTC()

		if f.filterItem(item) {
			continue
		}

		if err := f.articles.Store(ctx, model.Article{
			SourceID:    source.ID(),
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			PublishedAt: item.Date,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (f *Fetcher) filterItem(item model.Item) bool {
	categoriesSet := set.New(item.Categories...)
	for _, keyword := range f.filterKeywords {
		titleContainsKeyword := strings.Contains(strings.ToLower(item.Title), keyword)
		if categoriesSet.Has(keyword) || titleContainsKeyword {
			return true
		}
	}
	return false
}
