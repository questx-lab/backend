package search

import (
	"context"
	"errors"
	"path"

	"github.com/blevesearch/bleve/v2"
	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/xcontext"
)

const (
	CommunityDoc = "community"
	QuestDoc     = "quest"
	TemplateDoc  = "template"
)

type CommunityData struct {
	Handle       string
	DisplayName  string
	Introduction string
}

type QuestData struct {
	Title       string
	Description string
}

type TemplateData struct {
	Title       string
	Description string
}

type bleveIndex struct {
	logger   logger.Logger
	indexDir string
	indexes  *xsync.MapOf[string, bleve.Index]
}

func NewBleveIndex(ctx context.Context) *bleveIndex {
	return &bleveIndex{
		logger:   xcontext.Logger(ctx),
		indexDir: xcontext.Configs(ctx).SearchServer.IndexDir,
		indexes:  xsync.NewMapOf[bleve.Index](),
	}
}

func (i *bleveIndex) Index(document, id string, data any) error {
	index, err := i.getIndexByDocument(document)
	if err != nil {
		return err
	}

	record, err := index.Document(id)
	if err != nil {
		return err
	}

	// Delete if the record existed.
	if record != nil {
		if err := index.Delete(id); err != nil {
			return err
		}
	}

	return index.Index(id, data)
}

func (i *bleveIndex) Delete(document, id string) error {
	index, err := i.getIndexByDocument(document)
	if err != nil {
		return err
	}

	return index.Delete(id)
}

func (i *bleveIndex) Search(document, query string, offset, limit int) ([]string, error) {
	index, err := i.getIndexByDocument(document)
	if err != nil {
		return nil, err
	}

	req := bleve.NewSearchRequestOptions(bleve.NewMatchQuery(query), limit, offset, false)
	searchResults, err := index.Search(req)
	if err != nil {
		return nil, err
	}

	ids := []string{}
	for _, match := range searchResults.Hits {
		ids = append(ids, match.ID)
	}

	return ids, nil
}

func (i *bleveIndex) Close() {
	i.logger.Infof("Closing all indexers...")

	i.indexes.Range(func(document string, index bleve.Index) bool {
		if err := index.Close(); err != nil {
			i.logger.Errorf("Cannot close indexer %s: %v", document, err)
		}

		return true
	})

	i.logger.Infof("Closing all indexers...done")
}

func (i *bleveIndex) getIndexByDocument(document string) (bleve.Index, error) {
	index, ok := i.indexes.Load(document)
	if !ok {
		i.logger.Infof("A new document index is added: %s", document)

		var err error
		indexPath := path.Join(i.indexDir, string(document))
		index, err = bleve.New(indexPath, bleve.NewIndexMapping())
		if err != nil {
			if !errors.Is(err, bleve.ErrorIndexPathExists) {
				return nil, err
			}

			index, err = bleve.Open(indexPath)
			if err != nil {
				return nil, err
			}
		}

		i.indexes.Store(document, index)
	}

	return index, nil
}
