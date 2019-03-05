package consul

const (
	passingStatus  = "passing"
	warningStatus  = "warning"
	criticalStatus = "critical"
)

//catalog service response struct
type catalogService struct {
	Node                     string
	Address                  string
	ServiceID                string
	ServiceName              string
	ServiceTags              []string
	ServiceAddress           string
	ServicePort              int
	ServiceEnableTagOverride bool
	CreateIndex              int
	ModifyIndex              int
}

//health check response struct
type healthInfo struct {
	NodeRef    healthNode    `json:"Node"`
	ServiceRef healthService `json:"Service"`
	ChecksRef  []healthCheck `json:"Checks"`
}

//health check node info
type healthNode struct {
	Node            string
	Address         string
	TaggedAddresses map[string]string
	CreateIndex     int
	ModifyIndex     int
}

//health check service info
type healthService struct {
	ID                string
	Service           string
	Tags              []string
	Address           string
	Port              int
	EnableTagOverride bool
	CreateIndex       int
	ModifyIndex       int
}

//health check info
type healthCheck struct {
	Node        string
	CheckID     string
	Name        string
	Status      string
	Notes       string
	Output      string
	ServiceID   string
	ServiceName string
	CreateIndex int
	ModifyIndex int
}

type consulCallBack func(dao *RedisBaseDao, body []byte, masterPwd, slavePwd, service string)

func (dao *RedisBaseDao) consulListen(consulAgent, endPoint, service, masterPwd, slavePwd string, fn consulCallBack) {
	consul_url := fmt.Sprintf("http://%s/v1/%s/service/%s", consulAgent, endPoint, service)
	consul_url_wait := fmt.Sprintf("http://%s/v1/%s/service/%s?wait=60s&index=", consulAgent, endPoint, service)
	var index string
	var resp *http.Response
	var err error

	dao.instanceWait.Add(1)
	defer dao.instanceWait.Done()

	for dao.useable {
		if index == "" {
			resp, err = http.Get(consul_url + index)
		} else {
			resp, err = http.Get(consul_url_wait + index)
		}
		if err != nil {
			continue
		}

		consulIndex, ok := resp.Header["X-Consul-Index"]
		if !ok {
			index = ""
			resp.Body.Close()
			continue
		}
		if index == consulIndex[0] {
			resp.Body.Close()
			continue
		}
		logkit.Debugf("endPoint %s, service %s, old index %s, new index %s\n", endPoint, service, index, consulIndex[0])
		index = consulIndex[0]

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		fn(dao, body, masterPwd, slavePwd, service)
	}
}
