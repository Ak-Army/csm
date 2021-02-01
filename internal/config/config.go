package config

import (
	"context"
	"log"
	"path/filepath"
	"sync"

	"github.com/Ak-Army/config"
	"github.com/Ak-Army/config/backend/file"
)

type Config struct {
	SnippetPath       string `config:"snippet-path,required"`
	DBFileName        string `config:"db-file-name,required"`
	Editor            string `config:"editor,required"`
	GitlabAccessToken string `config:"gitlab-access-token"`
	GitlabURL         string `config:"gitlab-url"`
}

type ConfigStore struct {
	mu     sync.Mutex
	config *Config
	err    error
}

var c *ConfigStore

func init() {
	loader, err := config.NewLoader(context.Background(),
		file.New(file.WithPath("config/config.json")),
	)
	c = &ConfigStore{}
	if err != nil {
		log.Fatal(err)
	}
	err = loader.Load(c)
	if err != nil {
		log.Fatal(err)
	}
}

func Get() *ConfigStore {
	return c
}

func (c *Config) DbFilePath() string {
	return filepath.Join(c.SnippetPath, c.DBFileName)
}

func (c *ConfigStore) NewSnapshot() interface{} {
	return &Config{
		SnippetPath: "~/csm",
		DBFileName:  "db.gob",
		Editor:      "vim",
	}
}

func (c *ConfigStore) SetSnapshot(confInterface interface{}, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	conf := confInterface.(*Config)
	c.config = conf
	c.err = err
}

func (c *ConfigStore) Config() (*Config, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.config, c.err
}
