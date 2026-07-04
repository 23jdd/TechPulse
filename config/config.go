package config

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Mode string

const (
	Dev     Mode = "dev"
	Release Mode = "release"
)

type Config struct {
	Mode         Mode   `yaml:"mode"`
	Path         string `yaml:"path"`
	ApiPort      int    `yaml:"api_port"`
	ApiAddress   string `yaml:"api_address"`
	ClientId     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Domain       string `yaml:"domain"`
}

func MustLoadConfig() *Config {
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	err := v.ReadInConfig()
	if err != nil {
		panic(err)
	}
	var c Config
	// 指定使用 yaml 标签进行字段映射（默认是 mapstructure 标签）
	if err := v.Unmarshal(&c, viper.DecoderConfigOption(func(config *mapstructure.DecoderConfig) {
		config.TagName = "yaml"
	})); err != nil {
		panic(err)
	}
	return &c
}
