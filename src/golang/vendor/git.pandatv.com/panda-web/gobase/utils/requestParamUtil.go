package utils

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
)

var EMPTY_HASH = make(map[string]interface{}, 0)
var EMPTY_ARRAY = make([]interface{}, 0)

func ParseIntParam(form url.Values, keys ...string) int {
	for _, key := range keys {
		v := form.Get(key)
		if v != "" {
			value, _ := strconv.Atoi(v)
			return value
		}
	}
	return 0
}

func ParseStringParam(form url.Values, keys ...string) string {
	for _, key := range keys {
		v := form.Get(key)
		if v != "" {
			return v
		}
	}
	return ""
}

func ParseJsonHashParam(form url.Values, keys ...string) map[string]interface{} {
	params := ParseJsonParam(form, keys...)
	if hash, ok := params.(map[string]interface{}); ok {
		return hash
	}

	return EMPTY_HASH
}

func ParseJsonArrayParam(form url.Values, keys ...string) []interface{} {
	params := ParseJsonParam(form, keys...)
	if array, ok := params.([]interface{}); ok {
		return array
	}

	return EMPTY_ARRAY
}

//return json or json array
func ParseJsonParam(form url.Values, keys ...string) interface{} {
	var params interface{}
	for _, key := range keys {
		v := form.Get(key)
		if v != "" {
			json.Unmarshal([]byte(v), &params)
			break
		}
	}
	return params
}

func ParseJsonHashParamFromPHP(form url.Values, keys ...string) map[string]interface{} {
	params := ParseJsonParamFromPHP(form, keys...)
	if hash, ok := params.(map[string]interface{}); ok {
		return hash
	}
	return EMPTY_HASH
}

func ParseJsonArrayParamFromPHP(form url.Values, keys ...string) []interface{} {
	params := ParseJsonParamFromPHP(form, keys...)
	if array, ok := params.([]interface{}); ok {
		return array
	}
	return EMPTY_ARRAY
}

func ParseJsonParamFromPHP(form url.Values, keys ...string) interface{} {
	var params = make(map[string]interface{}, 0)
	for _, key := range keys {
		prefix := key + "["
		for k, v := range form {
			if strings.HasPrefix(k, prefix) {
				keys := parsePhpArrayKeys(k)
				paramValue := parseParamValue(v[0])
				generateJson(params, paramValue, keys...)
			}
		}
		if len(params) > 0 {
			break
		}
	}
	array := parseArrayParams("", params)
	if hasValue(array) {
		return array
	}
	return params
}
func parseArrayParams(k string, params map[string]interface{}) []interface{} {
	array := make([]interface{}, len(params))
	isArray := true
	for k, v := range params {
		if nv, ok := v.(map[string]interface{}); ok {
			ar := parseArrayParams(k, nv)
			if hasValue(ar) {
				params[k] = ar
				v = ar
			}
		}
		if !isArray {
			continue
		}
		i, err := strconv.Atoi(k)
		if err != nil {
			continue
			isArray = false
		}
		if i >= len(params) {
			isArray = false
			continue
		}
		array[i] = v
	}
	if !isArray {
		return EMPTY_ARRAY
	}
	return array
}

func hasValue(value []interface{}) bool {
	for _, v := range value {
		if v != nil {
			return true
		}
	}
	return false
}

func CombineMap(source, dest map[string]interface{}) {
	for k, v := range source {
		dest[k] = v
	}
}

func parseParamValue(v string) interface{} {
	var value interface{}
	if strings.HasPrefix(v, "{") || strings.HasPrefix(v, "[") {
		json.Unmarshal([]byte(v), &value)
		if value != nil {
			return value
		}
	}

	return v
}

func generateJson(json map[string]interface{}, value interface{}, keys ...string) {

	if len(keys) == 1 {
		json[keys[0]] = value
		return
	}

	jsonKey := keys[0]

	childJson := json[jsonKey]
	if childJson == nil {
		newJson := make(map[string]interface{}, 0)
		json[jsonKey] = newJson
		generateJson(newJson, value, keys[1:]...)
		return
	}
	if j, ok := childJson.(map[string]interface{}); ok {
		generateJson(j, value, keys[1:]...)
	}
}

// php 特有数据结构 from[key1][key2].... = 1
// 解析成json 会包含多级 key
func parsePhpArrayKeys(key string) []string {
	keys := strings.Split(key, "[")[1:]
	ret := []string{}

	for _, k := range keys {
		ret = append(ret, strings.TrimSuffix(k, "]"))
	}
	return ret
}

func ParseStringParamRMPrefix(form url.Values, prefix string, keys ...string) string {
	p := ParseStringParam(form, keys...)
	if strings.HasPrefix(p, prefix) {
		p = strings.TrimPrefix(p, prefix)
	}

	return p
}
