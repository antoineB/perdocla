package main

// PRAGMA foreign_keys = ON;
// PRAGMA optimize; (avant de fermer la base de données)
// PRAGMA busy_timeout=10000;
// PRAGMA journal_mode=WAL;
// PRAGMA synchronous=NORMAL;

import (
	"fmt"
	// "strings"
	// "log"
	// "os"

	// "image/jpeg"
	// "path/filepath"

	// "database/sql"

	_ "github.com/mattn/go-sqlite3"
	// snowball "github.com/snowballstem/snowball/go"
	// "github.com/mattn/go-gtk/glib"
	// "github.com/mattn/go-gtk/gtk"
	"perdoccla/src"
	// "perdoccla/snowball_french"
	// "github.com/otiai10/gosseract/v2"
	// "github.com/gen2brain/go-fitz"
)


func main() {
	// db, err := sql.Open("sqlite3", "./test.db")
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	// defer db.Close()
	// rows, err := db.Query("SELECT json_quote(jsonb_array($1))", "1,2")
	// fmt.Println(rows)

	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	// text := ""
	// if rows.Next() {
	// 	rows.Scan(&text)
	// 	fmt.Println(text)
	// 	tmps := strings.Split(text[2 : len(text)-2], ",")
	// 	for _, tmp := range tmps {
	// 		fmt.Println(tmp)
	// 	}
	// }


	err := src.Exec()
	if err != nil {
		fmt.Println(err)
		return;
	}

	// sb, err0 := src.Parse(db, os.Args[1:])
	// if err0 != nil {
	// 	fmt.Println(err)
	// 	return;
	// }

	// err = sb.Run(db)
	// if err != nil {
	// 	fmt.Println(err)
	// }


	// client := gosseract.NewClient()
	// defer client.Close()
	// file, _ := os.Open("/home/antoine/Document/clef_usb/banque/LivretA1.jpg")
	// defer file.Close()
	// stats, statsErr := file.Stat()
    // if statsErr != nil {
	// 	fmt.Println(statsErr)
	// 	return;
    // }
	// var size int64 = stats.Size()
    // bytes := make([]byte, size)
	// bufr := bufio.NewReader(file)
    // _, err0 := bufr.Read(bytes)
	// fmt.Println(err0)
	// err1 := client.SetImageFromBytes(bytes)
	// fmt.Println(err1)
	// text, err := client.Text()
	// fmt.Println(err)
	// fmt.Println(text)

	// enlever la pontuation (remplacer par des espaces)
	// minuscule
	// enlever les stop words
	// séparer chaque mot avec espace

	// env := snowball.NewEnv("Manger")
	// snowball_french.Stem(env)
	// fmt.Printf("stemmed word is: %s", env.Current())


	// doc, err := fitz.New("/home/antoine/Documents/Attestation de télétravail.pdf")
	// if err != nil {
	// 	panic(err)
	// }

	// defer doc.Close()

	// client := gosseract.NewClient()
	// defer client.Close()
	// client.SetLanguage("fra")

	// var img []byte

	// // Extract pages as images
	// for n := 0; n < doc.NumPage(); n++ {
	// 	img, err = doc.ImagePNG(n, 300)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	client.SetImageFromBytes(img)
	// 	text, _ := client.Text()
	// 	fmt.Println(text)
	// }
}
