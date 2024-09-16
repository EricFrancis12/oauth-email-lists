package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type FormDataMap map[string]io.Reader

func (fdm FormDataMap) Upload(url string) error {
	// Prepare a form that will be submitted to the url
	var (
		b      = bytes.Buffer{}
		client = http.Client{}
	)

	mpw := multipart.NewWriter(&b)
	for key, rdr := range fdm {
		var wrtr io.Writer
		if clsr, ok := rdr.(io.Closer); ok {
			defer clsr.Close()
		}
		// Add the file
		if file, ok := rdr.(*os.File); ok {
			w, err := mpw.CreateFormFile(key, file.Name())
			if err != nil {
				return err
			}
			wrtr = w
		} else {
			// Add other fields
			w, err := mpw.CreateFormField(key)
			if err != nil {
				return err
			}
			wrtr = w
		}
		_, err := io.Copy(wrtr, rdr)
		if err != nil {
			return err
		}
	}
	mpw.Close()

	req, err := http.NewRequest(http.MethodPost, url, &b)
	if err != nil {
		return err
	}
	req.Header.Set(HTTPHeaderContentType, mpw.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected %d status code, but got %d", http.StatusOK, resp.StatusCode)
	}

	return nil
}
