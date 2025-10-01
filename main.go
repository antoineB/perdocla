package main

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
