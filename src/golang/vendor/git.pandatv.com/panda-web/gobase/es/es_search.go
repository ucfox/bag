package es

import (
	"strings"

	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
)

type ESSearch struct {
	*ESClient
	indexs    []string
	types     []string
	fields    []string
	query     elastic.Query
	aggs      []elastic.Aggregation
	aggsField []string
	scrollId  string
	sort      []elastic.Sorter
	from      *int
	size      *int
	distinct  bool
}

func (search *ESSearch) Index(index string) *ESSearch {
	if index != "" {
		search.indexs = strings.Split(index, ",")
	}

	return search
}

func (search *ESSearch) Type(typ string) *ESSearch {
	if typ != "" {
		search.types = strings.Split(typ, ",")
	}
	return search
}

func (search *ESSearch) Query(query string) *ESSearch {
	search.query = parseQueryCondition(query)
	return search
}

func (search *ESSearch) ScrollId(scrollId string) *ESSearch {
	search.scrollId = scrollId
	return search
}

func (search *ESSearch) GroupBy(groupBy, distinctField string, count int) *ESSearch {
	aggs, distinct := parseTermAggs(groupBy, distinctField, count)
	if aggs != nil {
		search.aggs = append(search.aggs, aggs)
		search.aggsField = append(search.aggsField, groupBy)
		if distinct {
			search.distinct = true
		}
	}
	return search
}

func (search *ESSearch) Distinct(distinctField string) *ESSearch {
	aggs := parseCardinalityAggs(distinctField)
	if aggs != nil {
		search.aggs = append(search.aggs, aggs)
		search.distinct = true
	}

	return search
}

func (search *ESSearch) DateHistogram(field, interval, distinctField string) *ESSearch {
	aggs, distinct := parseDateHistogramAggs(field, interval, distinctField)
	if aggs != nil {
		search.aggs = append(search.aggs, aggs)
		search.aggsField = append(search.aggsField, field)
		if distinct {
			search.distinct = true
		}
	}
	return search
}

func (search *ESSearch) Sort(sort string) *ESSearch {
	search.sort = parseSort(sort)
	return search
}

func (search *ESSearch) Field(field string) *ESSearch {
	if field != "" {
		search.fields = strings.Split(field, ",")
	}
	return search
}

func (search *ESSearch) Size(size int) *ESSearch {
	search.size = &size
	return search
}

func (search *ESSearch) From(from int) *ESSearch {
	search.from = &from
	return search
}

func (search *ESSearch) parseSearchSource() *elastic.SearchSource {
	searchSource := elastic.NewSearchSource()

	assembleESQuery(searchSource, search)
	assembleESField(searchSource, search)
	assembleESAggs(searchSource, search)
	assembleESSort(searchSource, search)

	if search.from != nil {
		searchSource.From(*search.from)
	}
	if search.size != nil {
		searchSource.Size(*search.size)
	}
	return searchSource
}

func (search *ESSearch) ScrollDo() (*ESSearchResult, error) {
	scrollService := search.Scroll()
	scrollService.Scroll(ES_SCROLL_TIMEOUT)
	if search.scrollId != "" {
		searchResult, err := scrollService.ScrollId(search.scrollId).Do(context.Background())
		return parseQuerySearchResult(searchResult), err
	}
	if len(search.indexs) != 0 {
		scrollService.Index(search.indexs...)
	}

	if len(search.types) != 0 {
		scrollService.Type(search.types...)
	}
	searchSource := search.parseSearchSource()

	searchResult, err := scrollService.SearchSource(searchSource).Do(context.Background())

	return parseQuerySearchResult(searchResult), err
}

func (search *ESSearch) Do() (*ESSearchResult, error) {
	searchService := search.Search()
	if len(search.indexs) != 0 {
		searchService.Index(search.indexs...)
	}

	if len(search.types) != 0 {
		searchService.Type(search.types...)
	}

	searchSource := search.parseSearchSource()
	searchResult, err := searchService.SearchSource(searchSource).Do(context.Background())

	esResult := parseQuerySearchResult(searchResult)

	if searchResult != nil && searchResult.Aggregations != nil {
		bucketItem, _ := searchResult.Aggregations.Terms(ES_AGGS_NAME)
		fields := search.aggsField
		if len(fields) > 0 {
			fields = fields[1:]
		}
		esResult.Bucket = parseGroupByResult(bucketItem, fields, search.distinct)
	}
	return esResult, err
}
