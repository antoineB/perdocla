package src

import (
	"os"
	"database/sql"
	"crypto/sha256"
	"github.com/h2non/filetype"
	"strconv"
	"time"
)


func InsertDocument(connection *sql.DB, filename string) (int, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return -1, err
	}

	// TODO: est ce que ça plante si c'est un dossier, un lien symbolique
	content, err := os.ReadFile(filename)
	if err != nil {
		return -1, err
	}

	fileType, err := filetype.Match(content)
	if err != nil {
		return -1, err
	}

	h := sha256.New()
	h.Write([]byte(content))
	bs := h.Sum(nil)

	id := 0
	err = connection.QueryRow(
		`INSERT INTO document(sha256, "binary", filename, mime_type, content_length)
         VALUES($1, $2, $3, $4, $5)
         RETURNING id`,
		bs,
		content,
		filename,
		fileType.MIME.Type,
		info.Size(),
	).Scan(&id)
	// TODO: meilleur erreur pour l'ajout en dupliquer du dossier
	if err != nil {
		return -1, err
	}

	// TODO: variabilisé la langue
	// text, err := ExtractText(filename, French)
	// if err != nil {
	// 	return id, err
	// }
	// connection.Query("INSERT INTO document_inverted_index(document_id, word, positions) VALUES($1, $2, jsonb_")

	return id, nil
}

func addTag(connection *sql.DB, tag string) (int, error) {
	id := -1
	rows, err := connection.Query(
		`SELECT id FROM tag WHERE name = $1`,
		tag,
	)
    if err != nil {
        return -1, err
    }

	if rows.Next() {
		rows.Scan(&id)
		return id, nil
	}

	if err = rows.Err(); err != nil {
		return -1, err
	}

	err = connection.QueryRow(
		`INSERT INTO tag(name)
		 VALUES($1)
		 RETURNING id`,
		tag,
	).Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func AddTagToDocument(connection *sql.DB, documentId int, tag string) error {
	tagId, err := addTag(connection, tag)
	if err != nil {
		return err
	}

	_, err = connection.Exec(
		`INSERT INTO tag_document(tag_id, document_id)
		 VALUES($1, $2)`,
		tagId,
		documentId,
	)
	return err
}

func UpdateDateToDocument(connection *sql.DB, documentId int, date time.Time) error {
	_, err := connection.Exec(
	`UPDATE document
     SET user_date = $1
     WHERE id = $2`,
		date,
		documentId,
	)
	return err
}

func SearchByTag(connection *sql.DB, tag string) ([]int, error) {
	rows, err := connection.Query(
		`SELECT d.id
		 FROM document AS d
		 JOIN tag_document AS td ON td.document_id = d.id
		 JOIN tag AS t ON t.id = td.tag_id
		 WHERE t.name = $1`,
		tag,
	)
	if err != nil {
		return nil, err
	}

    var documents []int
    for rows.Next() {
		var id int
		err = rows.Scan(&id)
        if err != nil {
            return nil, err
        }
        documents = append(documents, id)
    }
	err = rows.Err()
    if err != nil {
        return nil, err
    }

	return documents, nil
}

type Document struct {
	id int
	createdAt time.Time
	filename string
	mimetype string
}

func SearchDocuments(connection *sql.DB, searchText string, date string, mime string, filename string, tags []string) ([]*Document, error) {
	// TODO: searchText n'est pas implementé
	// TODO: date

	count := 1
	args := make([]interface{}, 0)
	sqlQuery := "SELECT d.id, d.created_at, d.filename, d.mime_type"
	fromPart := "FROM document AS d"
	wherePart := "WHERE TRUE"

	if filename != "" {
		wherePart = wherePart + " AND d.filename = $" + strconv.Itoa(count)
		args = append(args, filename)
		count++
	}

	if mime != "" {
		wherePart = wherePart + " AND d.mime_type = $" + strconv.Itoa(count)
		args = append(args, mime)
		count++
	}

	if len(tags) > 0 {
		fromPart = fromPart +
` JOIN tag_document AS td ON d.id = td.document_id
  JOIN tag AS t ON t.id = td.tag_id`

		wherePart = wherePart + " IN("
		for _, tag := range tags {
			wherePart = wherePart + "$" + strconv.Itoa(count)
			args = append(args, tag)
			count++
		}
		wherePart = wherePart + ")"
	}

	sqlQuery = sqlQuery + " " + fromPart + " " + wherePart

	rows, err := connection.Query(
		sqlQuery,
		args...,
	)
	if err != nil {
		return nil, err
	}

    var documents []*Document
    for rows.Next() {
		var id int
		var datetime string
		var filename string
		var mimetype string
		err = rows.Scan(&id, &datetime, &filename, &mimetype)
        if err != nil {
            return nil, err
        }
		createdAt, errDate := time.Parse(time.DateTime, datetime)
		if errDate != nil {
			return nil, errDate
		}
		doc := Document{id: id, createdAt: createdAt, filename: filename, mimetype: mimetype}
        documents = append(documents, &doc)
    }
	err = rows.Err()
    if err != nil {
        return nil, err
    }

	return documents, nil
}

func GetDocument(connection *sql.DB, documentId int) (*Document, error) {
	row := connection.QueryRow(
		"SELECT filename, created_at, filename, mime_type FROM document WHERE id = $1",
		documentId,
	)

	var datetime string
	var filename string
	var mimetype string
	err := row.Scan(&datetime, &filename, &mimetype)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	createdAt, errDate := time.Parse(time.DateTime, datetime)
	if errDate != nil {
		return nil, errDate
	}

	document := Document{id: documentId, createdAt: createdAt, filename: filename, mimetype: mimetype}

	return &document, nil;
}

func CreateDatabaseSchema(connection *sql.DB) error {
	sql1 :=
`CREATE TABLE document(
	id integer PRIMARY KEY,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	user_date TIMESTAMP,
	sha256 bytes UNIQUE,
	"binary" bytes,
	filename string,
	mime_type string,
	content_length integer
);`

	sql2 := "CREATE TABLE tag(id integer PRIMARY KEY, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, name string UNIQUE);"

	sql3 := "CREATE TABLE tag_document(tag_id int, document_id int, UNIQUE(tag_id, document_id));"

	sql4 := `CREATE TABLE "language"(id integer PRIMARY KEY, name string UNIQUE);`

	sql5 := "CREATE TABLE document_text_content(document_id int, language_id int, text string, UNIQUE(document_id, language_id));"

	for _, sqlStmt := range [...]string{sql1, sql2, sql3, sql4, sql5} {
		_, err := connection.Exec(sqlStmt)
		if err != nil {
			return err
		}
	}
	return nil
}