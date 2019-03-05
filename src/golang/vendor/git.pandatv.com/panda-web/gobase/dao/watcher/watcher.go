package watcher

import (
	"errors"
	"strings"

	janna "git.pandatv.com/panda-public/janna/client"
	"git.pandatv.com/panda-web/gobase/log"
)

const (
	masterTag = "master"
	slaveTag  = "slave"
)

var (
	ErrWrongFormatService = errors.New("wrong format service")
	ErrNotTargetTag       = errors.New("not target tag")
)

type Watcher struct {
	servConf janna.ServiceClient
	closeCh  chan struct{}
}

func NewWatcher(servConf janna.ServiceClient) (*Watcher, error) {
	return &Watcher{servConf, make(chan struct{})}, nil
}

func (this *Watcher) GetAllInstance(callName, serviceName, targetTag string) ([]Service, error) {
	servicePrefixKey := janna.ServicePrefixKey(callName, serviceName)
	instanceInfos, err := this.servConf.ServiceGetAll(servicePrefixKey)
	if err != nil {
		return nil, err
	}

	serv, err := filterInstances(instanceInfos, callName, serviceName, targetTag, janna.OptPut)
	if err != nil {
		return nil, err
	}

	return serv, nil
}

func (this *Watcher) WatchInstance(callName, serviceName, targetTag string) (chan Service, error) {
	servicePrefixKey := janna.ServicePrefixKey(callName, serviceName)
	err := this.servConf.ServiceAddWatch(servicePrefixKey)
	if err != nil {
		return nil, err
	}

	ch, err := this.servConf.ServiceGetWatch()
	if err != nil {
		return nil, err
	}

	chServ := make(chan Service, 10)

	go func() {
		defer close(chServ)
		for {
			select {
			case event := <-ch:
				serv, err := filterInstance(event.Service, callName, serviceName, targetTag, event.Opt)
				if err == ErrNotTargetTag {
					continue
				}
				if err != nil {
					logkit.Warnf("service %+v is wrong format", event)
					continue
				}
				chServ <- serv
			case <-this.closeCh:
				return
			}
		}
	}()

	return chServ, nil
}

func (this *Watcher) Close() error {
	close(this.closeCh)

	return nil
}

func filterInstances(instanceInfos []janna.Service, callName, serviceName, targetTag, optType string) ([]Service, error) {
	servList := make([]Service, 0)
	for _, instanceInfo := range instanceInfos {
		serv, err := filterInstance(instanceInfo, callName, serviceName, targetTag, optType)
		if err == ErrNotTargetTag {
			continue
		}
		if err != nil {
			logkit.Warnf("service %+v error:%s", instanceInfo, err)
			continue
		}
		servList = append(servList, serv)
	}

	return servList, nil
}

func filterInstance(instanceInfo janna.Service, callName, serviceName, targetTag, optType string) (Service, error) {
	serv := Service{}

	callNameTemp, serviceNameTemp, serviceId, err := janna.SplitServiceKey(instanceInfo.Key)
	if err != nil || callNameTemp != callName || serviceName != serviceNameTemp {
		return serv, ErrWrongFormatService
	}

	//对于delete事件，tag依赖WithPreKv
	if optType == janna.OptDelete && len(instanceInfo.Tag) == 0 {
		instanceInfo.Tag = strings.Split(serviceId, "_")
	}
	validBool, masterBool, tag := checkTag(targetTag, instanceInfo.Tag)
	if !validBool {
		return serv, ErrNotTargetTag
	}

	serv.Address = instanceInfo.Address
	serv.Port = instanceInfo.Port
	serv.Weight = instanceInfo.Weight
	serv.Master = masterBool
	serv.Tag = tag
	serv.Id = serviceId
	serv.Opt = optType
	serv.User = instanceInfo.User
	serv.Password = instanceInfo.Password

	return serv, nil
}

func checkTag(targetTag string, tags []string) (bool, bool, []string) {
	targetBool := false
	masterBool := false
	validBool := false
	resultTags := make([]string, 0)

	if targetTag == "" {
		targetBool = true
	}

	for _, tag := range tags {
		lowTag := strings.ToLower(tag)
		if lowTag == masterTag || lowTag == slaveTag {
			validBool = true
			if lowTag == masterTag {
				masterBool = true
			}
		} else {
			resultTags = append(resultTags, tag)
		}

		if targetTag == tag {
			targetBool = true
		}
	}

	if !targetBool {
		validBool = false
	}

	return validBool, masterBool, resultTags
}
