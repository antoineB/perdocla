package main

// PRAGMA foreign_keys = ON;
// PRAGMA optimize; (avant de fermer la base de données)
// PRAGMA busy_timeout=10000;
// PRAGMA journal_mode=WAL;
// PRAGMA synchronous=NORMAL;

import (
	"fmt"
	"perdoccla/src"
)


func main() {
	err := src.Exec()
	if err != nil {
		fmt.Println(err)
		return;
	}
}
