package config

import (
	"path/filepath"
)

/* Paths */
var DefaultContent = `{"training": []}`
var	file = "training.json"
var BaseDB = "TRAINING"
var BaseLocal = "DATABASES"
var	BaseBackup = "/media/veikko/VK DATA/"

var LocalFile = filepath.Join(BaseLocal, BaseDB, file)
var BackupFile = filepath.Join(BaseBackup, BaseLocal, BaseDB, file)

