package data

import (
	"path"
)

type Config interface {
	LoadFromFile(p string)
}

var (
	PlayerInit    EntityConfig
	Player2Init   EntityConfig
	CoreConfig    EntityConfig
	TurretConfigs map[string]EntityConfig
	EnemyConfigs  map[string]EntityConfig
)

func LoadConfigurations() error {
	// Load the players configuration
	config, err := NewPlayerConfig(1)
	PlayerInit = config
	if err != nil {
		return err
	}

	config, err = NewPlayerConfig(2)
	Player2Init = config
	if err != nil {
		return err
	}

	// Load the core configuration
	CoreConfig, err = NewCoreConfig()
	if err != nil {
		return err
	}

	// Traverse the turret config folder and load all turret configurations
	TurretConfigs = make(map[string]EntityConfig)
	turretFiles, err := GetPathFiles(path.Join("entities", "turrets"))
	println("Loading turret configs:")
	if err != nil {
		return err
	}
	for _, fileName := range turretFiles {
		println("\t", fileName)
		TurretConfigs[fileName], err = NewTurretConfig(fileName)
		if err != nil {
			return err
		}
	}

	// Traverse the enemy config folder and load all enemy configurations
	EnemyConfigs = make(map[string]EntityConfig)
	enemyFiles, err := GetPathFiles(path.Join("entities", "enemies"))
	println("Loading enemy configs:")
	if err != nil {
		return err
	}
	for _, fileName := range enemyFiles {
		println("\t", fileName)
		EnemyConfigs[fileName], err = NewEnemyConfig(fileName)
		if err != nil {
			return err
		}
	}
	return nil
}
