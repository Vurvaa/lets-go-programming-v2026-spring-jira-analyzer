package models

type DBConfig struct {
	HostDB     string `yaml:"db_host"`
	PortDB     int    `yaml:"db_port"`
	UserDB     string `yaml:"db_user"`
	PasswordDB string `yaml:"db_passwd"`
	NameDB     string `yaml:"db_name"`
}

type ServerConfig struct {
	Repository    string `yaml:"jiraUrl"`
	ConnectorPort uint   `yaml:"connector_port"`
	ConnectorHost string `yaml:"connector_host"`
}
