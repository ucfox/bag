package redis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	//"time"
	//"git.pandatv.com/panda-web/gobase/log"
	//"github.com/garyburd/redigo/redis"
)

func getRedisServer(consulAgent, service string) (string, []string) {
	var resp *http.Response
	var err error
	consulUrl := fmt.Sprintf("http://%s/v1/catalog/service/%s", consulAgent, service)

	for i := 0; i < 3; i++ {
		resp, err = http.Get(consulUrl)
		if err == nil {
			break
		}
	}
	if err != nil {
		logkit.Errorf("get server error from agent %s service %s, http get err:%s\n", consulAgent, service, err)
		return "", []string{""}
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logkit.Errorf("get server error from agent %s service %s, read body err:%s\n", consulAgent, service, err)
		return "", []string{""}
	}

	var serinfo []catalogService
	err = json.Unmarshal(body, &serinfo)
	if err != nil {
		logkit.Errorf("get server error from agent %s service %s, json unmarshal err:%s\n", consulAgent, service, err)
		return "", []string{""}
	}

	var master string
	var slave []string

	for _, ser := range serinfo {
		for _, tag := range ser.ServiceTags {
			if masterTag == strings.ToLower(tag) {
				master = fmt.Sprintf("%s:%d", ser.ServiceAddress, ser.ServicePort)
			} else if slaveTag == strings.ToLower(tag) {
				slave = append(slave, fmt.Sprintf("%s:%d", ser.ServiceAddress, ser.ServicePort))
			}
		}
	}

	if len(slave) == 0 {
		slave = append(slave, "")
	}
	logkit.Infof("get server success from agent %s service %s, master:%s, slave:%s\n", consulAgent, service, master, slave)

	return master, slave
}

func healthCheckPool(dao *RedisBaseDao, body []byte, masterPwd, slavePwd, service string) {
	var healthinfo []healthInfo
	err := json.Unmarshal(body, &healthinfo)
	if err != nil {
		logkit.Errorf("json unmarshal %s's health info err:%s\n", service, err)
		return
	}

	for _, health := range healthinfo {
		for _, check := range health.ChecksRef {
			if check.ServiceName == service && check.ServiceID == health.ServiceRef.ID {
				parseHealthInfo(dao, &health, &check)
			}
		}
	}
}

func parseHealthInfo(dao *RedisBaseDao, health *healthInfo, check *healthCheck) {
	serviceAddr := fmt.Sprintf("%s:%d", health.ServiceRef.Address, health.ServiceRef.Port)
	status := false
	if check.Status != criticalStatus {
		status = true
	}

	for _, tag := range health.ServiceRef.Tags {
		if masterTag == strings.ToLower(tag) {
			dao.masterMux.Lock()
			if dao.masterStat.Addr == serviceAddr {
				dao.masterStat.Active = status
				logkit.Infof("master %s health %t\n", serviceAddr, status)
			}
			dao.masterMux.Unlock()
		} else if slaveTag == strings.ToLower(tag) {
			dao.slaveMux.Lock()
			for i, stat := range dao.slaveStat {
				if stat.Addr == serviceAddr {
					dao.slaveStat[i].Active = status
					logkit.Infof("slave %s health %t\n", serviceAddr, status)
				}
			}
			dao.slaveMux.Unlock()
		}
	}
}

func updatePool(dao *RedisBaseDao, body []byte, masterPwd, slavePwd, service string) {
	var serinfo []catalogService
	err := json.Unmarshal(body, &serinfo)
	if err != nil {
		return
	}

	slave := make(map[string]bool)
	master := dao.masterStat.Addr

	for _, ser := range serinfo {
		for _, tag := range ser.ServiceTags {
			if slaveTag == strings.ToLower(tag) {
				slave[fmt.Sprintf("%s:%d", ser.ServiceAddress, ser.ServicePort)] = true
			} else if masterTag == strings.ToLower(tag) {
				master = fmt.Sprintf("%s:%d", ser.ServiceAddress, ser.ServicePort)
			}
		}
	}

	tempReadStat := make([]*addrActive, 0, len(slave))
	deleteSlave := []string{}
	addSlavePool := make(map[string]*redis.Pool)
	for _, stat := range dao.slaveStat {
		_, ok := slave[stat.Addr]
		if ok {
			tempReadStat = append(tempReadStat, stat)
		} else {
			deleteSlave = append(deleteSlave, stat.Addr)
		}
	}

	for sl, _ := range slave {
		_, ok := dao.slavePool[sl]
		if !ok {
			addSlavePool[sl] = newRedisPoolCustom(sl, slavePwd, Default_connect_timeout, Default_read_timeout, Default_write_timeout, Default_idle_timeout, Default_max_active, Default_max_idle, Default_wait)
			tempReadStat = append(tempReadStat, &addrActive{sl, true})
		}
	}

	if len(deleteSlave) > 0 || len(addSlavePool) > 0 {
		logkit.Infof("slave pool change")
		dao.slaveMux.Lock()
		for _, addr := range deleteSlave {
			tempPool := dao.slavePool[addr]
			delete(dao.slavePool, addr)
			go tempPool.Close()
			logkit.Infof("delete slave %s\n", addr)
		}

		for sl, pool := range addSlavePool {
			dao.slavePool[sl] = pool
			logkit.Infof("add slave %s\n", sl)
		}

		dao.slaveStat = tempReadStat
		dao.slaveMux.Unlock()
	}

	var newMasterPool *redis.Pool
	if dao.masterStat.Addr != master {
		newMasterPool = newRedisPoolCustom(master, masterPwd, Default_connect_timeout, Default_read_timeout, Default_write_timeout, Default_idle_timeout, Default_max_active, Default_max_idle, Default_wait)
	}

	if newMasterPool != nil {
		dao.masterMux.Lock()
		logkit.Infof("change master from %s to %s\n", dao.masterStat.Addr, master)
		tempPool := dao.masterPool
		dao.masterPool = newMasterPool
		go tempPool.Close()
		dao.masterStat.Addr = master
		dao.masterStat.Active = true
		dao.masterMux.Unlock()
	}
}
