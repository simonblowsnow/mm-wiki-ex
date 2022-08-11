package controllers

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/astaxie/beego"
	"github.com/simonblowsnow/mm-wiki-ex/app/utils"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	// "github.com/astaxie/beego/logs"
)

// 实现多态
type FileReader interface {
	Reader() *tar.Reader
	IsDir() bool
	Copy(fw *os.File) (int64, error)
}
type GzFileReader struct {
	reader *gzip.Reader
	file   *tar.Reader
	info   os.FileInfo
}
type ComFileReader struct {
	reader *os.File
	file   *zip.File
}

func (g *GzFileReader) Reader() *tar.Reader {
	return tar.NewReader(g.reader)
}
func (c *ComFileReader) Reader() *tar.Reader {
	return tar.NewReader(c.reader)
}
func (g *GzFileReader) IsDir() bool {
	return g.info.IsDir()
}
func (c *ComFileReader) IsDir() bool {
	return c.file.FileInfo().IsDir()
}
func (g *GzFileReader) Copy(fw *os.File) (int64, error) {
	return io.Copy(fw, g.file)
}
func (c *ComFileReader) Copy(fw *os.File) (int64, error) {
	f, err := c.file.Open()
	if err != nil {
		return 0, errors.New("Open File Error")
	}
	defer f.Close()
	return io.Copy(fw, f)
}

// decoder := mahonia.NewDecoder("gbk").ConvertString(s)

type FileList struct {
	Names []string `json: names`
	Types []int    `json: types`
}

func (fl *FileList) setValue(i int, fi os.FileInfo, name string) {
	fl.Names[i] = name
	fl.Types[i] = 0
	if fi.IsDir() {
		fl.Types[i] = 1
	}
}

// 因无法提前得知数组长度，只能通过动态添加方式
func (fl *FileList) setValueD(fi os.FileInfo, name string) {
	fl.Names = append(fl.Names, name)
	if fi == nil {
		fl.Types = append(fl.Types, 0)
		return
	}
	isDir := 0
	if fi.IsDir() {
		isDir = 1
	}
	fl.Types = append(fl.Types, isDir)
}

func newFileList(count int) FileList {
	names := make([]string, count)
	types := make([]int, count)
	return FileList{Names: names, Types: types}
}

type Compressor struct {
	pageFile  string
	path      string
	name      string
	ext       string
	folder    string
	fileRoot  string
	serveRoot string
	exist     bool
	docType   int
}

func CreateCompressor(pageFile string) *Compressor {
	absPath := utils.Document.GetAbsPageFileByPageFile(pageFile)
	ext := strings.ToLower(path.Ext(pageFile))
	_, filename := filepath.Split(pageFile)

	return &Compressor{
		pageFile: pageFile,
		path:     absPath,
		name:     filename,
		ext:      ext,
		exist:    false,
	}
}

func (c *Compressor) GetName(pageFile string, ext string, filename string) string {
	if ext == "" {
		ext = strings.ToLower(path.Ext(pageFile))
	}
	if filename == "" {
		_, filename = filepath.Split(pageFile)
	}
	// 获取除后缀外文件名，处理.tar.gz特殊情况
	name := strings.TrimSuffix(filename, path.Ext(pageFile))
	if ext == ".gz" && len(name) >= 4 && strings.ToLower(name[len(name)-4:]) == ".tar" {
		name = name[:len(name)-4]
	}
	return name
}

// 用于线上解压
func (c *Compressor) InitCompress(spaceId string, extract bool) error {
	c.folder = c.GetName(c.pageFile, c.ext, c.name)

	// 检查服务根目录是否存在
	docRootDir := beego.AppConfig.String("document::root_dir")
	wwwRoot := filepath.Join(docRootDir, "markdowns/serve", spaceId)
	flag, _ := utils.File.PathIsExists(wwwRoot)
	if !flag {
		err := utils.Document.CreateFolder(wwwRoot)
		if err != nil {
			return errors.New("创建服务目录失败，请联系管理员！")
		}
	}

	c.fileRoot = filepath.Join(wwwRoot, c.folder)
	c.serveRoot = filepath.Join("serve", spaceId, c.folder)
	// 检测是否存在，也用于获取服务地址
	c.exist, _ = utils.File.PathIsExists(c.fileRoot)
	if extract {
		if c.exist {
			return errors.New("检测到已解压的目录，可能已被解压，若需覆盖请先点击清除后再次操作！")
		}
		if err := utils.Document.CreateFolder(c.fileRoot); err != nil {
			return err
		}
	}
	return nil
}

// ==========================================解压相关=========================================================
func (c *Compressor) GetFileList(extract bool) (FileList, error) {
	var files FileList
	var err error
	ext, absPath := c.ext, c.path
	if ext == ".zip" {
		files, err = c.GetZipFileList(absPath, extract)
	} else if ext == ".tar" {
		files, err = c.GetTarFileList(absPath, extract)
	} else if ext == ".gz" || ext == ".tgz" {
		files, err = c.GetGZipFileList(absPath, extract)
	}
	return files, err
}

// zip文件，仅获取文件列表
func (c *Compressor) GetZipFileList(zipFile string, extract bool) (FileList, error) {
	zr, err := zip.OpenReader(zipFile)
	defer zr.Close()
	if err != nil {
		return FileList{}, err
	}

	fs := newFileList(len(zr.File))
	// TODO：此处遇到node_modules这样的极端目录是否影响效率？
	for i, file := range zr.File {
		fpath := GetFileNameUtf8(file)
		// 解压模式过程
		if extract {
			zr := &ComFileReader{file: file}
			c.ExtractFile(zr, fpath)
			continue
		}
		fs.setValue(i, file.FileInfo(), fpath)
	}
	return fs, nil
}

// tar，仅获取文件列表
func (c *Compressor) GetTarFileList(tarFile string, extract bool) (FileList, error) {
	nf := FileList{}
	fr, err := os.Open(tarFile)
	if err != nil {
		return nf, err
	}
	defer fr.Close()

	return c._GetTarFileList(&ComFileReader{reader: fr}, extract)
}

// tar.gz，仅获取文件列表
func (c *Compressor) GetGZipFileList(gzFile string, extract bool) (FileList, error) {
	fs := FileList{}
	fr, err := os.Open(gzFile)
	if err != nil {
		return fs, err
	}
	defer fr.Close()

	gr, err := gzip.NewReader(fr)
	if err != nil {
		return fs, err
	}
	defer gr.Close()

	// 单纯.gz文件，通常内部为单个文件
	if IsGZ(gzFile) {
		if extract {
			if err := c.CopyGZ(gr); err != nil {
				return fs, err
			}
		}
		fs.setValueD(nil, gr.Name)
		return fs, nil
	}

	return c._GetTarFileList(&GzFileReader{reader: gr}, extract)
}

// 获取tar和tar.gz文件列表通用过程
func (c *Compressor) _GetTarFileList(fr FileReader, extract bool) (FileList, error) {
	fs := FileList{}
	tr := fr.Reader()
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fs, err
		}
		// 解压模式过程
		if extract {
			cr := &GzFileReader{file: tr, info: h.FileInfo()}
			c.ExtractFile(cr, h.Name)
			continue
		}
		fs.setValueD(h.FileInfo(), h.Name)
	}
	return fs, nil
}

// 判断单纯.gz后缀文件
func IsGZ(name string) bool {
	ps := strings.Split(strings.ToLower(name), ".")
	l := len(ps)
	if l > 1 && ps[l-1] == "gz" && ps[l-2] != "tar" {
		return true
	}
	return false
}

func (c *Compressor) ExtractFile(fr FileReader, name string) error {
	dstName := filepath.Join(c.fileRoot, name)
	if fr.IsDir() {
		if err := utils.Document.CreateFolder(dstName); err != nil {
			return err
		}
	} else {
		return CopyFile(fr, dstName)
	}
	return nil
}
func (c *Compressor) CopyGZ(fr *gzip.Reader) error {
	name := fr.Name
	dstName := filepath.Join(c.fileRoot, name)
	fd := utils.Document.OpenFile(dstName)
	if fd == nil {
		return errors.New("Open New File Failed")
	}
	if _, err := io.Copy(fd, fr); err != nil {
		return err
	}
	fd.Close()
	return nil
}

func CopyFile(fr FileReader, dstName string) error {
	fd := utils.Document.OpenFile(dstName)
	if fd == nil {
		return errors.New("Open New File Failed")
	}

	if _, err := fr.Copy(fd); err != nil {
		return err
	}
	fd.Close()
	return nil
}

// ================================================单文件提取===============================================
// tar.gz格式内部有些文件夹可能带有./前缀
func GetInnerFilePath(pageFile string, innerFile string) (string, string) {
	iFile := innerFile
	idx := strings.Index(innerFile, "./")
	if idx != -1 {
		iFile = innerFile[2:]
	}
	pagePath, filename := filepath.Split(pageFile)
	filePath := filepath.Join(pagePath, "_temp", filename, iFile)
	filePath = strings.ReplaceAll(filePath, "\\", "/")
	return iFile, filePath
}

func (c *Compressor) ExtractInnerFile(absPath string, innerFile string, dst *os.File) error {
	var err error
	ext := c.ext
	if ext == ".zip" {
		err = ExtractZipInnerFile(absPath, innerFile, dst)
	} else if ext == ".tar" {
		err = ExtractTarInnerFile(absPath, innerFile, dst)
	} else if ext == ".gz" || ext == ".tgz" {
		err = ExtractGZipInnerFile(absPath, innerFile, dst)
	}
	return err
}

// ZIP解压单个文件
func ExtractZipInnerFile(zipFile string, innerFile string, dst *os.File) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fName := GetFileNameUtf8(f)
		if fName != innerFile {
			continue
		}

		src, err := f.Open()
		if err != nil {
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
		src.Close()
		break
	}
	return nil
}

// TAR解压单个文件
func ExtractTarInnerFile(tarFile string, innerFile string, dst *os.File) error {
	fr, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer fr.Close()

	return _ExtractTarInnerFile(nil, fr, innerFile, dst)
}

// GZIP解压单个文件
func ExtractGZipInnerFile(gzFile string, innerFile string, dst *os.File) error {
	fr, err := os.Open(gzFile)
	if err != nil {
		return err
	}
	defer fr.Close()

	gr, err := gzip.NewReader(fr)
	if err != nil {
		return err
	}
	defer gr.Close()

	// 单纯.gz文件，通常内部为单个文件
	if IsGZ(gzFile) {
		if _, err := io.Copy(dst, gr); err != nil {
			return err
		}
		return nil
	}

	return _ExtractTarInnerFile(gr, nil, innerFile, dst)
}

// TAR和GZIP解压单个文件通用过程
func _ExtractTarInnerFile(gr *gzip.Reader, fr *os.File, innerFile string, dst *os.File) error {
	var tr *tar.Reader
	if fr == nil {
		tr = tar.NewReader(gr)
	} else {
		tr = tar.NewReader(fr)
	}

	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if h.Name != innerFile {
			continue
		}

		if _, err := io.Copy(dst, tr); err != nil {
			return err
		}
		break
	}
	return nil
}

// =========================================华丽丽分界线======================================================
//解决文件名乱码问题，如果标示位是0，则是默认的本地编码，默认为gbk
func GetFileNameUtf8(file *zip.File) string {
	if file.Flags == 0 {
		i := bytes.NewReader([]byte(file.Name))
		decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
		content, _ := ioutil.ReadAll(decoder)
		return string(content)
	} else {
		return file.Name
	}
}
