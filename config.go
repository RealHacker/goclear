package goclear

type Configuration struct {
	MaxDepth int
	DBPath string
}

var Config Configuration

func InitializeConfig() {
	Config.MaxDepth = 5
	Config.DBPath = "./goclear.db"
}
