package cmd

type Config struct {
	port int
}

func New() *Config {
	c := new(Config)
	c.port = 9999
	return c
}

func (c *Config) SetPort(port int) {
	c.port = port
}
