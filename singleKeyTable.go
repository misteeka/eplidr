package eplidr

import (
	"context"
	"database/sql"
)

type SingleKeyTable struct {
	Table *Table
	key   string
}

func NewSingleKeyTable(name string, key string, shardsCount uint, creatingQuery []string, drivers Drivers) *SingleKeyTable {
	// params:
	// [0] dataSource
	// [1]
	return &SingleKeyTable{
		Table: NewTable(name, shardsCount, creatingQuery, drivers),
		key:   key,
	}
}

func SingleKeyImplementation(keyTable *Table, key string) *SingleKeyTable {
	return &SingleKeyTable{
		Table: keyTable,
		key:   key,
	}
}

func columnSliceToString(columns ...string) string {
	result := ""
	for i := 0; i < len(columns); i++ {
		result += " `" + columns[i] + "`"
	}
	return result
}

func (table *SingleKeyTable) GetString(key interface{}, column string) (string, bool, error) {
	return table.Table.GetString(table.key, key, column)
}
func (table *SingleKeyTable) GetInt(key interface{}, column string) (int, bool, error) {
	return table.Table.GetInt(table.key, key, column)
}
func (table *SingleKeyTable) GetInt64(key interface{}, column string) (int64, bool, error) {
	return table.Table.GetInt64(table.key, key, column)
}
func (table *SingleKeyTable) GetFloat(key interface{}, column string) (float64, bool, error) {
	return table.Table.GetFloat(table.key, key, column)
}
func (table *SingleKeyTable) GetUint(key interface{}, column string) (uint64, bool, error) {
	var result uint64
	err, found := table.Get(key, []string{column}, []interface{}{&result})
	if err != nil {
		return 0, found, err
	}
	return result, found, nil
}
func (table *SingleKeyTable) GetBoolean(key interface{}, column string) (bool, bool, error) {
	var result bool
	err, found := table.Get(key, []string{column}, []interface{}{&result})
	if err != nil {
		return false, found, err
	}
	return result, found, nil
}

func (table *SingleKeyTable) Get(key interface{}, columns []string, data []interface{}) (error, bool) {
	return table.Table.Get(table.key, key, columns, data)
}
func (table *SingleKeyTable) Set(key interface{}, columns []string, data []interface{}) error {
	return table.Table.Set(table.key, key, columns, data)
}

func (table *SingleKeyTable) SingleSet(key interface{}, column string, data interface{}) error {
	return table.Table.Set(table.key, key, []string{column}, []interface{}{data})
}

func (table *SingleKeyTable) Put(key interface{}, columns []string, values []interface{}) error {
	return table.Table.Put(table.key, key, columns, values)
}

func (table *SingleKeyTable) Remove(key interface{}) error {
	return table.Table.Remove(table.key, key)
}

func (table *SingleKeyTable) ReleaseRows(rows *sql.Rows) error {
	return rows.Close()
}

func (table *SingleKeyTable) Begin(key interface{}) (*Tx, error) {
	shard := table.Table.getShard(key)
	driver, err := table.Table.Drivers[shard].Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{
		driver: driver,
		shard:  shard,
	}, nil
}

func (table *SingleKeyTable) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, nil
}

func (table *SingleKeyTable) Exec(query string, key interface{}) (sql.Result, error) {
	return table.Table.Exec(query, key)
}
func (table *SingleKeyTable) Query(query string, key interface{}) (*sql.Rows, error) {
	return table.Table.Query(query, key)
}
