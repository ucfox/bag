package es

import (
	"bytes"
	"encoding/json"
)

type EsDecoder struct{}

func (d *EsDecoder) Decode(data []byte, v interface{}) error {
	decode := json.NewDecoder(bytes.NewReader(data))
	decode.UseNumber()
	return decode.Decode(v)
}
