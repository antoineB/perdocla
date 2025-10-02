package src

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"database/sql"
	"fmt"
	"time"
)

type SubCommandRunner interface {
	Run(connection *sql.DB) error
}

type ListCommand struct {
	args []string
}

func (sb ListCommand) Run(connection *sql.DB) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	var search string
	var date string
	var mime string
	var filename string
	var tags string
	fs.StringVar(&search, "search", "", "Search text")
	fs.StringVar(&date, "date", "", "Date of documents")
	fs.StringVar(&mime, "mime", "", "Mime type of the documents")
	fs.StringVar(&filename, "filename", "", "Filename of the document")
	fs.StringVar(&tags, "tags", "", "Tags of the filename")

	fs.Parse(sb.args)

	var tagSplits []string
	if tags == "" {
		tagSplits = []string{}
	} else {
		tagSplits = strings.Split(tags, ",")
	}

	documents, err := SearchDocuments(connection, search, date, mime, filename, tagSplits)
	if err != nil {
		return err
	}

	for _, d := range documents {
		fmt.Println(d.id, ": ", d.filename, " (", d.createdAt, ")")
	}

	return nil
}

type AddCommand struct {
	args []string
}

func ProcessAddingFile(connection *sql.DB, filename string, tags string) error {
	tx, err := connection.Begin()
	if err != nil {
		return err
	}

	id, err := InsertDocument(tx, filename)
	if err != nil {
		defer tx.Rollback()
		return err
	}
	fmt.Printf("Inserted document %s\n", filename)

	if tags != "" {
		var text string
		for tag := range strings.SplitSeq(tags, ",") {
			err = AddTagToDocument(tx, id, tag)
			if err != nil {
				defer tx.Rollback()
				return err
			}
			text = text + "Added tags " + tag + "\n"
		}
	}

	defer tx.Commit()
	return nil
}

func addCommandWalkTree(connection *sql.DB, filename string, tags string) error {
	fi, _ := os.Stat(filename)

	if fi.IsDir() {
		filenames, err := os.ReadDir(filename)
		if err != nil {
			return err
		}
		var loopErrors error
		for _, subFilename := range filenames {
			// TODO: est ce qu'il faut construire le nom
			err := addCommandWalkTree(connection, filename + "/" + subFilename.Name(), tags)
			if err != nil {
				loopErrors = errors.Join(loopErrors, err)
			}
		}
		if loopErrors != nil {
			return loopErrors
		}
	} else {
		err := ProcessAddingFile(connection, filename, tags)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Processing file %s failed with %v\n", filename, err)
			// TODO: return nil ?
			return err
		}
	}
	return nil
}

func (sb AddCommand) Run(connection *sql.DB) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	var tags string
	fs.StringVar(&tags, "tags", "", "Tags to add to the filename")

	fs.Parse(sb.args)

	remainingArgs := fs.Args()
	if len(remainingArgs) == 0 {
		return fmt.Errorf("Missing filename")
	}

	filename := remainingArgs[0]
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "The file %s doesn't exist\n", filename)
		}
		return err
	}

	return addCommandWalkTree(connection, filename, tags)
}

type GetCommand struct {
	args []string
}

func (sb GetCommand) Run(connection *sql.DB) error {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	var tags string
	var date string
	var output string
	var consult bool
	fs.StringVar(&tags, "tags", "", "Tags to add to the document")
	fs.StringVar(&date, "date", "", "Date to add to the document")
	fs.StringVar(&output, "output", "", "File where to write the content of the document")
	fs.BoolVar(&consult, "consult", false, "Consult file using the program associated")

	fs.Parse(sb.args)

	remainingArgs := fs.Args()
	if len(remainingArgs) == 0 {
		return fmt.Errorf("Missing document id")
	}

	id, err := strconv.Atoi(remainingArgs[0])
	if err != nil {
		fmt.Println("There is no document id spécified")
		return nil
	}

	document, err := GetDocument(connection, id)

	if err != nil {
		return err
	}

	if document == nil {
		fmt.Println("No document with this id")
		return nil
	}

	fmt.Println(document.id, ": ", document.filename, " (", document.createdAt, ")")

	if tags != "" {
		tx, err := connection.Begin()
		if err != nil {
			return err
		}
		for tag := range strings.SplitSeq(tags, ",") {
			err = AddTagToDocument(tx, id, tag)
			// TODO: if the tag already exists print an error message and keep going
			if err != nil {
				tx.Rollback()
				return err
			}
		}
		tx.Commit()
	}

	if date != "" {
		parsedDate, errDate := time.Parse(time.DateOnly, date)
		if errDate != nil {
			return errDate
		}
		err := UpdateDateToDocument(connection, id, parsedDate)
		if err != nil {
			return err
		}
	}

	if output != "" {
		fd, err := os.OpenFile(output, os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		defer fd.Close()
		fd.Write(document.binary)
	}

	if consult {
		var executable string
		var commandOptions []string
		err := connection.QueryRow(
			"SELECT executable FROM consult_executable WHERE mime_type = $1",
			document.mimetype,
		).Scan(&executable)
		if err == sql.ErrNoRows {
			switch runtime.GOOS {
			case "windows":
				commandOptions = []string{"cmd", "/C", "start"}
			case "darwin":
				commandOptions = []string{"open", "-n"}
			case "linux":
				commandOptions = []string{"xdg-open"}
			default:
				return fmt.Errorf(
					"Unknown operating system: %s and no executable specified for %s\n",
					runtime.GOOS,
					document.mimetype,
				)
			}

		} else if err != nil {
			return err
		} else {
			commandOptions = []string{executable}
		}

		file, err := os.CreateTemp("", "perdocla")
		if err != nil {
			return err
		}
		tempFilename := file.Name()
		defer os.Remove(tempFilename)

		_, err = file.Write(document.binary)
		file.Close()
		if err != nil {
			return err
		}

		commandOptions = append(commandOptions, tempFilename)

		cmd := exec.Command(commandOptions[0], commandOptions[1:]...)
		err = cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func extractDbArgument(args []string) (string, []string, error) {
	// check if there is the option -db
	for index, arg := range args {
		if arg == "-db" {
			if len(args) == index + 1 {
				return "", args, fmt.Errorf("-db option should be followed by a filename")
			}
			tmp := args[index + 1]
			return tmp, append(args[0:index], args[index + 2:]...), nil
		}
	}
	return "", args, nil
}

func GetConnection(dbName string) (*sql.DB, error) {
	if dbName == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dbName = homeDir + "/perdoc.db"
	}

	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func encryptCommand(dbName string, args []string) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	var keyFilename string
	fs.StringVar(&keyFilename, "key", "", "The filename containing the key to read from or if empty to write to")

	fs.Parse(args)

	if keyFilename == "" {
		return fmt.Errorf("The filename of the key is not set")
	}

	_, err := os.Stat(keyFilename)
	var key []byte
	if os.IsNotExist(err) {
		key, err = generateKey(keyFilename)
		if err != nil {
			return err
		}
	} else {
		key, err = readKeyFile(keyFilename)
		if err != nil {
			return  err
		}
	}

	remainingArgs := fs.Args()
	if len(remainingArgs) == 0 {
		return fmt.Errorf("The filename store content of encryption is not set")
	}

	outputFilename := remainingArgs[0]
	_, err = os.Stat(outputFilename)
	if err == nil {
		return fmt.Errorf("The file %s already exists", outputFilename)
	} else if !os.IsNotExist(err) {
		return err
	}

	content, err := encryptFile(dbName, key)
	if err != nil {
		return err
	}

	err = os.WriteFile(outputFilename, content, 0664)
	if err != nil {
		return err
	}

	return nil
}

func decryptCommand(dbName string, args []string) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	var keyFilename string
	fs.StringVar(&keyFilename, "key", "", "The filename containing the key to read from or if empty to write to")

	fs.Parse(args)

	if keyFilename == "" {
		return fmt.Errorf("The filename of the key is not set")
	}
	key, err := readKeyFile(keyFilename)
	if err != nil {
		return  err
	}

	remainingArgs := fs.Args()
	if len(remainingArgs) == 0 {
		return fmt.Errorf("The filename to read from is not set")
	}

	inputFilename := remainingArgs[0]
	_, err = os.Stat(inputFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("The file %s doesn't exists", inputFilename)
		}
		return err
	}

	content, err := decryptFile(inputFilename, key)
	if err != nil {
		return err
	}

	err = os.WriteFile(dbName, content, 0664)
	if err != nil {
		return err
	}

	return nil
}

func Exec() error {
	args := os.Args[1:]

	dbName, cleanArgs, err := extractDbArgument(args)
	if err != nil {
		return err
	}

	if len(cleanArgs) == 0 {
		return fmt.Errorf("Missing subcommand")
	}
	subCommand := cleanArgs[0]

	if subCommand == "decrypt" {
		// Check that dbName DOESN'T exist
		_, err = os.Stat(dbName)
		if err == nil {
			return fmt.Errorf("The file %s already exists", dbName)
		} else if !os.IsNotExist(err) {
			return err
		}
		return decryptCommand(dbName, cleanArgs[1:])
	}

	// Check that dbName exists
	_, errStat := os.Stat(dbName)

	if subCommand == "encrypt" {
		if errStat != nil {
			return err
		}
		return encryptCommand(dbName, cleanArgs[1:])
	}

	// TODO: Est ce que les valeurs sont utiles
	connection, err := GetConnection(dbName + "?_journal_mode=DELETE&_locking_mode=NORMAL&_txlock=exclusive")
	if err != nil {
		return err
	}
	defer connection.Close()

	if subCommand == "createdb" && (errStat != nil || os.IsNotExist(errStat)) {
		CreateDatabaseSchema(connection)
		return nil
	}

	if errStat != nil {
		return errStat
	}

	var sb SubCommandRunner
	switch subCommand {
	case "list":
		sb = ListCommand{args: cleanArgs[1:]}
	case "add":
		sb = AddCommand{args: cleanArgs[1:]}
	case "get":
		sb = GetCommand{args: cleanArgs[1:]}
	}

	err = sb.Run(connection)
	if err != nil {
		return err
	}

	return nil
}