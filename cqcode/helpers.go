package cqcode

import (
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"net/url"
)

const (
	CacheEnabled  = 1
	CacheDisabled = 0
)

type NetResource struct {
	Cache int `cq:"cache"`
}

func (r *NetResource) EnableCache() {
	r.Cache = CacheEnabled
}

func (r *NetResource) DisableCache() {
	r.Cache = CacheDisabled
}

type NetImage struct {
	*Image
	*NetResource
}

type NetRecord struct {
	*Record
	*NetResource
}

func NewImageBase64(file interface{}) (*Image, error) {
	fileid, err := NewFileBase64(file)
	if err != nil {
		return &Image{}, err
	}
	return &Image{
		FileID: fileid,
	}, nil
}

func NewRecordBase64(file interface{}) (*Record, error) {
	fileid, err := NewFileBase64(file)
	if err != nil {
		return &Record{}, err
	}
	return &Record{
		FileID: fileid,
	}, nil
}

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

func NewImageLocal(file string) *Image {
	return &Image{
		FileID: NewFileLocal(file),
	}
}

func NewRecordLocal(file string) *Record {
	return &Record{
		FileID: NewFileLocal(file),
	}
}

func NewFileLocal(file string) string {
	return "file://" + file
}

func NewImageWeb(url *url.URL) *NetImage {
	return &NetImage{
		Image: &Image{
			FileID: url.String(),
		},
		NetResource: &NetResource{
			Cache: CacheEnabled,
		},
	}
}

func NewRecordWeb(url *url.URL) *NetRecord {
	return &NetRecord{
		Record: &Record{
			FileID: url.String(),
		},
		NetResource: &NetResource{
			Cache: CacheEnabled,
		},
	}
}
