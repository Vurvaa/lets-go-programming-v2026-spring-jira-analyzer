package serverConfig

type ServerConfig struct {
	Repository    string `yaml:"repository"`
	ConnectorPort uint   `yaml:"connector_port"`
	ConnectorHost string `yaml:"connector_host"`
}

func NewServerConfig(repository, host string, port uint) *ServerConfig {
	return &ServerConfig{
		Repository:    repository,
		ConnectorPort: port,
		ConnectorHost: host,
	}
}
