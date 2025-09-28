package src

import (
	"bytes"
	"errors"
	"os"
	"os/exec"

	"github.com/h2non/filetype"
	"github.com/otiai10/gosseract/v2"
)

func tesseractExtract(content []byte, language Language) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()
	err := client.SetImageFromBytes(content)
	switch language {
	case French:
		client.SetLanguage("fra")
	case English:
		client.SetLanguage("eng")
	}
	if err != nil {
		return "", err
	}
	text, err := client.Text()
	if err != nil {
		return "", err
	}
	return text, nil
}

func ExtractText(filename string, language Language) (string, error) {
	_, err := os.Stat(filename)
	if err != nil {
		return "", err
	}

	// TODO: est ce que ça plante si c'est un dossier, un lien symbolique
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	fileType, err := filetype.Match(content)
	if err != nil {
		return "", err
	}

	switch fileType.MIME.Value {
	case "image/jpeg":
		return tesseractExtract(content, language)
	case "image/png":
		return tesseractExtract(content, language)
	case "application/pdf":
		return mutoolConvert(filename, language)
	}

	return "", nil
}

func mutoolConvert(filename string, language Language) (string, error) {
	_, err := exec.LookPath("mutool")
	if err != nil {
		return "", err
	}

	dirName, err := os.MkdirTemp("", "perdocla")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dirName)

	cmd := exec.Command("mutool", "convert", "-F", "png", "-o", dirName + "/out_%d.png", filename)
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	files, err := os.ReadDir(dirName)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	var loopErrors error
	for _, file := range files {
		// TODO: est ce que ça plante si c'est un dossier, un lien symbolique
		content, err := os.ReadFile(dirName + "/" + file.Name())
		if err != nil {
			loopErrors = errors.Join(loopErrors, err)
			continue
		}

		text, err := tesseractExtract(content, language)
		if err != nil {
			loopErrors = errors.Join(loopErrors, err)
			continue
		}
		buffer.WriteString(" ")
		buffer.WriteString(text)
	}
	if loopErrors != nil {
		return "", loopErrors
	}

	return buffer.String(), nil
}