package eplidr

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Tx struct {
	driver *sql.Tx
	shard  uint
}

func (tx *Tx) getTableName(table string) string {
	return table + strconv.FormatUint(uint64(tx.shard), 10)
}

func (tx *Tx) GetString(table string, keyName interface{}, key interface{}, column string, lock bool) (string, bool, error) {
	var result string
	err, found := tx.Get(table, keyName, key, []string{column}, []interface{}{&result}, lock)
	if err != nil {
		return "", found, err
	}
	return result, found, nil
}
func (tx *Tx) GetInt(table string, keyName interface{}, key interface{}, column string, lock bool) (int, bool, error) {
	var result int
	err, found := tx.Get(table, keyName, key, []string{column}, []interface{}{&result}, lock)
	if err != nil {
		return 0, found, err
	}
	return result, found, nil
}
func (tx *Tx) GetInt64(table string, keyName interface{}, key interface{}, column string, lock bool) (int64, bool, error) {
	var result int64
	err, found := tx.Get(table, keyName, key, []string{column}, []interface{}{&result}, lock)
	if err != nil {
		return 0, found, err
	}
	return result, found, nil
}
func (tx *Tx) GetFloat(table string, keyName interface{}, key interface{}, column string, lock bool) (float64, bool, error) {
	var result float64
	err, found := tx.Get(table, keyName, key, []string{column}, []interface{}{&result}, lock)
	if err != nil {
		return 0, found, err
	}
	return result, found, nil
}
func (tx *Tx) GetUint(table string, keyName interface{}, key interface{}, column string, lock bool) (uint64, bool, error) {
	var result uint64
	err, found := tx.Get(table, keyName, key, []string{column}, []interface{}{&result}, lock)
	if err != nil {
		return 0, found, err
	}
	return result, found, nil
}
func (tx *Tx) GetBoolean(table string, keyName interface{}, key interface{}, column string, lock bool) (bool, bool, error) {
	var result bool
	err, found := tx.Get(table, keyName, key, []string{column}, []interface{}{&result}, lock)
	if err != nil {
		return false, found, err
	}
	return result, found, nil
}

func (tx *Tx) Get(table string, keyName interface{}, key interface{}, columns []string, data []interface{}, lock bool) (error, bool) {
	var lockStatement string
	if lock {
		lockStatement = " FOR UPDATE;"
	} else {
		lockStatement = ""
	}
	query := fmt.Sprintf("SELECT %s FROM `%s` WHERE `%v` = %s%s;", columnSliceToString(columns...), tx.getTableName(table), keyName, value(key), lockStatement)
	rows, err := tx.Query(table, query)
	if err != nil {
		rows.Close()
		return err, false
	}
	if rows.Next() {
		err := rows.Scan(data...)
		if err != nil {
			rows.Close()
			return err, true
		}
		rows.Close()
	} else {
		rows.Close()
		return nil, false
	}
	return nil, true
}
func (tx *Tx) Put(table string, columns []string, values []interface{}) error {
	if len(columns) != len(values) {
		return errors.New("keyTable.Put : len(columns) != len(data) ")
	}
	columnsString := ""
	valuesString := ""
	for i := 0; i < len(columns); i++ {
		if i == len(columns)-1 {
			columnsString += fmt.Sprintf("`%s`", columns[i])
		} else {
			columnsString += fmt.Sprintf("`%s`, ", columns[i])
		}
	}
	for i := 0; i < len(values); i++ {
		if i == len(values)-1 {
			valuesString += fmt.Sprintf("%s", value(values[i]))
		} else {
			valuesString += fmt.Sprintf("%s, ", value(values[i]))
		}
	}
	query := fmt.Sprintf("INSERT INTO `%s` (%s) values (%s);", tx.getTableName(table), columnsString, valuesString)
	_, err := tx.driver.Exec(query)
	if err != nil {
		return err
	}
	return nil
}
func (tx *Tx) Set(table string, keyName interface{}, key interface{}, columns []string, values []interface{}) error {
	if len(columns) != len(values) {
		return errors.New("keyTable.Set : len(columns) != len(values) ")
	}
	s := ""
	for i := 0; i < len(columns); i++ {
		if i == len(columns)-1 {
			s += fmt.Sprintf("`%s` = %s", columns[i], value(values[i]))
		} else {
			s += fmt.Sprintf("`%s` = %s, ", columns[i], value(values[i]))
		}
	}
	query := fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s` = %s;", tx.getTableName(table), s, keyName, value(key))
	_, err := tx.driver.Exec(query)
	if err != nil {
		return err
	}
	return nil
}
func (tx *Tx) Remove(table string, keyName interface{}, key interface{}) error {
	query := fmt.Sprintf("DELETE FROM `%s` WHERE `%s` = %s;", tx.getTableName(table), keyName, value(key))
	_, err := tx.driver.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (tx *Tx) Exec(table string, query string) (sql.Result, error) {
	query = strings.ReplaceAll(query, "{table}", fmt.Sprintf("`%s`", tx.getTableName(table)))
	return tx.driver.Exec(query)
}

func (tx *Tx) Query(table string, query string) (*sql.Rows, error) {
	query = strings.ReplaceAll(query, "{table}", fmt.Sprintf("`%s`", tx.getTableName(table)))
	return tx.driver.Query(query)
}

func (tx *Tx) SingleSet(table string, keyName string, key interface{}, column string, value interface{}) error {
	return tx.Set(table, keyName, key, []string{column}, []interface{}{value})
}

func (tx *Tx) Commit() error {
	return tx.driver.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.driver.Rollback()
}
func (tx *Tx) Fail() {
	tx.driver.Rollback()
}
