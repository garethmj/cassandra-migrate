package cql

import (
	"github.com/BurntSushi/toml"
)

type MigrationConfig struct {
	Scripts      Scripts
	Environments map[string]Environment
}

type Scripts struct {
	Path string
}

type Environment struct {
	Keyspace       string
	CassandraHosts string
}

func NewMigrationConfig(confPath string) (*MigrationConfig, error) {
	conf := &MigrationConfig{}
	if _, err := toml.DecodeFile(confPath, &conf); err != nil {
		return conf, err
	}

	return conf, nil
}
