package utils

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"encoding/json"

)

var FILETYPES = LoadTypeConf()


func GetFileExt(filename string) string {
	return strings.ToLower(path.Ext(filename))
}

func GetFileType(filename string) string {
	ext := strings.ToLower(path.Ext(filename))
	fileExt, flag := FILETYPES[ext]
	if !flag {
		return "unknown"
	}
	return fileExt
}

func GetTempFilePath(pageFile string) string {
	_, filename := filepath.Split(pageFile)
	absPath := Document.GetAbsPageFileByPageFile(pageFile)
	tempPath := filepath.Join(filepath.Dir(absPath), "_temp", filename)
	return tempPath	
}

func LoadTypeConf() map[string]string{
	conf := make(map[string]string)
	f, err := os.Open("filetypes.json")
	if err != nil {
		return conf
	}
	defer f.Close()
	jsonData, err := ioutil.ReadAll(f) 
	if err != nil {
		return conf
	}
	
    err = json.Unmarshal(jsonData, &conf)
    if err != nil {
        return conf
    }
	return conf
}

