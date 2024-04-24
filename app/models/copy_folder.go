package models

import (
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/astaxie/beego/logs"
)

type Node struct {
	data DocFileTree
	next *Node
	pre  *Node
}

type NodeList struct {
	link     *Node
	count    int
	maxCount int
}

func WalkFolder(folder string, parentId int, call Callback) error {
	files, err := os.ReadDir(folder)
	if err != nil {
		return err
	}
	for _, file := range files {
		name := file.Name()
		if file.IsDir() {
			id := call(name, 1, parentId)
			WalkFolder(name, id, call)
		} else {
			call(name, 0, parentId)
		}
	}
	return nil
}

// File copies a single file from src to dst
func (nl *NodeList) File(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// Dir copies a whole directory recursively
func (nl *NodeList) CopyDir(src string, dst string, parent *Node) error {
	var err error
	var fds []os.DirEntry
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	dc := DocFileTree{
		Name:        srcinfo.Name(),
		PageFolder:  path.Join(parent.data.PageFolder, srcinfo.Name()),
		AbsFolder:   dst,
		ServeFolder: src,
		FileType:    Document_Type_Dir,
		HasReadMe:   false,
	}

	node := &Node{data: dc, next: nil, pre: parent}
	nl.link.next = node
	nl.link = node

	if fds, err = os.ReadDir(src); err != nil {
		return err
	}

	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())
		if nl.count++; nl.count > nl.maxCount {
			return errors.New("操作失败：压缩包所含文件数超过限制" + strconv.Itoa(nl.maxCount))
		}

		if fd.IsDir() {
			if err = nl.CopyDir(srcfp, dstfp, node); err != nil {
				logs.Info(err)
				return err
			}
		} else {
			if err = nl.File(srcfp, dstfp); err != nil {
				logs.Info(err)
				return err
			}

			df := DocFileTree{
				Name:        fd.Name(),
				PageFolder:  path.Join(dc.PageFolder, fd.Name()),
				AbsFolder:   dstfp,
				ServeFolder: srcfp,
				FileType:    Document_Type_File,
			}
			// 检查目录下是否有readme文件
			if strings.ToLower(df.Name) == "readme.md" {
				node.data.HasReadMe = true
			}
			if ext := strings.ToLower(path.Ext(df.Name)); ext == ".md" {
				df.FileType = Document_Type_Page
				df.Name = df.Name[:len(df.Name)-3]
			}

			nodeF := &Node{data: df, next: nil, pre: node}
			nl.link.next = nodeF
			nl.link = nodeF
		}
	}
	return nil
}
