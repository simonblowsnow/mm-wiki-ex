package utils

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/astaxie/beego"
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

// pageFile 通常不是纯目录，故需要取其目录
func GetRepoAbsPath(pageFile string) string {
	webRoot, _ := GetWebRoot()
	return filepath.Join(webRoot, filepath.Dir(pageFile))
}
func GetRepoPageDir(pageFile string) string {
	return filepath.Join("serve", filepath.Dir(pageFile))
}

func GetRepoPageFile(pageFile string, build bool) (string, error) {
	webRoot, _ := GetWebRoot()
	filePath := filepath.Join("serve", pageFile)
	if build {
		_, err := CheckAndCreate(filepath.Join(webRoot, filepath.Dir(pageFile)))
		return filePath, err
	}
	return filePath, nil
}

// 同时检查服务根目录是否存在
func GetWebRoot() (string, error) {
	docRootDir := beego.AppConfig.String("document::root_dir")
	webRoot := filepath.Join(docRootDir, "markdowns/serve")
	return CheckAndCreate(webRoot)
}

// 通用过程，检查目录是否存在
func CheckAndCreate(folder string) (string, error) {
	if flag, _ := File.PathIsExists(folder); !flag {
		if err := Document.CreateFolder(folder); err != nil {
			return folder, errors.New("创建服务目录失败，请联系管理员！")
		}
	}
	return folder, nil
}

func LoadTypeConf() map[string]string {
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
