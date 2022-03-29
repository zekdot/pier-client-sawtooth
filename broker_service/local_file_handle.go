package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type LocalFileHandle struct {
	metaMap map[string]string
}

func NewLocalFileHandle() (*LocalFileHandle, error) {
	metaMap := make(map[string]string)
	localFileHandle := &LocalFileHandle{
		metaMap: metaMap,
	}
	err := localFileHandle.readFromFile(STORAGE_FILE_NAME)
	if err != nil {
		return nil, err
	}
	return localFileHandle, nil
}

func checkFileExists(filename string) bool {
	var res = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		res = false
	}
	return res
}

func (localFileHandle *LocalFileHandle)writeToFile(filename string) error {
	textBytes,_ := json.Marshal(localFileHandle.metaMap)
	log.Println("write " + string(textBytes))
	err := ioutil.WriteFile(filename, textBytes, 0666)
	return err
}

func (localFileHandle *LocalFileHandle)readFromFile(filename string) error {
	if !checkFileExists(filename) {
		return nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	textBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	json.Unmarshal(textBytes, &localFileHandle.metaMap)
	return nil
}

func (localFileHandle *LocalFileHandle)SetValue(key string, value string) {
	localFileHandle.metaMap[key] = value
	localFileHandle.writeToFile(STORAGE_FILE_NAME)
}

func (localFileHandle *LocalFileHandle)GetValue(key string) string {
	value, res := localFileHandle.metaMap[key]
	if !res {
		return ""
	}
	return value
}
