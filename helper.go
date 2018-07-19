package main

import (
	"archive/zip"
	"bytes"
)

func randomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := range result {
		result[i] = chars[r.Intn(len(chars))]
	}
	return string(result)
}

func createZip(filename string, content string) (zipContent []byte, err error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create(filename)
	if err != nil {
		return
	}
	if _, err = f.Write([]byte(content)); err != nil {
		return
	}
	if err = w.Close(); err != nil {
		return
	}
	return buf.Bytes(), nil
}

func getKeysFromMap(in map[string]string) []string {
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	return keys
}
