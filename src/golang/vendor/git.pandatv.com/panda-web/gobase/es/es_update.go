package es

import (
	"sync"

	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
)

type ESUpdate struct {
	*ESClient
	docs []*UpdateDoc
	*sync.Mutex
}

type UpdateDoc struct {
	Index string
	Type  string
	Id    string
	Doc   interface{}
}

func (update *ESUpdate) UpdateDoc(index, typ, id string, doc interface{}) *ESUpdate {
	if index == "" || typ == "" {
		return update
	}
	update.Lock()
	defer update.Unlock()
	update.docs = append(update.docs, &UpdateDoc{index, typ, id, doc})
	return update
}

func (update *ESUpdate) UpdateDocs(docs ...*UpdateDoc) *ESUpdate {
	update.Lock()
	defer update.Unlock()
	update.docs = append(update.docs, docs...)
	return update
}

func (update *ESUpdate) copyData() []*UpdateDoc {
	if len(update.docs) == 0 {
		return update.docs
	}
	update.Lock()
	defer update.Unlock()
	docs := update.docs
	update.docs = make([]*UpdateDoc, 0)
	return docs
}

func (update *ESUpdate) Do() (*ESUpdateResult, error) {
	docs := update.copyData()
	if len(docs) == 0 {
		return &ESUpdateResult{Success: true}, nil
	}
	server := update.Bulk()
	bulks := []elastic.BulkableRequest{}

	for _, doc := range docs {
		if doc.Index == "" || doc.Type == "" {
			continue
		}
		if doc.Id != "" {
			bulkUpdate := elastic.NewBulkUpdateRequest().Index(doc.Index).Type(doc.Type).Id(doc.Id).Doc(doc.Doc).DocAsUpsert(true)
			bulks = append(bulks, bulkUpdate)
		} else {
			bulkIndex := elastic.NewBulkIndexRequest().Index(doc.Index).Type(doc.Type).Doc(doc.Doc)
			bulks = append(bulks, bulkIndex)
		}
	}

	server.Add(bulks...)

	resp, err := server.Do(context.Background())
	if err != nil {
		return nil, err
	}
	return &ESUpdateResult{len(resp.Succeeded()) > 0, resp.Failed()}, nil
}
