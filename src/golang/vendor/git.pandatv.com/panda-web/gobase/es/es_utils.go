package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"git.pandatv.com/panda-web/gobase/log"
	"gopkg.in/olivere/elastic.v5"
)

const (
	NOT_FLAG         string = "-"
	OR_FLAG                 = "|"
	MUST_FLAG               = "+"
	RANGE_FLAG              = "<>"
	SHOULD_FLAG             = "||"
	AND_FLAG                = "&&"
	KEY_VALUE_FLAG          = ":"
	ES_DISTINCT_FLAG        = "distinct"
	ES_GROUPBY_FLAG         = "group_by_state"

	ES_DISTINCT_PRECISION = 3000
)

func parseQuerySearchResult(searchResult *elastic.SearchResult) *ESSearchResult {
	if searchResult == nil || searchResult.Hits == nil {
		return nil
	}
	totalCount := getTotalCount(searchResult)
	result := make([]map[string]interface{}, 0)
	for _, hit := range searchResult.Hits.Hits {
		var source map[string]interface{}
		if hit.Source != nil {
			decode := json.NewDecoder(bytes.NewReader(*hit.Source))
			decode.UseNumber()
			decode.Decode(&source)
		} else if hit.Fields != nil {
			source = hit.Fields
		} else {
			continue
		}
		result = append(result, source)
	}
	scrollId := searchResult.ScrollId
	return &ESSearchResult{TotalNum: totalCount, ScrollId: scrollId, Source: result}
}

func parseGroupByResult(bucket *elastic.AggregationBucketKeyItems, fields []string, uniq bool) map[string]interface{} {
	result := make(map[string]interface{})
	if bucket == nil || len(bucket.Buckets) == 0 {
		return result
	}
	childKey := ""
	if len(fields) > 0 {
		childKey = fields[0]
		fields = fields[1:]
	}

	for _, aggs := range bucket.Buckets {
		count := aggs.DocCount
		if uniq {
			distValue, _ := aggs.ValueCount(ES_DISTINCT_FLAG)
			count = int64(*distValue.Value)
		}
		result[aggs.KeyNumber.String()] = count
		b, _ := aggs.Terms(ES_GROUPBY_FLAG)
		child := parseGroupByResult(b, fields, uniq)
		if len(child) != 0 {
			key := fmt.Sprintf("%v_%s", aggs.KeyNumber.String(), childKey)
			result[key] = child
		}
	}

	return result
}

func parseCardinalityAggs(distinctField string) elastic.Aggregation {
	if distinctField != "" {
		return elastic.NewCardinalityAggregation().Field(distinctField).PrecisionThreshold(ES_DISTINCT_PRECISION)
	}
	return nil
}

func parseTermAggs(field, distinctField string, size int) (elastic.Aggregation, bool) {
	if field == "" {
		return nil, false
	}
	termAggs := elastic.NewTermsAggregation().Field(field).Size(size).ShardSize(ES_DISTINCT_PRECISION)
	card := parseCardinalityAggs(distinctField)
	if card != nil {
		termAggs = termAggs.SubAggregation(ES_DISTINCT_FLAG, card)
	}

	return termAggs, card != nil
}

func parseDateHistogramAggs(field, interval, distinctField string) (elastic.Aggregation, bool) {
	if field == "" {
		return nil, false
	}

	dateAggs := elastic.NewDateHistogramAggregation().Field(field).Interval(interval)
	card := parseCardinalityAggs(distinctField)
	if card != nil {
		dateAggs = dateAggs.SubAggregation(ES_GROUPBY_FLAG, card)
	}

	return dateAggs, card != nil
}

func parseQueryCondition(query string) elastic.Query {
	boolQuery := elastic.NewBoolQuery()
	orConditions := strings.Split(query, SHOULD_FLAG)
	if len(orConditions) == 1 {
		return parseAndQueryConditon(orConditions[0])
	}
	for _, cond := range orConditions {
		boolQuery.Should(parseAndQueryConditon(cond))
	}
	return boolQuery
}
func parseAndQueryConditon(andQuery string) elastic.Query {
	andConditions := strings.Split(andQuery, AND_FLAG)
	if len(andConditions) == 1 {
		return parseFieldsQueryConditon(andConditions[0])
	}
	andBoolQuery := elastic.NewBoolQuery()
	for _, ad := range andConditions {
		andBoolQuery.Must(parseFieldsQueryConditon(ad))
	}

	return andBoolQuery
}

func parseBaseCondition(condition string) elastic.Query {
	if condition == "" {
		return nil
	}
	kv := strings.Split(condition, KEY_VALUE_FLAG)
	if len(kv) == 1 {
		return elastic.NewQueryStringQuery(kv[0]).AutoGeneratePhraseQueries(true)
	}

	return elastic.NewQueryStringQuery(kv[1]).Field(kv[0]).AutoGeneratePhraseQueries(true)
}
func parseRangeCondition(field string) *elastic.RangeQuery {
	splits := strings.Split(field, KEY_VALUE_FLAG)
	if len(splits) < 2 {
		return nil
	}
	name := splits[0]
	rangeQuery := elastic.NewRangeQuery(name)
	flag := false
	for _, v := range splits[1:] {
		if strings.HasPrefix(v, ">=") {
			flag = true
			rangeQuery.Gte(strings.TrimPrefix(v, ">="))
		} else if strings.HasPrefix(v, ">") {
			flag = true
			rangeQuery.Gt(strings.TrimPrefix(v, ">"))
		} else if strings.HasPrefix(v, "<=") {
			flag = true
			rangeQuery.Lte(strings.TrimPrefix(v, "<="))
		} else if strings.HasPrefix(v, "<") {
			flag = true
			rangeQuery.Lt(strings.TrimPrefix(v, "<"))
		}
	}

	if !flag {
		return nil
	}

	return rangeQuery
}
func parseFieldsQueryConditon(fieldsQuery string) elastic.Query {
	querys := strings.Split(fieldsQuery, ",")
	if len(querys) == 1 {
		flag, field := parseFieldFlag(querys[0])
		switch flag {
		case RANGE_FLAG:
			return parseRangeCondition(field)
		case MUST_FLAG:
			return parseBaseCondition(field)
		}
	}
	boolQuery := elastic.NewBoolQuery()
	for _, query := range querys {
		if query == "" {
			continue
		}
		flag, field := parseFieldFlag(query)
		switch flag {
		case MUST_FLAG:
			boolQuery.Must(parseBaseCondition(field))
		case OR_FLAG:
			boolQuery.Should(parseBaseCondition(field))
			boolQuery.MinimumShouldMatch("1")
		case NOT_FLAG:
			boolQuery.MustNot(parseBaseCondition(field))
		case RANGE_FLAG:
			rangeQuery := parseRangeCondition(field)
			if rangeQuery != nil {
				boolQuery.Must(rangeQuery)
			}
		}
	}

	return boolQuery
}

func parseFieldFlag(query string) (string, string) {
	if strings.HasPrefix(query, NOT_FLAG) {
		return NOT_FLAG, strings.TrimPrefix(query, NOT_FLAG)
	}
	if strings.HasPrefix(query, OR_FLAG) {
		return OR_FLAG, strings.TrimPrefix(query, OR_FLAG)
	}
	if strings.HasPrefix(query, RANGE_FLAG) {
		return RANGE_FLAG, strings.TrimPrefix(query, RANGE_FLAG)
	}
	return MUST_FLAG, strings.TrimPrefix(query, MUST_FLAG)
}
func parseSort(sort string) (sorts []elastic.Sorter) {
	if sort == "" {
		return
	}
	ss := strings.Split(sort, ",")
	for _, fs := range ss {
		fsSplit := strings.Split(fs, KEY_VALUE_FLAG)
		fieldSort := elastic.NewFieldSort(fsSplit[0])
		if len(fsSplit) > 1 {
			if fsSplit[1] == "asc" {
				fieldSort.Asc()
			} else {
				fieldSort.Desc()
			}
		}

		sorts = append(sorts, fieldSort)
	}
	return
}

func assembleESQuery(searchSource *elastic.SearchSource, search *ESSearch) {
	if search.query == nil {
		return
	}
	if logkit.IsDebug() {
		s, _ := search.query.Source()
		d, _ := json.Marshal(s)
		logkit.Debugf("query is %s", d)
	}

	searchSource.Query(search.query)
	search.query = nil
}

func assembleESField(searchSource *elastic.SearchSource, search *ESSearch) {
	if len(search.fields) == 0 {
		return
	}
	if logkit.IsDebug() {
		logkit.Debugf("fields is %s", search.fields)
	}

	sourceContext := elastic.NewFetchSourceContext(true).Include(search.fields...)
	searchSource.FetchSourceContext(sourceContext)
	search.fields = nil
}

func assembleESAggs(searchSource *elastic.SearchSource, search *ESSearch) {
	if len(search.aggs) == 0 {
		return
	}
	aggsArr := search.aggs
	aggs := aggsArr[len(aggsArr)-1]
	for i := len(aggsArr) - 2; i >= 0; i-- {
		if term, ok := aggsArr[i].(*elastic.TermsAggregation); ok {
			aggs = term.SubAggregation(ES_GROUPBY_FLAG, aggs)
		} else if date, ok := aggsArr[i].(*elastic.DateHistogramAggregation); ok {
			aggs = date.SubAggregation(ES_GROUPBY_FLAG, aggs)
		}
	}
	if logkit.IsDebug() {
		s, _ := aggs.Source()
		d, _ := json.Marshal(s)
		logkit.Debugf("aggs is %s", d)
	}
	search.aggs = nil
	searchSource.Aggregation(ES_AGGS_NAME, aggs)
}

func assembleESSort(searchSource *elastic.SearchSource, search *ESSearch) {
	if len(search.sort) == 0 {
		return
	}
	if logkit.IsDebug() {
		d, _ := json.Marshal(search.sort)
		logkit.Debugf("sorts is %s", d)
	}
	searchSource.SortBy(search.sort...)
	search.sort = nil
}

func getTotalCount(result *elastic.SearchResult) int64 {
	disValue, _ := result.Aggregations.ValueCount(ES_AGGS_NAME)
	if disValue != nil && disValue.Value != nil {
		return int64(*disValue.Value)
	}

	return result.Hits.TotalHits

}
