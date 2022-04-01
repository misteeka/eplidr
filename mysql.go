package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
	"time"
)
type DB struct {
	Drivers          []*sql.DB
	tableName        string
	shardsCount      uint
	createTableQuery string

	hashFunc     func(interface{}) uint
}

func StandardGetShardFunc(key interface{}) uint {
	return fnv32(key.(string))
}

func fnv32(key string) uint {
	hash := uint(2166136261)
	const prime32 = uint(16777619)
	keyLength := len(key)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint(key[i])
	}
	return hash
}

func getDrivers(dataSources ...string) []*sql.DB {
	drivers := make([]*sql.DB, len(dataSources))
	for i := 0; i < len(dataSources); i++ {
		fmt.Println(dataSources[i])
		driver, err := sql.Open("mysql", dataSources[i])
		if err != nil {
			fmt.Println("getDrivers error " + err.Error())
			continue
		}
		driver.SetMaxOpenConns(5)
		driver.SetMaxIdleConns(5)
		driver.SetConnMaxLifetime(time.Minute * 5)
		drivers[i] = driver
	}
	return drivers
}

func (db *DB) getShardFunc(key interface{}) uint {
	return db.hashFunc(key) % db.shardsCount
}

func (db *DB) createTables() {
	for i := 0; i < len(db.Drivers); i++ {
		_, err := db.Drivers[i].Exec(strings.Replace(db.createTableQuery, "{table_name}", db.tableName+strconv.Itoa(i), 1))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
}

func New(tableName string, shardsCount uint, createTables bool, createTableQuery string, dataSources ...string) *DB {
	db := DB{
		Drivers:          getDrivers(dataSources...),
		createTableQuery: createTableQuery,
		tableName:        tableName,
		shardsCount:      shardsCount,
		hashFunc:     	  StandardGetShardFunc,
	}
	if createTables {
		db.createTables()
	}
	return &db
}

func (db *DB) Execute(query string, key string, args ...interface{}) error {
	shardNum := db.getShardFunc(key)
	query = strings.Replace(query, "{table_name}", db.tableName+strconv.Itoa(int(shardNum)), 1)
	query = strings.Replace(query, "{key}", key, 1)
	_, err := db.Drivers[shardNum].Exec(query, args...)
	return err
}

func (db *DB) Query(query string, key string, args ...interface{}) (*sql.Rows, error) {
	shardNum := db.getShardFunc(key)
	query = strings.Replace(query, "{table_name}", db.tableName+strconv.Itoa(int(shardNum)), 1)
	query = strings.Replace(query, "{key}", key, 1)
	return db.Drivers[shardNum].Query(query, args...)
}
