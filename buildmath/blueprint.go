package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"io"
)

func DecodeBluePrint(blueprint []byte) BluePrint {
	buff := make([]byte, 1024*5000)
	n, err := base64.StdEncoding.Decode(buff, []byte(blueprint[1:]))
	if err != nil {
		panic(err)
	}
	var dec bytes.Buffer

	r, err := zlib.NewReader(bytes.NewReader(buff[:n]))
	if err != nil {
		panic(err)
	}
	defer r.Close()

	io.Copy(&dec, r)

	var data BlueprintData
	err = json.Unmarshal(dec.Bytes(), &data)
	if err != nil {
		panic(err)
	}

	return data.Blueprint
}

func EncodeBluePrint(bluePrint BluePrint) []byte {
	data, err := json.Marshal(&bluePrint)
	if err != nil {
		panic(err)
	}
	var buff bytes.Buffer
	_, err = zlib.NewWriter(&buff).Write(data)
	if err != nil {
		panic(err)
	}
	enc := make([]byte, 1024*5000)
	base64.StdEncoding.Encode(enc, buff.Bytes())
	return append([]byte{'0'}, enc[:base64.StdEncoding.EncodedLen(len(buff.Bytes()))]...)
}

type BlueprintData struct {
	Blueprint BluePrint `json:"blueprint"`
}

type Entity struct {
	Direction    int64    `json:"direction"`
	EntityNumber int64    `json:"entity_number"`
	Name         string   `json:"name"`
	Position     Position `json:"position"`
}

type BluePrint struct {
	Entities []Entity `json:"entities"`
	Icons    []Icon   `json:"icons"`
	Item     string   `json:"item"`
	Version  int64    `json:"version"`
}

type Icon struct {
	Index  int64  `json:"index"`
	Signal Signal `json:"signal"`
}

type Signal struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
