package es

import (
	"context"
	"strings"
	"sync"

	"gopkg.in/olivere/elastic.v5"
)

type ESClient struct {
	*elastic.Client
}

const (
	HTTP_PREFIX       = "http://"
	ES_SCROLL_TIMEOUT = "1m" // 1 minutes
	ES_AGGS_NAME      = "aggs"
)

func NewESClient(url ...string) (*ESClient, error) {
	urls := []string{}
	for _, u := range url {
		if !strings.HasPrefix(u, HTTP_PREFIX) {
			u = HTTP_PREFIX + u
		}
		urls = append(urls, u)
	}
	urlFuc := elastic.SetURL(urls...)
	decoder := elastic.SetDecoder(&EsDecoder{})
	client, err := elastic.NewClient(urlFuc, decoder)
	if err != nil {
		return nil, err
	}

	return &ESClient{Client: client}, nil
}

func (client *ESClient) NewSearch() *ESSearch {
	return &ESSearch{ESClient: client}
}

func (client *ESClient) NewUpdate() *ESUpdate {
	return &ESUpdate{ESClient: client, docs: make([]*UpdateDoc, 0), Mutex: &sync.Mutex{}}
}

func (client *ESClient) DeleteDoc(index, typ, id string) error {
	_, err := client.Delete().Index(index).Type(typ).Id(id).Do(context.Background())
	return err
}
