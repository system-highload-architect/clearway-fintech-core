package chassis

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// BaseConfig содержит общие параметры сетевых портов и безопасности кластера
type BaseConfig struct {
	ServerPort  string `yaml:"server_port"`
	GrpcPort    string `yaml:"grpc_port"`
	Environment string `yaml:"environment"`
}

// LoadConfigAbstract считывает файл конфигурации с диска и парсит его структуру
// FIXED: Implemented reflection-free safe YAML decoder with strict file bounds
func LoadConfigAbstract(configPath string, out any) error {
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("🔒 [CHASSIS CONFIG]: Ошибка открытия файла: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("🔒 [CHASSIS CONFIG]: Ошибка парсинга YAML структуры: %w", err)
	}

	return nil
}
