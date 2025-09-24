package src

import (
	"sort"
	"strconv"
	"strings"
	"unicode"

	snowball "github.com/snowballstem/snowball/go"

	"database/sql"

	"perdoccla/snowball_english"
	"perdoccla/snowball_french"
)

type Language int64

const (
	French Language = iota
	English
)

func nextWord(text []rune) (string, []rune) {
	word := ""
	started := false
	nb := 0
	for _, r := range text {
		if unicode.IsLetter(r) {
			started = true
			word = word + string(r)
			nb = nb + 1
		} else {
			nb = nb + 1
			if started {
				return word, text[nb:]
			}
		}
	}

	return "", nil
}

func StemText(text string, language Language) []string {
	runes := []rune(text)
	var word string
	words := []string{}
	for {
		word, runes = nextWord(runes)
		if (word == "") {
			return words
		}
		env := snowball.NewEnv(word)
		switch language {
		case French:
			snowball_french.Stem(env)
		case English:
			snowball_english.Stem(env)
		}
		words = append(words, env.Current())
	}
}

func convertJsonArray(array string) []int {
	elements := strings.Split(array[2:len(array) - 2], ",")
	elementsInt := []int{}
	for _, element := range elements {
		elementInt, _ := strconv.Atoi(element)
		elementsInt = append(elementsInt, elementInt)
	}
	return elementsInt
}

func SearchInvertedIndex(connection *sql.DB, words []string) ([]int, error) {
	rows, err := connection.Query(
		"SELECT document_id, word, json_quote(positions) FROM document_inverted_index WHERE word IN $1",
		words,
	)
    if err != nil {
        return nil, err
    }

	wordIndex := make(map[string]int)
	for index, word := range words {
		wordIndex[word] = index
	}

	documentPositions := make([]map[int][]int, len(words))

	documentIds := []int{}
	var documentId int
	var word string
	var positionsStr string
	if rows.Next() {
		rows.Scan(&documentId, &word, &positionsStr)
		positions := convertJsonArray(positionsStr)
		sort.Ints(positions)
		documentPositions[wordIndex[word]][documentId] = positions
		documentIds = append(documentIds, documentId)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// TODO: Calculate document score
	// for documentId := range documentIds {
	// }

	return documentIds, nil
}


