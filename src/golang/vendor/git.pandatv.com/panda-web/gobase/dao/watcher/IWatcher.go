package watcher

type Service struct {
	Address  string
	Port     int
	Weight   int
	Master   bool
	Tag      []string
	Id       string
	Opt      string
	User     string
	Password string
}

type IWatcher interface {
	GetAllInstance(string, string, string) ([]Service, error)
	WatchInstance(string, string, string) (chan Service, error)
	Close() error
}
