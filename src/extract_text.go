package src

import (
	"bytes"
	"os"

	"github.com/h2non/filetype"
	"github.com/otiai10/gosseract/v2"
	"github.com/gen2brain/go-fitz"
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
		doc, err := fitz.New("/home/antoine/Documents/Attestation de télétravail.pdf")
		if err != nil {
			return "", err
		}
		defer doc.Close()
		var buffer bytes.Buffer
		for n := 0; n < doc.NumPage(); n++ {
 			img, err := doc.ImagePNG(n, 300)
			if err != nil {
				return "", nil
			}
			text, err := tesseractExtract(img, language)
			if err != nil {
				return "", err
			}
			buffer.WriteString(" ")
			buffer.WriteString(text)
		}
		return buffer.String(), nil
	}

	return "", nil
}
