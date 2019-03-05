### Usage

#### Init

guuid := NewGuuid("soho", "demo")

#### Get

uuid := guuid.Get()

#### GetTimeFromUuid

time := GetTimeFromUuid(uuid)

#### benchmark test
go test git.pandatv.com/panda-web/gobase/id -bench=. -benchtime=3s

