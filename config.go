package config

import (
	"encoding/json"
	"os"
)

// Config 结构体表示配置信息
type Config struct {
	APIKey      string `json:"api_key"`
	SecretKey   string `json:"secret_key"`
	OtherConfig string `json:"other_config"`
}

// ReadConfig 从指定路径读取配置文件并返回 Config 结构体
func ReadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
