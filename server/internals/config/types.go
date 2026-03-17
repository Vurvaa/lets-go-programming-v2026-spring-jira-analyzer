package config

type DBConfig struct {
	HostDB     string `yaml:"db_host"`
	PortDB     int    `yaml:"db_port"`
	UserDB     string `yaml:"db_user"`
	PasswordDB string `yaml:"db_passwd"`
	NameDB     string `yaml:"db_name"`
}
