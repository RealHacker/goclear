package goclear

type Configuration struct {
	MaxDepth int
	DBPath string
}

var Config Configuration

func InitializeConfig() {
	// These will come from a configuration file
	Config.MaxDepth = 5
	Config.DBPath = "root@/goclear"
}
