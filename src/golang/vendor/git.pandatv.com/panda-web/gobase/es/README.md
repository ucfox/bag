# **ESClient 使用说明**
    本封装的目的是最大化简化ES的使用成本
    备注：需要对数据在ES中存储的基本结构有个简单了解

## 一：初始化：
1. 将数据包  引入 到项目中。
2. 用ES集群的url（包括ip和端口）初始化ESclient 即可

## 二：方法说明：

#### Index ： 查询索引，不指定查全部数据
#### Type：查询索引类型（可不指定）
#### Query ： 查询条件（备注会有说明使用方式）
#### ScrollId ： 循环查询Id 用于查询数据多于10000的翻页条件
#### GroupBy ： 循环查询Id 用于查询数据多于10000的翻页条件
* groupBy: 分组字段
* distinctField: 去重字段（不去重，不指定）
* count: 返回的分组个数
#### Distinct 去重
#### DateHistogram 根据时间分组
* field: 分组字段
* interval: 分组间隔（支持时间格式的缩写）
* distinctField: 去重字段（不去重，不指定）
#### Sort: 排序
#### Field：返回指定字段的搜索结果

#### Do 搜索执行指令 （必须执行之一）
#### ScrollDo 循环搜索指令（和上面的必须指定其中一个，否则搜索结果不执行）

## 参数使用说明

- query：查询条件
    - **只有查询条件** 如 query＝测试 ，则进行全索引模糊匹配（效率低） 。备注：如果查询条件里面包含`空格` 会被`分词`成两个条件的 ’或’ 条件查询
    - **指定es索引字段查询** 如 query＝action:bind ,则只对es中的action 字段进行匹配查询（效率高）
    - **对指定字段的 ‘与’ ‘或’ ‘非’ 查询** 如 query＝|action:bind(或) , query=-action:bind(非) ，其他如 query＝+action:bind， query=action:bind均为 与 条件查询  多条件之间 通过 ‘,’ 分割
    - **指定字段的 范围查询** 如 query=<>count:>10:<100 即为查询 count 在 10 到100 之间的数据，需要包含那一端数据，则在相应端加 ＝
    - **复杂条件的组合查询** 如 A and B 或者 C and D 通过 ‘||’ 组合，  A or B 并且 C or D 用 ‘&&’ 组合
- sort ： 排序规则 默认 根据查询条件的相关度 进行排序 。 拼装规则 ： es 索引的字段名 加 排序 方式（升序 asc ，降序 desc ）如 sort＝count:desc 多个排序规则 以 ‘,’ 分割

## 备注
- 翻页查询： 由于 es 本身的限制 ，通过page 和 size 最大只能查询 10000个数据，如果需要 查询更多只能通过 scroll 的方式来获取，scrollId 会在系统里面 记录 你上一次 查询的条件，并继续查询。每次都会生成新的 scrollId ，没有数据 则不再生成scrollId ，并返回 EOF 错误
- groupBy 查询时，field 为 es 的索引字段，如果索引 有默认的索引模版，在 string 类型的字段 后面 需要添加 .raw 字段

## 使用样例

### 初始化客户端

> client, err := es.NewESClient(urls...) 

### 以 riven 数据统计进行展示
* 查询在线总人数
    
```
必要条件:
query＝"-fields.alived:false,action:bind"
index="logstash-kafka-riven-gateway-log-2017.03.13,logstash-kafka-riven-gateway-log-2017.03.14"

方法调用
result, err := client.NewSearch().Index(index).Query(query).Do()

转化成 ES 的语句
{
  "query": {
    "bool": {
      "must": {
        "query_string": {
          "auto_generate_phrase_queries": true,
          "fields": [
            "action"
          ],
          "query": "bind"
        }
      },
      "must_not": {
        "query_string": {
          "auto_generate_phrase_queries": true,
          "fields": [
            "fields.alived"
          ],
          "query": "false"
        }
      }
    }
  },
  "from": 0,
  "size": 0
}
result 数据: 
{
	"total_num":7
}

``` 


* 查询在线登录人数

```
必要条件
query="-fields.alived:false,action:bind,<>fields.uid:>0"
index="logstash-kafka-riven-gateway-log-2017.03.13,logstash-kafka-riven-gateway-log-2017.03.14"
from=0
size=0

方法调用
result, err := client.NewSearch().Index(index).Query(query).From(from).Size(size).Do()

转化成 ES 的语句
{
  "query": {
    "bool": {
      "must": [
        {
          "query_string": {
            "auto_generate_phrase_queries": true,
            "fields": [
              "action"
            ],
            "query": "bind"
          }
        },
        {
          "range": {
            "fields.uid": {
              "from": "0",
              "include_lower": false,
              "include_upper": true,
              "to": null
            }
          }
        }
      ],
      "must_not": {
        "query_string": {
          "auto_generate_phrase_queries": true,
          "fields": [
            "fields.alived"
          ],
          "query": "false"
        }
      }
    }
  },
  "from": 0,
  "size": 0
}

result 数据:
{
	"total_num":7
}

```

* 查询各机房在线人数

```
必要条件
query＝"-fields.alived:false,action:bind"
index="logstash-kafka-riven-gateway-log-2017.03.13,logstash-kafka-riven-gateway-log-2017.03.14"
fields="fields.idc_id.raw"
count=1000 //分组返回的个数,根据需要设定。 必须指定
distinctField="" //需要对结果进行去重的字段，不需要去重置空
from=0
size=0

方法调用
result, err := client.NewSearch().Index(index).Query(query).GroupBy(field,distinctField,count).From(from).Size(size).Do()

转化成 ES 的语句
{
  "query": {
    "bool": {
      "must": {
        "query_string": {
          "auto_generate_phrase_queries": true,
          "fields": [
            "action"
          ],
          "query": "bind"
        }
      },
      "must_not": {
        "query_string": {
          "auto_generate_phrase_queries": true,
          "fields": [
            "fields.alived"
          ],
          "query": "false"
        }
      }
    }
  },
  "from": 0,
  "size": 0,
  "aggs": {
    "aggs": {
      "terms": {
        "field": "fields.idc_id.raw",
        "size": 1000
      }
    }
  }
}

result 数据

{
	"total_num":7,
	"bjac":7
}
```

* 查询在2017-03-14 19:00:00 之前每个小时各机房的上线人数

```
必要条件
query＝"-fields.alived:false,action:bind,<>fields.bind_time:<1489489200000"
index="logstash-kafka-riven-gateway-log-2017.03.13,logstash-kafka-riven-gateway-log-2017.03.14"
groupByFields="fields.idc_id.raw"
count=1000 //分组返回的个数,根据需要设定。 必须指定
distinctField="" //需要对结果进行去重的字段，不需要去重置空
dateHistogramField="fields.bind_time" // 时间分组字段
interval="1h"
from=0
size=0

方法调用
result, err := client.NewSearch().Index(index).Query(query).DateHistogram(dateHistogramField,interval,distinctField).GroupBy(groupByFields,distinctField,count).From(from).Size(size).Do()  // 注意 DateHistogram 和  GroupBy的顺序 。 类似于 mysql 的group by 顺序决定了分组模式

转换成ES语句
{
  "query": {
    "bool": {
      "must": [{
        "query_string": {
          "auto_generate_phrase_queries": true,
          "fields": [
            "action"
          ],
          "query": "bind"
        }
      },
	  {
        "range": {
	      "fields.bind_time": {
		    "from": null,
		    "include_lower": true,
		    "include_upper": false,
		    "to": "1489489200000"
		   }
		}
	  }	  
	  ],
      "must_not": {
        "query_string": {
          "auto_generate_phrase_queries": true,
          "fields": [
            "fields.alived"
          ],
          "query": "false"
        }
      }
    }
  },
  "from": 0,
  "size": 0,
  "aggs": {
    "group": {
      "aggregations": {
        "group_by_state": {
          "terms": {
            "field": "fields.idc_id.raw",
            "size": 1000
          }
       }
     },
     "date_histogram": {
       "field": "fields.bind_time",
       "interval": "1h"
     }
  }
 }
}

result 结果
{
    "TotalNum": 3,
    "bucket": {
      "1489474800000": 2,
      "1489474800000_fields.idc_id.raw": {
        "bjac": 2
       },
      "1489478400000": 0,
      "1489482000000": 1,
      "1489482000000_fields.idc_id.raw": {
        "bjac": 1
       }
    }
}

```
