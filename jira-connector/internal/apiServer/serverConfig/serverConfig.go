package serverConfig

type ServerConfig struct {
	ConnectorPort uint   `yaml:"connector_port"`
	ConnectorHost string `yaml:"connector_host"`
}

func NewServerConfig(host string, port uint) *ServerConfig {
	return &ServerConfig{
		ConnectorPort: port,
		ConnectorHost: host,
	}
}
