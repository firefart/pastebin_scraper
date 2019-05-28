package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strconv"
	"time"
)

func randomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := range result {
		result[i] = chars[r.Intn(len(chars))]
	}
	return string(result)
}

func createZip(filename string, content string) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create(filename)
	if err != nil {
		return nil, err
	}
	if _, err = f.Write([]byte(content)); err != nil {
		return nil, err
	}
	if err = w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func getKeysFromMap(in map[string][]string) []string {
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	return keys
}

func dateToString(in string) string {
	i, err := strconv.ParseInt(in, 10, 64)
	if err != nil {
		return fmt.Sprintf("could not parse date %q: %v", in, err)
	}
	if i == 0 {
		return ""
	}
	return time.Unix(i, 0).Local().Format(time.ANSIC)
}
