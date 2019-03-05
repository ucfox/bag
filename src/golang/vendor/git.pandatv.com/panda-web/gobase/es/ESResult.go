package es

type ESSearchResult struct {
	TotalNum int64                    `json:total_num`
	ScrollId string                   `json:"scroll_id,omitempty"`
	Source   []map[string]interface{} `json:"source,omitempty"`
	Bucket   map[string]interface{}   `json:"bucket,omitempty"`
}

type ESUpdateResult struct {
	Success bool
	Msg     interface{}
}
