package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	// "github.com/astaxie/beego/logs"
)

var Document = NewDocument("./data", "./data/markdowns")

const (
	Document_Default_FileName = "README"
	Document_Page_Suffix      = ".md"
)

const (
	Document_Type_Page = 1
	Document_Type_Dir  = 2
	Document_Type_File = 3
	Document_Type_Git  = 4
)

func NewDocument(documentAbsDir string, markdownAbsDir string) *document {
	return &document{
		DocumentAbsDir: documentAbsDir,
		MarkdownAbsDir: markdownAbsDir,
	}
}

type document struct {
	DocumentAbsDir string
	MarkdownAbsDir string
	lock           sync.Mutex
}

type DocTree struct {
	spaceId    string
	user       string
	name       string
	path       string
	docId      int
	parentId   string
	pageFolder string
	absFolder  string
	tempFolder string
	type int
}

// 【修改·增加文档类型为3时不加后缀逻辑】
// get document page file by parentPath
func (d *document) GetPageFileByParentPath(name string, docType int, parentPath string) (pageFile string) {
	if docType == Document_Type_Page {
		pageFile = fmt.Sprintf("%s/%s%s", parentPath, name, Document_Page_Suffix)
	} else if docType == Document_Type_Dir || docType == Document_Type_Git {
		pageFile = fmt.Sprintf("%s/%s/%s%s", parentPath, name, Document_Default_FileName, Document_Page_Suffix)
	} else {
		pageFile = fmt.Sprintf("%s/%s", parentPath, name)
	}
	return
}

//get document path by spaceName
func (d *document) GetDefaultPageFileBySpaceName(name string) string {
	return fmt.Sprintf("%s/%s%s", name, Document_Default_FileName, Document_Page_Suffix)
}

// get document abs pageFile
func (d *document) GetAbsPageFileByPageFile(pageFile string) string {
	return d.MarkdownAbsDir + "/" + pageFile
}

// get document content by pageFile
func (d *document) GetContentByPageFile(pageFile string) (content string, err error) {
	return File.GetFileContents(d.GetAbsPageFileByPageFile(pageFile))
}

// 【新增】create folder
func (d *document) CreateFolder(absPath string) error {
	if absPath == "" {
		return nil
	}
	d.lock.Lock()
	// 应该不用先检查是否存在，MkdirAll不会覆盖
	err := os.MkdirAll(absPath, 0777)
	if err != nil {
		d.lock.Unlock()
		return err
	}
	d.lock.Unlock()
	return nil
}

// 【新增】一定记着释放啊！！！
func (d *document) OpenFile(filename string) *os.File {
	d.lock.Lock()
	dst, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		d.lock.Unlock()
		return nil
	}
	d.lock.Unlock()
	return dst
}

// create document
func (d *document) Create(pageFile string) error {
	if pageFile == "" {
		return nil
	}
	d.lock.Lock()
	absFilePath := d.GetAbsPageFileByPageFile(pageFile)
	absDir := filepath.Dir(absFilePath)
	err := os.MkdirAll(absDir, 0777)

	if err != nil {
		d.lock.Unlock()
		return err
	}
	d.lock.Unlock()
	return File.CreateFile(absFilePath)
}

// create and write document
func (d *document) CreateAndWrite(pageFile string, content string) error {
	if pageFile == "" {
		return nil
	}
	d.lock.Lock()
	absFilePath := d.GetAbsPageFileByPageFile(pageFile)
	absDir := filepath.Dir(absFilePath)
	err := os.MkdirAll(absDir, 0777)
	if err != nil {
		d.lock.Unlock()
		return err
	}
	d.lock.Unlock()
	return File.WriteFile(absFilePath, content)
}

// replace document content
func (d *document) Replace(pageFile string, content string) error {
	if pageFile == "" {
		return nil
	}
	d.lock.Lock()
	absFilePath := d.GetAbsPageFileByPageFile(pageFile)
	absDir := filepath.Dir(absFilePath)
	err := os.MkdirAll(absDir, 0777)
	if err != nil {
		d.lock.Unlock()
		return err
	}
	d.lock.Unlock()
	return ioutil.WriteFile(absFilePath, []byte(content), os.ModePerm)
}

// update document
func (d *document) Update(oldPageFile string, name string, content string, docType int, nameIsChange bool, onlyRename bool) (err error) {

	d.lock.Lock()
	defer d.lock.Unlock()

	absOldPageFile := d.GetAbsPageFileByPageFile(oldPageFile)
	//【修改·某些文件不能更改其内容，例如压缩包、非文本文件】
	if !onlyRename {
		err = ioutil.WriteFile(absOldPageFile, []byte(content), os.ModePerm)
		if err != nil {
			return
		}
	}
	if nameIsChange {
		filePath := filepath.Dir(absOldPageFile)
		if docType == Document_Type_Dir || docType == Document_Type_Git {
			err = os.Rename(filePath, filepath.Dir(filePath)+"/"+name)
		} else {
			err = os.Rename(absOldPageFile, filePath+"/"+name) // +Document_Page_Suffix
		}
		if err != nil {
			return
		}
	}
	return nil
}

func (d *document) UpdateSpaceName(oldSpaceName string, newName string) error {

	d.lock.Lock()
	defer d.lock.Unlock()

	spaceOldDir := d.GetAbsPageFileByPageFile(oldSpaceName)
	spaceNewDir := d.GetAbsPageFileByPageFile(newName)
	if spaceNewDir == spaceOldDir {
		return nil
	}
	err := os.Rename(spaceOldDir, spaceNewDir)
	return err
}

// delete document
func (d *document) Delete(path string, docType int) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	absPageFile := d.GetAbsPageFileByPageFile(path)
	ok, _ := File.PathIsExists(absPageFile)
	if !ok {
		return nil
	}
	if docType == Document_Type_Dir || docType == Document_Type_Git {
		return os.RemoveAll(filepath.Dir(absPageFile))
	}
	return os.Remove(absPageFile)
}

func (d *document) DeleteSpace(name string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	absSpaceDir := d.GetAbsPageFileByPageFile(name)

	ok, _ := File.PathIsExists(absSpaceDir)
	if !ok {
		return nil
	}

	return os.RemoveAll(absSpaceDir)
}

func (d *document) PageIsExists(pageFile string) bool {
	absPageFile := d.GetAbsPageFileByPageFile(pageFile)
	ok, err := File.PathIsExists(absPageFile)
	if err != nil {
		return false
	}
	return ok
}

func (d *document) Move(movePath string, targetPath string, docType int) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	// 云转储时迁移文件目录不在本系统体系
	absOldPageFile := movePath
	if docType != Document_Type_Git {
		absOldPageFile = d.GetAbsPageFileByPageFile(movePath)
	}
	absTargetPageFile := d.GetAbsPageFileByPageFile(targetPath)

	if docType == Document_Type_Dir {
		return os.Rename(filepath.Dir(absOldPageFile), filepath.Dir(absTargetPageFile))
	}
	return os.Rename(absOldPageFile, absTargetPageFile)
}

// delete document attachment
func (d *document) DeleteAttachment(attachments []map[string]string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if len(attachments) == 0 {
		return nil
	}

	// delete attachment file
	for _, attachment := range attachments {
		if len(attachment) == 0 || attachment["path"] == "" {
			continue
		}

		file := filepath.Join(d.DocumentAbsDir, attachment["path"])
		_ = os.Remove(file)
	}
	return nil
}
