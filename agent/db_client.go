package main

import (
	"fmt"
	"net/url"
	"os"
	"runtime"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	rDB  *gorm.DB // rDB is the read-only database instance
	rwDB *gorm.DB // rwDB is the read-write database instance
)

const dbName = "agent.db"

func CreateDatabaseFileIfNotExist() error {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		file, err := os.Create(dbName)
		if err != nil {
			return fmt.Errorf("failed to create database file: %v", err)
		}
		file.Close()
	}
	return nil
}

func InitiateDatabaseInstances() error {
	readDBInstance, err := OpenSqliteDatabase(dbName, true)
	if err != nil {
		return fmt.Errorf("failed to open read-only database: %v", err)
	}
	rDB = readDBInstance
	readWriteDBInstance, err := OpenSqliteDatabase(dbName, false)
	if err != nil {
		return fmt.Errorf("failed to open read-write database: %v", err)
	}
	rwDB = readWriteDBInstance
	return nil
}

func MigrateDatabase() error {
	if rwDB == nil {
		return fmt.Errorf("read-write database instance is nil or not initialized")
	}
	err := rwDB.AutoMigrate(&AgentConfig{}, &DNSEntry{}, &Volume{}, &VolumeMount{}, &EnvironmentVariable{}, &Container{}, &WireguardPeer{}, &StaticRoute{}, &NFRule{})
	if err != nil {
		return fmt.Errorf("failed to migrate Agent table: %v", err)
	}
	return nil
}

func OpenSqliteDatabase(file string, readonly bool) (*gorm.DB, error) {
	dbString := SQLiteDbString(file, readonly)
	gormDb, err := gorm.Open(sqlite.Open(dbString), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	db, err := gormDb.DB()
	if err != nil {
		return nil, err
	}

	pragmasToSet := []string{
		"temp_store=memory",
	}
	for _, pragma := range pragmasToSet {
		_, err = db.Exec("PRAGMA " + pragma + ";")
		if err != nil {
			return nil, err
		}
	}
	if readonly {
		db.SetMaxOpenConns(max(4, runtime.NumCPU()))
	} else {
		db.SetMaxOpenConns(1)
	}

	return gormDb, nil
}

func SQLiteDbString(file string, readonly bool) string {
	connectionParams := make(url.Values)
	connectionParams.Add("_busy_timeout", "5000")
	connectionParams.Add("_synchronous", "NORMAL")
	connectionParams.Add("_cache_size", "-20000")
	connectionParams.Add("_foreign_keys", "true")
	if readonly {
		connectionParams.Add("mode", "ro")
	} else {
		connectionParams.Add("_journal_mode", "WAL")
		connectionParams.Add("_txlock", "IMMEDIATE")
		connectionParams.Add("mode", "rwc")
	}

	return "file:" + file + "?" + connectionParams.Encode()
}
