package id

import (
	"bytes"
	"errors"
	"git.pandatv.com/panda-web/gobase/log"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	UUID_KEY     = "uuid"
	UUID_TIMEOUT = 1
)

var uuidBytes []byte

func init() {
	uuidBytes = []byte("get " + UUID_KEY + "\r\n")
}

var confs = map[string]([]string){
	"dev_test":  {"10.20.1.11:5001", "10.20.1.11:5001"},
	"demo_test": {"10.20.1.11:5001", "10.20.1.11:5001"},

	"demo_soho": {"10.20.1.11:5001", "10.20.1.11:5001"},

	"beta_bjac": {"10.110.66.17:5001", "10.110.66.17:5001"},
	"beta_bjza": {"10.120.6.13:5001", "10.120.6.13:5001"},

	"online_bjac": {"10.110.17.229:5001", "10.110.17.230:5001"},
	"online_bjza": {"10.120.5.22:5001", "10.120.5.23:5001"},
}

type Guuid struct {
	conf     []string
	posCount int32
}

func NewGuuid(idc string, env string) (*Guuid, error) {
	guuid := new(Guuid)
	guuid.conf = confs[env+"_"+idc]
	if len(guuid.conf) < 1 {
		return nil, errors.New("Guuid Newguuid conf:" + strings.Join(guuid.conf, ","))
	}
	return guuid, nil
}
func NewGuuidConf(confStr string) (*Guuid, error) {
	guuid := new(Guuid)
	guuid.conf = strings.Split(confStr, ",")
	if len(guuid.conf) < 1 {
		return nil, errors.New("Guuid Newguuid conf:" + strings.Join(guuid.conf, ","))
	}
	return guuid, nil
}

func (this *Guuid) Get() (uint64, error) {
	pos := this.incrAndGet()
	confLen := len(this.conf)
	for i := 0; i < confLen+1; i++ {
		id := get(this.conf[(pos+i)%confLen])
		if id > 0 {
			return id, nil
		}
	}
	return 0, errors.New("Guuid Get failure!")
}
func (this *Guuid) incrAndGet() int { // Round Robin choose one pool
	count := atomic.AddInt32(&this.posCount, 1)
	if count >= 0 && count < 2100000000 {
		return int(count)
	} else {
		atomic.StoreInt32(&this.posCount, 0)
		return 0
	}
}

func get(server string) uint64 {
	conn, cerr := net.DialTimeout("tcp", server, UUID_TIMEOUT*time.Second)
	if cerr != nil {
		logkit.Warnf("Guuid get connect err:%s", cerr.Error())
		return 0
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(UUID_TIMEOUT * time.Second))

	wn, werr := conn.Write(uuidBytes)
	if wn < 10 || werr != nil {
		logkit.Warnf("Guuid get write num:%d,write err:%s", wn, werr.Error())
		return 0
	}

	b := make([]byte, 40)
	rn, rerr := conn.Read(b)
	if rn < 40 || rerr != nil {
		logkit.Warnf("Guuid get read num:%d,err:%s", rn, rerr.Error())
		return 0
	}

	idfields := bytes.Fields(b)
	logkit.Debugf("Guuid get read num:%d,idfields:%q", rn, idfields)
	if len(idfields) < 6 {
		logkit.Warnf("Guuid get idfields len:%d", len(idfields))
		return 0
	}

	id, perr := strconv.ParseUint(string(idfields[4]), 10, 64)
	if perr != nil {
		logkit.Warnf("Guuid get parse uint64:%d,err:%s", id, perr.Error())
		return 0
	}
	return id
}
