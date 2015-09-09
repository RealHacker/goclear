package goclear

type Configuration struct {
	MaxDepth int
}

var Config Configuration

func init() {
	Config.MaxDepth = 5
}
