package config

type SQSConfig struct {
	Target              string `mapstructure:"target" default:"local"`
	Endpoint            string `mapstructure:"endpoint" default:""`
	AccountID           string `mapstructure:"account_id" default:""`
	Queue               string `mapstructure:"queue" default:""`
	Region              string `mapstructure:"region" default:""`
	MessageGroupID      string `mapstructure:"message_group_id" default:""`
	MaxNumberOfMessages int64  `mapstructure:"max_number_of_messages" default:"10"`
	WaitTimeSeconds     int64  `mapstructure:"wait_time_seconds" default:"20"`
	IsFIFO              bool   `mapstructure:"is_fifo" default:"false"`
}

type BindingPath struct {
	paths []string
}

func (c *SQSConfig) GetBindingPaths(paths ...string) []BindingPath {
	return []BindingPath{
		NewBindingPath(paths, "target"),
		NewBindingPath(paths, "endpoint"),
		NewBindingPath(paths, "account_id"),
		NewBindingPath(paths, "queue"),
		NewBindingPath(paths, "region"),
		NewBindingPath(paths, "message_group_id"),
		NewBindingPath(paths, "max_number_of_messages"),
		NewBindingPath(paths, "wait_time_seconds"),
		NewBindingPath(paths, "is_fifo"),
	}
}

func NewBindingPath(paths []string, path string) BindingPath {
	bindingPaths := make([]string, 0)

	bindingPaths = append(bindingPaths, paths...)
	bindingPaths = append(bindingPaths, path)

	return BindingPath{paths: bindingPaths}
}
