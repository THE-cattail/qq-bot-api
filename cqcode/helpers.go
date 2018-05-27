package cqcode

import (
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
)

func NewFileBase64(file interface{}) (string, error) {
	switch f := file.(type) {
	case string:
		data, err := ioutil.ReadFile(f)
		if err != nil {
			return "", err
		}
		encodeString := base64.StdEncoding.EncodeToString(data)
		return "base64://" + encodeString, nil

	case []byte:

		encodeString := base64.StdEncoding.EncodeToString(f)
		return "base64://" + encodeString, nil

	case io.Reader:

		data, err := ioutil.ReadAll(f)
		if err != nil {
			return "", err
		}

		encodeString := base64.StdEncoding.EncodeToString(data)
		return "base64://" + encodeString, nil

	default:
		return "", errors.New("bad file type")
	}
}

func NewFileLocal(file string) string {
	return "file://" + file
}
