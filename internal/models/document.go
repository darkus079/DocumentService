package models

import (
	"time"
)

type Document struct {
	ID      string    `json:"id" bson:"_id"`
	Name    string    `json:"name" bson:"name"`
	MIME    string    `json:"mime" bson:"mime"`
	File    bool      `json:"file" bson:"file"`
	Public  bool      `json:"public" bson:"public"`
	Created time.Time `json:"created" bson:"created"`
	Grant   []string  `json:"grant" bson:"grant"`
	Owner   string    `json:"owner" bson:"owner"`
	Content []byte    `json:"-" bson:"content,omitempty"`
}

type DocumentList struct {
	Docs []Document `json:"docs"`
}

type Filter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type QueryParams struct {
	Token string `json:"token"`
	Login string `json:"login,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

type DocumentMeta struct {
	Name   string   `json:"name"`
	File   bool     `json:"file"`
	Public bool     `json:"public"`
	Token  string   `json:"token"`
	MIME   string   `json:"mime"`
	Grant  []string `json:"grant"`
}

type DocumentUploadResponse struct {
	Data struct {
		JSON interface{} `json:"json,omitempty"`
		File string      `json:"file,omitempty"`
	} `json:"data"`
}

type DocumentDeleteResponse struct {
	Response map[string]bool `json:"response"`
}
