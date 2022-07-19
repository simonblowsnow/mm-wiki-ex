package controllers

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
	"github.com/phachon/mm-wiki/app"
	"github.com/phachon/mm-wiki/app/models"
	"github.com/phachon/mm-wiki/app/services"
	"github.com/phachon/mm-wiki/app/utils"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	//	"encoding/json"
	"github.com/astaxie/beego/logs"
)

type PageController struct {
	BaseController
}
type DataController struct {
	beego.Controller
}

type FileList struct {
	Names []string `json: names`
	Types []int    `json: types`
}

// *zip.Filezip.headerFileInfo

// func (f *FileList) setValue(i int, name string, isDir int) {
// 	f.Names[i] = name
// 	f.Types[i] = isDir
// }

func (fl *FileList) setValue(i int, fi os.FileInfo, name string) {
	fl.Names[i] = name
	fl.Types[i] = 0
	if fi.IsDir() {
		fl.Types[i] = 1
	}
}
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

var FILETYPES map[string]string = map[string]string{
	".png":   "image",
	".jpg":   "image",
	".bmp":   "image",
	".svg":   "image",
	".ico":   "image",
	".gif":   "image",
	".tif":   "image",
	".iif":   "image",
	".pcd":   "image",
	".apng":  "image",
	".jpeg":  "image",
	".webp":  "image",
	".doc":   "word",
	".ppt":   "ppt",
	".xls":   "excel",
	".docx":  "word",
	".pptx":  "ppt",
	".xlsx":  "excel",
	".pdf":   "pdf",
	".zip":   "pkg",
	".tar":   "pkg",
	".gz":    "pkg",
	".rar":   "pkg",
	".7z":    "pkg",
	".rar4":  "pkg",
	".mp4":   "video",
	".3gp":   "video",
	".avi":   "video",
	".wmv":   "video",
	".ogv":   "video",
	".flv":   "video",
	".mgp":   "video",
	".mkv":   "video",
	".mov":   "video",
	".m4v":   "video",
	".awf":   "video",
	".asx":   "video",
	".webm":  "video",
	".rmvb":  "video",
	".rm":    "video",
	".htm":   "code",
	".html":  "code",
	".java":  "code",
	".c":     "code",
	".h":     "code",
	".cc":    "code",
	".cpp":   "code",
	".cxx":   "code",
	".py":    "code",
	".js":    "code",
	".css":   "code",
	".sql":   "code",
	".php":   "code",
	".xml":   "code",
	".json":  "code",
	".jsp":   "code",
	".asp":   "code",
	".sh":    "code",
	".bat":   "code",
	".cmd":   "code",
	".ini":   "code",
	".vbs":   "code",
	".yaml":  "code",
	".swift": "code",
	".scss":  "code",
	".scala": "code",
	".ruby":  "code",
	".r":     "code",
	".jl":    "code",
	".cs":    "code",
	".frm":   "code",
	".less":  "code",
	".go":    "code",
	"txt":    "text",
	"wav":    "audio",
	"mp3":    "audio",
	"ogg":    "audio",
	"aac":    "audio",
}

// document page view
func (this *PageController) View() {

	documentId := this.GetString("document_id", "")
	if documentId == "" {
		this.ViewError("文档未找到！")
	}

	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("查找文档 " + documentId + " 失败：" + err.Error())
		this.ViewError("查找文档失败！")
	}
	if len(document) == 0 {
		this.ViewError("文档不存在！")
	}

	spaceId := document["space_id"]
	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("查找文档 " + documentId + " 所在空间失败：" + err.Error())
		this.ViewError("查找文档失败！")
	}
	if len(space) == 0 {
		this.ViewError("文档所在空间不存在！")
	}
	// check space visit_level
	isVisit, isEditor, _ := this.GetDocumentPrivilege(space)
	if !isVisit {
		this.ViewError("您没有权限访问该空间！")
	}

	// get parent documents by document
	parentDocuments, pageFile, err := models.DocumentModel.GetParentDocumentsByDocument(document)
	if err != nil {
		this.ErrorLog("查找父文档失败：" + err.Error())
		this.ViewError("查找父文档失败！")
	}
	if len(parentDocuments) == 0 {
		this.ViewError("父文档不存在！")
	}

	// 判断文件类型
	ext := strings.ToLower(path.Ext(pageFile))
	fileExt, flag := FILETYPES[ext]
	documentContent := ""
	// 已识别的非文本标准格式文件，不读其内容
	if !flag {
		fileExt = ext
		// get document content
		dc, err := utils.Document.GetContentByPageFile(pageFile)
		if err != nil {
			this.ErrorLog("查找文档 " + documentId + " 失败：" + err.Error())
			this.ViewError("文档不存在！")
			return
		}
		documentContent = dc
	}
	logs.Error("Test Content")

	// get edit user and create user
	users, err := models.UserModel.GetUsersByUserIds([]string{document["create_user_id"], document["edit_user_id"]})
	if err != nil {
		this.ErrorLog("查找文档 " + documentId + " 失败：" + err.Error())
		this.ViewError("查找文档失败！")
	}
	if len(users) == 0 {
		this.ViewError("文档创建用户不存在！")
	}

	var createUser = map[string]string{}
	var editUser = map[string]string{}
	for _, user := range users {
		if user["user_id"] == document["create_user_id"] {
			createUser = user
		}
		if user["user_id"] == document["edit_user_id"] {
			editUser = user
		}
	}

	collectionId := "0"
	collection, err := models.CollectionModel.GetCollectionByUserIdTypeAndResourceId(this.UserId, models.Collection_Type_Doc, documentId)
	//	logs.Error("," + "asasas")

	if err != nil {
		this.ErrorLog("查找文档 " + documentId + " 失败：" + err.Error())
		this.ViewError("文档查找失败！")
	}
	if len(collection) > 0 {
		collectionId = collection["collection_id"]
	}

	// 拼接文件网络地址
	href := this.Ctx.Request.Referer()
	u, _ := url.Parse(href)
	host := "http://" + u.Host
	if href[:5] == "https" {
		host = "https://" + u.Host
	}

	this.Data["is_editor"] = isEditor
	this.Data["space"] = space
	this.Data["create_user"] = createUser
	this.Data["edit_user"] = editUser
	this.Data["document"] = document
	this.Data["collection_id"] = collectionId
	this.Data["page_content"] = documentContent
	this.Data["parent_documents"] = parentDocuments
	this.Data["file_type"] = document["type"]
	this.Data["file_path"] = pageFile
	this.Data["file_ext"] = fileExt
	this.Data["file_url"] = host + "/file/" + pageFile
	this.Data["document_id"] = documentId

	this.viewLayout("page/view", "document_page")
}

// page edit
func (this *PageController) Edit() {

	documentId := this.GetString("document_id", "")
	if documentId == "" {
		this.ViewError("文档未找到！")
	}

	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("修改文档 " + documentId + " 失败：" + err.Error())
		this.ViewError("修改文档失败！")
	}
	if len(document) == 0 {
		this.ViewError("文档不存在！")
	}

	spaceId := document["space_id"]
	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("修改文档 " + documentId + " 失败：" + err.Error())
		this.ViewError("修改文档失败！")
	}
	if len(space) == 0 {
		this.ViewError("文档所在空间不存在！")
	}
	// check space visit_level
	_, isEditor, _ := this.GetDocumentPrivilege(space)
	if !isEditor {
		this.ViewError("您没有权限修改该空间下文档！")
	}

	// get parent documents by document
	_, pageFile, err := models.DocumentModel.GetParentDocumentsByDocument(document)
	if err != nil {
		this.ErrorLog("查找父文档失败：" + err.Error())
		this.ViewError("查找父文档失败！")
	}

	// get document content
	documentContent, err := utils.Document.GetContentByPageFile(pageFile)
	if err != nil {
		this.ErrorLog("查找文档 " + documentId + " 失败：" + err.Error())
		this.ViewError("文档不存在！")
	}

	autoFollowDoc := models.ConfigModel.GetConfigValueByKey(models.ConfigKeyAutoFollowdoc, "0")
	sendEmail := models.ConfigModel.GetConfigValueByKey(models.ConfigKeySendEmail, "0")

	this.Data["sendEmail"] = sendEmail
	this.Data["autoFollowDoc"] = autoFollowDoc
	this.Data["page_content"] = documentContent
	this.Data["document"] = document
	this.viewLayout("page/edit", "document_page")
}

// page modify
func (this *PageController) Modify() {

	if !this.IsPost() {
		this.ViewError("请求方式有误！", "/space/index")
	}
	documentId := this.GetString("document_id", "")
	newName := strings.TrimSpace(this.GetString("name", ""))
	documentContent := this.GetString("document_page_editor-markdown-doc", "")
	comment := strings.TrimSpace(this.GetString("comment", ""))
	isNoticeUser := strings.TrimSpace(this.GetString("is_notice_user", "0"))
	isFollowDoc := strings.TrimSpace(this.GetString("is_follow_doc", "0"))

	// rm document_page_editor-markdown-doc
	this.Ctx.Request.PostForm.Del("document_page_editor-markdown-doc")

	if documentId == "" {
		this.jsonError("您没有选择文档！")
	}
	if newName == "" {
		this.jsonError("文档名称不能为空！")
	}
	match, err := regexp.MatchString(`[\\\\/:*?\"<>、|]`, newName)
	if err != nil {
		this.jsonError("文档名称格式不正确！")
	}
	if match {
		this.jsonError("文档名称格式不正确！")
	}
	if newName == utils.Document_Default_FileName {
		this.jsonError("文档名称不能为 " + utils.Document_Default_FileName + " ！")
	}
	//if comment == "" {
	//	this.jsonError("必须输入此次修改的备注！")
	//}

	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("修改文档 " + documentId + " 失败：" + err.Error())
		this.jsonError("保存文档失败！")
	}
	if len(document) == 0 {
		this.jsonError("文档不存在！")
	}

	spaceId := document["space_id"]
	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("修改文档 " + documentId + " 失败：" + err.Error())
		this.jsonError("保存文档失败！")
	}
	if len(space) == 0 {
		this.jsonError("文档所在空间不存在！")
	}
	// check space document privilege
	_, isEditor, _ := this.GetDocumentPrivilege(space)
	if !isEditor {
		this.jsonError("您没有权限修改该空间下文档！")
	}

	// not allow update space document home page name
	if document["parent_id"] == "0" {
		newName = document["name"]
	}
	// check document name
	if newName != document["name"] {
		newDocument, err := models.DocumentModel.GetDocumentByNameParentIdAndSpaceId(newName,
			document["parent_id"], document["space_id"], utils.Convert.StringToInt(document["type"]))
		if err != nil {
			this.ErrorLog("修改文档失败：" + err.Error())
			this.jsonError("保存文档失败！")
		}
		if len(newDocument) != 0 {
			this.jsonError("该文档名称已经存在！")
		}
	}

	// update document and file content
	updateValue := map[string]interface{}{
		"name":         newName,
		"edit_user_id": this.UserId,
	}
	_, err = models.DocumentModel.UpdateDBAndFile(documentId, spaceId, document, documentContent, updateValue, comment)
	if err != nil {
		this.ErrorLog("修改文档 " + documentId + " 失败：" + err.Error())
		this.jsonError("修改文档失败！")
	}

	// send email to follow user
	if isNoticeUser == "1" {
		logInfo := this.GetLogInfoByCtx()
		url := fmt.Sprintf("%s:%d/document/index?document_id=%s", this.Ctx.Input.Site(), this.Ctx.Input.Port(), documentId)
		go func(documentId string, username string, comment string, url string) {
			err := sendEmail(documentId, username, comment, url)
			if err != nil {
				logInfo["message"] = "更新文档时发送邮件通知失败：" + err.Error()
				logInfo["level"] = models.Log_Level_Error
				models.LogModel.Insert(logInfo)
				logs.Error("更新文档时发送邮件通知失败：" + err.Error())
			}
		}(documentId, this.User["username"], comment, url)
	}
	// follow doc
	if isFollowDoc == "1" {
		go func(userId string, documentId string) {
			_, _ = models.FollowModel.FollowDocument(userId, documentId)
		}(this.UserId, documentId)
	}
	// 更新文档索引
	go func(documentId string) {
		_ = services.DocIndexService.ForceUpdateDocIndexByDocId(documentId)
	}(documentId)

	this.InfoLog("修改文档 " + documentId + " 成功")
	this.jsonSuccess("文档修改成功！", nil, "/document/index?document_id="+documentId)
}

// document share display
func (this *PageController) Display() {

	documentId := this.GetString("document_id", "")
	if documentId == "" {
		this.ViewError("文档未找到！")
	}

	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("查找文档 " + documentId + " 失败：" + err.Error())
		this.ViewError("查找文档失败！")
	}
	if len(document) == 0 {
		this.ViewError("文档不存在！")
	}

	// get document space
	spaceId := document["space_id"]
	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("分享文档 " + documentId + " 失败：" + err.Error())
		this.ViewError("保存文档失败！")
	}
	if len(space) == 0 {
		this.ViewError("文档所在空间不存在！")
	}

	// check space is allow display
	if space["is_share"] != fmt.Sprintf("%d", models.Space_Share_True) {
		this.ViewError("该文档不能被分享！")
	}

	// get parent documents by document
	parentDocuments, pageFile, err := models.DocumentModel.GetParentDocumentsByDocument(document)
	if err != nil {
		this.ErrorLog("查找父文档失败：" + err.Error())
		this.ViewError("查找父文档失败！")
	}
	if len(parentDocuments) == 0 {
		this.ViewError("父文档不存在！")
	}

	// get document content
	documentContent, err := utils.Document.GetContentByPageFile(pageFile)
	if err != nil {
		this.ErrorLog("查找文档 " + documentId + " 内容失败：" + err.Error())
		this.ViewError("文档不存在！")
	}

	// get edit user and create user
	users, err := models.UserModel.GetUsersByUserIds([]string{document["create_user_id"], document["edit_user_id"]})
	if err != nil {
		this.ErrorLog("查找文档 " + documentId + " 失败：" + err.Error())
		this.ViewError("查找文档失败！")
	}
	if len(users) == 0 {
		this.ViewError("文档创建用户不存在！")
	}

	var createUser = map[string]string{}
	var editUser = map[string]string{}
	for _, user := range users {
		if user["user_id"] == document["create_user_id"] {
			createUser = user
		}
		if user["user_id"] == document["edit_user_id"] {
			editUser = user
		}
	}

	this.Data["create_user"] = createUser
	this.Data["edit_user"] = editUser
	this.Data["document"] = document
	this.Data["page_content"] = documentContent
	this.Data["parent_documents"] = parentDocuments
	this.viewLayout("page/display", "document_share")
}

// export file
func (this *PageController) Export() {

	documentId := this.GetString("document_id", "")
	if documentId == "" {
		this.ViewError("文档未找到！")
	}

	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("查找文档 " + documentId + " 失败：" + err.Error())
		this.ViewError("查找文档失败！")
	}
	if len(document) == 0 {
		this.ViewError("文档不存在！")
	}

	spaceId := document["space_id"]
	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("查找文档 " + documentId + " 所在空间失败：" + err.Error())
		this.ViewError("查找文档失败！")
	}
	if len(space) == 0 {
		this.ViewError("文档所在空间不存在！")
	}

	// check space document privilege
	isVisit, _, _ := this.GetDocumentPrivilege(space)
	if !isVisit {
		this.ViewError("您没有权限导出该空间下文档！")
	}

	// check space is allow export
	if space["is_export"] != fmt.Sprintf("%d", models.Space_Download_True) {
		this.ViewError("该文档不允许被导出！")
	}

	// get parent documents by document
	parentDocuments, pageFile, err := models.DocumentModel.GetParentDocumentsByDocument(document)
	if err != nil {
		this.ErrorLog("查找父文档失败：" + err.Error())
		this.ViewError("查找父文档失败！")
	}
	if len(parentDocuments) == 0 {
		this.ViewError("父文档不存在！")
	}

	packFiles := []*utils.CompressFileInfo{}

	absPageFile := utils.Document.GetAbsPageFileByPageFile(pageFile)
	// pack document file
	packFiles = append(packFiles, &utils.CompressFileInfo{
		File:       absPageFile,
		PrefixPath: "",
	})

	// get document attachments
	attachments, err := models.AttachmentModel.GetAttachmentsByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("查找文档附件失败：" + err.Error())
		this.ViewError("查找文档附件失败！")
	}
	for _, attachment := range attachments {
		if attachment["path"] == "" {
			continue
		}
		path := attachment["path"]
		attachmentFile := filepath.Join(app.DocumentAbsDir, path)
		packFile := &utils.CompressFileInfo{
			File:       attachmentFile,
			PrefixPath: filepath.Dir(path),
		}
		packFiles = append(packFiles, packFile)
	}
	var dest = fmt.Sprintf("%s/mm_wiki/%s.zip", os.TempDir(), document["name"])
	err = utils.Zipx.PackFile(packFiles, dest)
	if err != nil {
		this.ErrorLog("导出文档附件失败：" + err.Error())
		this.ViewError("导出文档失败！")
	}

	this.Ctx.Output.Download(dest, document["name"]+".zip")
}

func sendEmail(documentId string, username string, comment string, url string) error {

	// get document by documentId
	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil {
		return errors.New("发送邮件通知查找文档失败：" + err.Error())
	}

	// get send email open config
	sendEmailConfig := models.ConfigModel.GetConfigValueByKey(models.ConfigKeySendEmail, "0")
	if sendEmailConfig == "0" {
		return nil
	}

	// get email config
	emailConfig, err := models.EmailModel.GetUsedEmail()
	if err != nil {
		return errors.New("发送邮件通知查找邮件服务器配置失败：" + err.Error())
	}
	if len(emailConfig) == 0 {
		return nil
	}

	// get follow doc user
	follows, err := models.FollowModel.GetFollowsByObjectIdAndType(documentId, models.Follow_Type_Doc)
	if err != nil {
		return errors.New("发送邮件查找关注文档用户失败：" + err.Error())
	}
	if len(follows) == 0 {
		return nil
	}
	userIds := []string{}
	for _, follow := range follows {
		userIds = append(userIds, follow["user_id"])
	}
	users, err := models.UserModel.GetUsersByUserIds(userIds)
	if err != nil {
		return errors.New("发送邮件查找关注文档用户失败：" + err.Error())
	}
	if len(users) == 0 {
		return nil
	}
	emails := []string{}
	for _, user := range users {
		if user["email"] != "" {
			emails = append(emails, user["email"])
		}
	}

	// get parent documents by document
	parentDocuments, pageFile, err := models.DocumentModel.GetParentDocumentsByDocument(document)
	if err != nil {
		return errors.New("查找文档内容失败: " + err.Error())
	}
	if len(parentDocuments) == 0 {
		return errors.New("查找文档内容失败")
	}
	// get document content
	documentContent, err := utils.Document.GetContentByPageFile(pageFile)
	if err != nil {
		return errors.New("查找文档内容失败: " + err.Error())
	}

	if len([]byte(documentContent)) > 500 {
		documentContent = string([]byte(documentContent)[:500])
	}

	documentValue := document
	documentValue["content"] = documentContent
	documentValue["username"] = username
	documentValue["comment"] = comment
	documentValue["url"] = url

	emailTemplate := beego.BConfig.WebConfig.ViewsPath + "/system/email/template.html"
	body, err := utils.Email.MakeDocumentHtmlBody(documentValue, emailTemplate)
	if err != nil {
		return errors.New("发送邮件生成模板失败：" + err.Error())
	}
	// start send email
	return utils.Email.Send(emailConfig, emails, "文档更新通知", body)
}

// ==========================================通用View==========================================
// document page view common, 因去掉了繁琐的验证，可能被用于非法访问
func (this *PageController) ViewCom() {

	documentId := this.GetString("document_id", "")
	document, _ := models.DocumentModel.GetDocumentByDocumentId(documentId)
	// get parent documents by document
	parentDocuments, pageFile, _ := models.DocumentModel.GetParentDocumentsByDocument(document)

	// 判断文件类型
	ext := strings.ToLower(path.Ext(pageFile))
	fileExt, flag := FILETYPES[ext]
	documentContent := ""
	if !flag || fileExt == "code" {
		if !flag {
			fileExt = ext
		}
		dc, _ := utils.Document.GetContentByPageFile(pageFile)
		documentContent = dc
	}

	href := this.Ctx.Request.Referer()
	u, _ := url.Parse(href)
	host := "http://" + u.Host
	if href[:5] == "https" {
		host = "https://" + u.Host
	}

	this.Data["document"] = document
	this.Data["parent_documents"] = parentDocuments
	this.Data["file_type"] = document["type"]
	this.Data["file_path"] = pageFile
	this.Data["file_ext"] = fileExt
	this.Data["file_url"] = host + "/file/" + pageFile
	this.Data["page_content"] = documentContent
	this.viewLayout("page/viewCom", "document_view")
	//
}

// 提供压缩包文件预览功能	DataController PageController
func (this *DataController) ViewPkg() {
	pageFile := this.GetString("filePath", "")
	absPath := utils.Document.GetAbsPageFileByPageFile(pageFile)
	ext := strings.ToLower(path.Ext(pageFile))

	var files FileList
	var err error

	if ext == ".zip" {
		files, err = GetZipFileList(absPath)
	} else if ext == ".tar" {
		files, err = GetTarFileList(absPath)
	} else if ext == ".gz" {
		files, err = GetGZipFileList(absPath)
	}

	if err == nil {
		this.Data["json"] = files
	}
	this.ServeJSON()
}

// 提供压缩包内部文件内容预览功能
// 默认解压到同目录下_temp文件夹下
func (this *DataController) ViewPkgFile() {
	pageFile := this.GetString("file", "")
	innerFile := this.GetString("innerFile", "")
	absPath := utils.Document.GetAbsPageFileByPageFile(pageFile)
	ext := strings.ToLower(path.Ext(pageFile))
	filename := filepath.Base(absPath)
	iFile := innerFile
	// tar.gz格式内部有些文件夹可能带有./前缀
	idx := strings.Index(innerFile, "./")
	if idx != -1 {
		iFile = innerFile[2:]
	}
	// 暂不检验请求路径合法性
	// fs := strings.Split(strings.ReplaceAll(innerFile, "\\", "/"), "/")
	innerPath, name := filepath.Split(iFile)
	dstPath := filepath.Join(filepath.Dir(absPath), "_temp", filename, innerPath)

	ok, _ := utils.File.PathIsExists(dstPath)
	if !ok {
		err := os.MkdirAll(dstPath, 0766)
		if err != nil {
			this.Abort("创建临时目录失败，请联系管理员！")
		}
	}
	dstName := filepath.Join(dstPath, name)

	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		this.Abort("创建目标文件失败，请联系管理员！")
	}

	// ==========================================测试TAR和GZ============================================
	if ext == ".zip" {
		err = ExtractZipInnerFile(absPath, innerFile, dst)
	} else if ext == ".tar" {
		err = ExtractTarInnerFile(absPath, innerFile, dst)
	} else if ext == ".gz" {
		// files, err = GetGZipFileList(absPath)
	}
	dst.Close()
	logs.Error(pageFile)
	logs.Error(innerFile)
	logs.Error(absPath)
	logs.Error(err)

	if err == nil {
		this.Data["json"] = "Hello"
	}
	this.ServeJSON()
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
		// fs.setValueD(nil, gr.Name)
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

// ========================================================================================
func _getFilePath(filename string) string {
	return ""
}

// zip解压单个制定文件
func UnZipOneFile(zipFile string, filename string) (FileList, error) {
	zr, err := zip.OpenReader(zipFile)
	defer zr.Close()
	if err != nil {
		return FileList{}, err
	}

	fs := newFileList(len(zr.File))
	// TODO：此处遇到node_modules这样的极端目录是否影响效率？
	for i, file := range zr.File {
		fpath := GetFileNameUtf8(file)
		fs.setValue(i, file.FileInfo(), fpath)
	}
	return fs, nil
}

// zip文件，仅获取文件列表
func GetZipFileList(zipFile string) (FileList, error) {
	zr, err := zip.OpenReader(zipFile)
	defer zr.Close()
	if err != nil {
		return FileList{}, err
	}

	fs := newFileList(len(zr.File))
	// TODO：此处遇到node_modules这样的极端目录是否影响效率？
	for i, file := range zr.File {
		fpath := GetFileNameUtf8(file)
		fs.setValue(i, file.FileInfo(), fpath)
	}
	return fs, nil
}

// 获取tar和tar.gz文件列表通用过程
// 不支持多态，只能用俩不同类型参数将就一下了
func _GetTarFileList(gr *gzip.Reader, fr *os.File) (FileList, error) {
	fs := FileList{}
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
			return fs, err
		}
		fs.setValueD(h.FileInfo(), h.Name)
	}
	return fs, nil
}

// tar
func GetTarFileList(tarFile string) (FileList, error) {
	nf := FileList{}
	fr, err := os.Open(tarFile)
	if err != nil {
		return nf, err
	}
	defer fr.Close()

	return _GetTarFileList(nil, fr)
}

// tar.gz
func GetGZipFileList(gzFile string) (FileList, error) {
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
		fs.setValueD(nil, gr.Name)
		return fs, nil
	}

	return _GetTarFileList(gr, nil)
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
