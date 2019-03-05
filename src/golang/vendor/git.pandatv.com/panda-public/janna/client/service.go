package client

import (
	"fmt"
	"strings"
)

//   /service/grank/redis/
//   /service/grank/redis/id_1
const (
	servicePrefixKeyFormat = "/service/%s/%s/"
	serviceKeyFormat       = "/service/%s/%s/%s"
)

/*
   demo:
   Service {
       ServicePath: "/service/grank/redis/redis1",
       Tag: ["master", "web"],
       Address: "10.20.1.20",
       Port: 6379,
	   Weight: 100,
	   User: "dev",
	   Password: "dev123",
   }
*/

type ServiceValue struct {
	Tag      []string
	Address  string
	Port     int
	Weight   int
	User     string
	Password string
}

type Service struct {
	Key string
	ServiceValue
}

type ServiceEvent struct {
	Opt string
	Service
}

type ServiceClient interface {
	ServiceRegister(*Service) error
	ServiceDeregister(string) error
	ServiceGet(string) (*Service, error)
	ServiceGetAll(string) ([]Service, error)
	ServiceAddWatch(string) error
	ServiceRemoveWatch(string) error
	ServiceGetWatch() (chan ServiceEvent, error)
	ServiceCloseWatch() error
}

//callName is the name of your service
//serviceName is the name you want to use
//serviceKey is one of instance of the service specified by serviceName
func ServiceKey(callName, serviceName, serviceKey string) string {
	return fmt.Sprintf(serviceKeyFormat, callName, serviceName, serviceKey)
}

func SplitServiceKey(key string) (string, string, string, error) {
	serviceKeySlice := strings.SplitN(key, "/", 5)
	if len(serviceKeySlice) != 5 {
		return "", "", "", ErrFormatKey
	}
	if serviceKeySlice[1] != "service" {
		return "", "", "", ErrFormatKey
	}

	return serviceKeySlice[2], serviceKeySlice[3], serviceKeySlice[4], nil
}

//callName is the name of your service
//serviceName is the name you want to use
func ServicePrefixKey(callName, serviceName string) string {
	return fmt.Sprintf(servicePrefixKeyFormat, callName, serviceName)
}

func SplitServicePrefixKey(key string) (string, string, error) {
	servicePrefixKeySlice := strings.SplitN(key, "/", 5)
	if len(servicePrefixKeySlice) != 5 {
		return "", "", ErrFormatKey
	}
	if servicePrefixKeySlice[1] != "service" {
		return "", "", ErrFormatKey
	}

	return servicePrefixKeySlice[2], servicePrefixKeySlice[3], nil
}
