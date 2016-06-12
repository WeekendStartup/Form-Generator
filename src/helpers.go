package main

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/nu7hatch/gouuid"
)

func gzipBytes(rawData []byte) ([]byte, error) {
	buffer := bytes.Buffer{}
	gz := gzip.NewWriter(&buffer)
	_, err := gz.Write(rawData)
	if err != nil {
		return nil, err
	}
	err = gz.Flush()
	if err != nil {
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func ungzipBytes(rawData []byte) ([]byte, error) {
	gzippedBuffer := bytes.NewBuffer(rawData)
	reader, err := gzip.NewReader(gzippedBuffer)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	ungzippedBuffer := bytes.Buffer{}
	_, err = io.Copy(&ungzippedBuffer, reader)
	if err != nil {
		return nil, err
	}
	return ungzippedBuffer.Bytes(), nil
}

func generateUUID() (string, error) {
	u4, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	u4S := u4.String()[len(u4.String())-10:]
	return u4S, nil
}
