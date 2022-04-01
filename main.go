package main

import (
	"fmt"
	"strconv"
)

func main() {
	db := New("table", 3, true, "CREATE TABLE {table_name} (`name` TEXT, `sgbdrn` INTEGER, `idrn` INTEGER);",
		"root:djDDI3-d3j3fH4FHhfhf-3h4uf3-dh3D3U392@tcp(localhost:41091)/db",
		"root:djDDI3-d3j3fH4FHhfhf-3h4uf3-dh3D3U392@tcp(localhost:41091)/db",
		"root:djDDI3-d3j3fH4FHhfhf-3h4uf3-dh3D3U392@tcp(localhost:41091)/db")
	for i := 0; i < 100; i++ {
		err := db.Execute("INSERT INTO {table_name} (`name`, `sgbdrn`, `idrn`) values (?,?,?);", "sgbdrn"+strconv.Itoa(i), "sgbdrn"+strconv.Itoa(i), 0, 100)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}
	// rows, err := db.Query("SELECT * FROM {table_name} WHERE `name` = {key};", "sgbdrn")
}
