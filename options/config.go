package options

import (
	"encoding/json"
)

type Client struct {
	AppId string `json:"appId"`
	Key   string `json:"key"`
}

type Redis struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	KeyPrefix string `json:"keyPrefix"`
	Db        int    `json:"db"`
}

type Sqlite struct {
	DatabasePath string `json:"databasePath"`
}

type DatabaseConfig struct {
	Redis           Redis  `json:"redis"`
	Sqlite          Sqlite `json:"sqlite"`
	PublishPresence bool   `json:"publishPresence"`
}

type Hosts []string

type Subscribers struct {
	Http  bool `json:"http"`
	Redis bool `json:"redis"`
}

type ApiOriginAllow struct {
	AllowCors    bool   `json:"allowCors"`
	AllowOrigin  string `json:"allowOrigin"`
	AllowMethods string `json:"allowMethods"`
	AllowHeaders string `json:"allowHeaders"`
}

type Config struct {
	AuthHost       interface{}       `json:"authHost"`
	AuthEndpoint   string            `json:"authEndpoint"`
	Clients        []Client          `json:"clients"`
	Database       string            `json:"database"`
	DatabaseConfig DatabaseConfig    `json:"databaseConfig"`
	DevMode        bool              `json:"devMode"`
	Host           interface{}       `json:"host"`
	Port           string            `json:"port"`
	Protocol       string            `json:"protocol"`
	Socketio       *ServerOptions    `json:"socketio"`
	SslCertPath    string            `json:"sslCertPath"`
	SslKeyPath     string            `json:"sslKeyPath"`
	Subscribers    Subscribers       `json:"subscribers"`
	ApiOriginAllow ApiOriginAllow    `json:"apiOriginAllow"`
	Headers        map[string]string `json:"header"`
}

func Assign(_old *Config, _new *Config) (*Config, error) {
	_default := &Config{}
	if old_data, err := json.Marshal(_old); err != nil {
		return _default, err
	} else {
		if err := json.Unmarshal(old_data, &_default); err != nil {
			return _default, err
		}
	}
	if new_data, err := json.Marshal(_new); err != nil {
		return _default, err
	} else {
		if err := json.Unmarshal(new_data, &_default); err != nil {
			return _default, err
		}
	}
	return _default, nil
}
