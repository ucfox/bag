package id

import (
	"git.pandatv.com/panda-web/gobase/log"
	"testing"
)

func init() {
	logkit.Init("id", logkit.LevelDebug)
}

func TestGetSuccess(t *testing.T) {
	guuid, nerr := NewGuuid("soho", "demo")
	if guuid == nil || nerr != nil {
		t.Error("TestGet NewGuuid Error err:" + nerr.Error())
	}
	uuid, gerr := guuid.Get()
	if uuid <= 0 || gerr != nil {
		t.Error("TestGet guuid Get Error err:" + gerr.Error())
	}

	time := GetTimeFromUuid(uuid)
	if time < 1500000000 || time > 1600000000 {
		t.Error("GetTimeFromUuid Error")
	}
}
func TestGetNewerr(t *testing.T) {
	guuid, nerr := NewGuuid("abc", "demo")
	if guuid != nil || nerr == nil {
		t.Error("TestGet NewGuuid Error err:" + nerr.Error())
	}
}

func TestGetFailover(t *testing.T) {
	guuid, nerr := NewGuuidConf("10.20.1.11:5001,10.20.1.11:5000")
	if guuid == nil || nerr != nil {
		t.Error("TestGet NewGuuid Error err:" + nerr.Error())
	}
	uuid, gerr := guuid.Get()
	if uuid <= 0 || gerr != nil {
		t.Error("TestGet guuid Get Error err:" + gerr.Error())
	}
}
func TestGetFailure(t *testing.T) {
	guuid, nerr := NewGuuidConf("10.20.1.11:5000,10.20.1.12:5000")
	if guuid == nil || nerr != nil {
		t.Error("TestGet NewGuuid Error err:" + nerr.Error())
	}
	uuid, gerr := guuid.Get()
	if uuid != 0 || gerr == nil {
		t.Error("TestGet guuid Get Error err:" + gerr.Error())
	}
}

func BenchmarkGets(b *testing.B) {
	guuid, nerr := NewGuuid("soho", "demo")
	if guuid == nil || nerr != nil {
		b.Error("BenchmarkGets NewGuuid Error err:" + nerr.Error())
	}
	for i := 0; i < b.N; i++ {
		uuid, gerr := guuid.Get()
		if uuid <= 0 || gerr != nil {
			b.Error("BenchmarkGets guuid Get Error err:" + gerr.Error())
		}
	}
}
