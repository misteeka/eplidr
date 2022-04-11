package eplidr

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Table struct {
	name        string
	shardsCount uint
	Drivers     []*sql.DB

	creatingQuery []string

	hashFunc func(interface{}) uint
}

type Drivers interface{}

func NewTable(name string, shardsCount uint, creatingQuery []string, driverParam Drivers) *Table {
	var table *Table
	switch dataSource := driverParam.(type) {
	case []*sql.DB:
		table = &Table{
			name:          name,
			Drivers:       dataSource,
			shardsCount:   shardsCount,
			creatingQuery: creatingQuery,
			hashFunc:      StandardGetShardFunc,
		}
	case *sql.DB:
		drivers := make([]*sql.DB, shardsCount)
		for i := 0; i < int(shardsCount); i++ {
			drivers[i] = dataSource
		}
		table = &Table{
			name:          name,
			Drivers:       drivers,
			shardsCount:   shardsCount,
			creatingQuery: creatingQuery,
			hashFunc:      StandardGetShardFunc,
		}
	}
	return table
}

func Serialize(value interface{}) string {
	switch array := value.(type) {
	case map[string]interface{}:
		var result string
		keys := make([]string, 0, len(array))
		values := make([]interface{}, 0, len(array))
		for i := 0; i < len(array); i++ {
			if i == len(array)-1 {
				result += fmt.Sprintf("%s:%v", keys[i], values[i])
			} else {
				result += fmt.Sprintf("%s:%v,", keys[i], values[i])
			}
		}
		return result
	case []interface{}:
		var result string
		for i := 0; i < len(array); i++ {
			if i == len(array)-1 {
				result += fmt.Sprintf("%v", array[i])
			} else {
				result += fmt.Sprintf("%v,", array[i])
			}
		}
		return result
	default:
		return ""
	}
}
func DeserializeMap(serializedData string) {
	var result map[string]interface{}
	for i := 0; i < len(serializedData); i++ {
		key := ""
		value := ""
		result[key] = value
	}
}
func DeserializeSlice(serializedData string) {
	for i := 0; i < len(serializedData); i++ {

	}

}

func value(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("'%s'", v)
	case []interface{}: // Serialize s
		return fmt.Sprintf("'%v'", v[0])
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'E', -1, 64)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (table *Table) getName(shard uint) string {
	return table.name + strconv.FormatUint(uint64(shard), 10)
}

func (table *Table) getShard(key interface{}) uint {
	return table.hashFunc(key) % table.shardsCount
}
func (table *Table) init() {
	for i := 0; i < len(table.creatingQuery); i++ {
		table.creatingQuery[i] = strings.ReplaceAll(table.creatingQuery[i], "uint64", "BIGINT UNSIGNED")
		table.creatingQuery[i] = strings.ReplaceAll(table.creatingQuery[i], "int", "INTEGER")
		table.creatingQuery[i] = strings.ReplaceAll(table.creatingQuery[i], "{nn}", "NOT NULL")
		table.creatingQuery[i] = strings.ReplaceAll(table.creatingQuery[i], "{n}", "NULL")
		for a := 0; a < len(table.Drivers); a++ {
			_, err := table.Drivers[a].Exec(strings.Replace(table.creatingQuery[i], "{table}", fmt.Sprintf("`%s`", table.getName(uint(a))), 1))
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	}
}

func (table *Table) GetString(keyName interface{}, key interface{}, column string) (string, bool, error) {
	var result string
	err, found := table.Get(keyName, key, []string{column}, []interface{}{&result})
	if err != nil {
		return "", found, err
	}
	return result, found, nil
}
func (table *Table) GetInt(keyName interface{}, key interface{}, column string) (int, bool, error) {
	var result int
	err, found := table.Get(keyName, key, []string{column}, []interface{}{&result})
	if err != nil {
		return 0, found, err
	}
	return result, found, nil
}
func (table *Table) GetInt64(keyName interface{}, key interface{}, column string) (int64, bool, error) {
	var result int64
	err, found := table.Get(keyName, key, []string{column}, []interface{}{&result})
	if err != nil {
		return 0, found, err
	}
	return result, found, nil
}
func (table *Table) GetFloat(keyName interface{}, key interface{}, column string) (float64, bool, error) {
	var result float64
	err, found := table.Get(keyName, key, []string{column}, []interface{}{&result})
	if err != nil {
		return 0, found, err
	}
	return result, found, nil
}
func (table *Table) GetUint64(keyName interface{}, key interface{}, column string) (uint64, bool, error) {
	var result uint64
	err, found := table.Get(keyName, key, []string{column}, []interface{}{&result})
	if err != nil {
		return 0, found, err
	}
	return result, found, nil
}
func (table *Table) GetUint(keyName interface{}, key interface{}, column string) (uint, bool, error) {
	var result uint
	err, found := table.Get(keyName, key, []string{column}, []interface{}{&result})
	if err != nil {
		return 0, found, err
	}
	return result, found, nil
}
func (table *Table) GetBoolean(keyName interface{}, key interface{}, column string) (bool, bool, error) {
	var result bool
	err, found := table.Get(keyName, key, []string{column}, []interface{}{&result})
	if err != nil {
		return false, found, err
	}
	return result, found, nil
}

func (table *Table) Get(keyName interface{}, key interface{}, columns []string, data []interface{}) (error, bool) {
	query := fmt.Sprintf("SELECT %s FROM {table} WHERE `%v` = %s;", columnSliceToString(columns...), keyName, value(key))
	rows, err := table.Query(query, value(key))
	if err != nil {
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
func (table *Table) Put(keyName interface{}, key interface{}, columns []string, values []interface{}) error {
	// `%s` = ?
	if len(columns) != len(values) {
		return errors.New("keyTable.Put : len(columns) != len(data) ")
	}
	columnsString := ""
	valuesString := ""
	columns = append(columns, fmt.Sprintf("%v", keyName))
	values = append(values, key)
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
	query := fmt.Sprintf("INSERT INTO {table} (%s) values (%s);", columnsString, valuesString)
	_, err := table.Exec(query, value(key))
	if err != nil {
		return err
	}
	return nil
}
func (table *Table) Set(keyName interface{}, key interface{}, columns []string, values []interface{}) error {
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
	query := fmt.Sprintf("UPDATE {table} SET %s WHERE `%s` = %s;", s, keyName, value(key))
	_, err := table.Exec(query, value(key))
	if err != nil {
		return err
	}
	return nil
}
func (table *Table) Remove(keyName interface{}, key interface{}) error {
	query := fmt.Sprintf("DELETE FROM `{table}` WHERE `%s` = %s;", keyName, value(key))
	_, err := table.Exec(query, value(key))
	if err != nil {
		return err
	}
	return nil
}

func (table *Table) Exec(query string, key interface{}) (sql.Result, error) {
	shardNum := table.getShard(key)
	query = strings.Replace(query, "{table}", fmt.Sprintf("`%s`", table.getName(shardNum)), 1)
	return table.Drivers[shardNum].Exec(query)
}
func (table *Table) Query(query string, key interface{}) (*sql.Rows, error) {
	shardNum := table.getShard(key)
	query = strings.Replace(query, "{table}", fmt.Sprintf("`%s`", table.getName(shardNum)), 1)
	return table.Drivers[shardNum].Query(query)
}

func (table *Table) ReleaseRows(rows *sql.Rows) error {
	return rows.Close()
}

func (table *Table) RawTx(key interface{}) (*sql.Tx, error) {
	return table.Drivers[table.getShard(key)].Begin()
}

func (table *Table) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return table.Drivers[0].BeginTx(ctx, opts)
}

func (table *Table) ExecTx(query ...string) error {
	tx, err := table.Drivers[0].Begin()
	if err != nil {
		return err
	}
	for i := 0; i < len(query); i++ {
		_, err = tx.Exec(query[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (table *Table) SingleSet(keyName string, key interface{}, column string, value interface{}) error {
	return table.Set(keyName, key, []string{column}, []interface{}{value})
}

func (table *Table) Drop() {
	for i := 0; i < len(table.Drivers); i++ {
		table.Drivers[i].Exec(fmt.Sprintf("DROP TABLE %s;", table.getName(uint(i))))
	}
}

func (table *Table) GlobalExecUnsafe(query string) error {
	for i := 0; i < len(table.Drivers); i++ {
		_, err := table.Drivers[i].Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}
