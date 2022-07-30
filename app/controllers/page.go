package controllers

import (
	// "archive/tar"
	// "archive/zip"
	// "bytes"
	// "compress/gzip"
	// "io"
	// "io/ioutil"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/simonblowsnow/mm-wiki-ex/app"
	"github.com/simonblowsnow/mm-wiki-ex/app/models"
	"github.com/simonblowsnow/mm-wiki-ex/app/services"
	"github.com/simonblowsnow/mm-wiki-ex/app/utils"

	// "golang.org/x/text/encoding/simplifiedchinese"
	// "golang.org/x/text/transform"

	"github.com/astaxie/beego/logs"
)

type PageController struct {
	BaseController
}
type DataController struct {
	beego.Controller
}

type DocInfo struct {
	documentId string
	pageFile   string
	isEditor   bool
	spaceId    string
	document   map[string]string
}

var FILETYPES = utils.FILETYPES

// ==================================================== 业务代码开始 ===================================================
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
	documentContent := "该类型文件不支持预览！"

	// 已识别的非文本标准格式文件，不读其内容
	if !flag || ext == ".md" {
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
	this.Data["file_suffix"] = ext
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

// =====================================================通用View=============================================================
// document page view common, 因去掉了繁琐的验证，可能被用于非法访问
func (this *PageController) ViewCom() {
	info, err := GetDocInfo(this.BaseController)
	if err != nil {
		this.ViewError(err.Error())
	}
	RequestFile(this, info.pageFile, info)

	this.viewLayout("page/viewCom", "document_view")
}

// 压缩包内部文件预览，已提前解压
func (this *PageController) ViewPkgCom() {

	info, err := GetDocInfo(this.BaseController)
	if err != nil {
		this.ViewError(err.Error())
	}
	innerFile := this.GetString("innerFile", "")
	_, filePath := GetInnerFilePath(info.pageFile, innerFile)
	RequestFile(this, filePath, info)

	this.viewLayout("page/viewCom", "document_view")
}

// 文件请求通用过程
func RequestFile(self *PageController, pageFile string, info DocInfo) {
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
	host := GetHost(self.Ctx)

	document := map[string]string{"type": "3"}
	self.Data["space_id"] = info.spaceId
	self.Data["document_id"] = info.documentId
	self.Data["document"] = document
	self.Data["file_type"] = document["type"]
	self.Data["file_path"] = pageFile
	self.Data["file_ext"] = fileExt
	self.Data["file_url"] = host + "/file/" + pageFile
	self.Data["page_content"] = documentContent
}

func GetHost(ctx *context.Context) string {
	href := ctx.Request.Referer()
	u, _ := url.Parse(href)
	host := "http://" + u.Host
	if href[:5] == "https" {
		host = "https://" + u.Host
	}
	return host
}

func ComCompress(self *DataController, extract bool) (*Compressor, error) {
	documentId := self.GetString("document_id", "")
	pageFile := self.GetString("file", "")
	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil || len(document) == 0 {
		return nil, errors.New("文件空间查找错误！")
	}
	c := CreateCompressor(pageFile)
	err = c.InitCompress(document["space_id"], extract)
	if extract && err == nil {
		_, err = c.GetFileList(true)
	}
	return c, err
}

// 在线解压
func (this *DataController) Decompress() {
	if _, err := ComCompress(this, true); err != nil {
		this.Abort(err.Error())
	}
	this.ServeJSON()
}

// 获取在线解压后的服务地址
func (this *DataController) GetServeUrl() {
	pre := this.GetString("pre", "")
	c, err := ComCompress(this, false)
	if err != nil {
		this.Abort(err.Error())
	}
	if !c.exist && pre == "" {
		this.Abort("该文件未解压，请先执行在线解压操作！")
	}

	host := GetHost(this.Ctx)
	url := strings.ReplaceAll(host+"/file/"+c.serveRoot, "\\", "/")
	this.Data["json"] = url
	this.ServeJSON()
}

// 在线解压清除
func (this *DataController) DelCompress() {
	c, err := ComCompress(this, false)
	if err != nil {
		this.Abort(err.Error())
	}
	if !c.exist {
		this.Abort("该文件未在线解压，无需清除！")
	}
	// 后边必须加个斜杠，否则会把上一级目录删除掉
	err = utils.Document.Delete(c.serveRoot+"/", 2)
	if err != nil {
		this.Abort(err.Error())
	}
	this.ServeJSON()
}

// 提供压缩包文件预览功能	DataController PageController
func (this *DataController) ViewPkg() {
	pageFile := this.GetString("filePath", "")
	c := CreateCompressor(pageFile)
	files, err := c.GetFileList(false)

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
	_, filename := filepath.Split(pageFile)
	iFile, _ := GetInnerFilePath(pageFile, innerFile)

	// 暂不检验请求路径合法性
	// fs := strings.Split(strings.ReplaceAll(innerFile, "\\", "/"), "/")
	innerPath, name := filepath.Split(iFile)
	dstPath := filepath.Join(filepath.Dir(absPath), "_temp", filename, innerPath)
	dstName := filepath.Join(dstPath, name)
	// 文件已存在则不再重复解压
	flag, _ := utils.File.PathIsExists(dstName)
	if flag {
		this.ServeJSON()
		return
	}
	ok, err := utils.File.PathIsExists(dstPath)
	if !ok {
		err := utils.Document.CreateFolder(dstPath)
		if err != nil {
			this.Abort("创建临时目录失败，请联系管理员！")
		}
	}
	dst := utils.Document.OpenFile(dstName)
	if dst == nil {
		this.Abort("创建目标文件失败，请联系管理员！")
	}

	c := CreateCompressor(pageFile)
	err = c.ExtractInnerFile(absPath, innerFile, dst)
	dst.Close()

	if err == nil {
		this.Data["json"] = dstName
	}
	this.ServeJSON()
}

// 通用过程
func GetDocInfo(self BaseController) (DocInfo, error) {
	res := DocInfo{}
	documentId := self.GetString("document_id", "")
	if documentId == "" {
		return res, errors.New("文档未找到！")
	}
	res.documentId = documentId
	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil {
		return res, errors.New("查找文档 " + documentId + " 失败：" + err.Error())
	}
	if len(document) == 0 {
		return res, errors.New("文档不存在！")
	}
	res.document = document

	spaceId := document["space_id"]
	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		return res, errors.New("查找文档 " + documentId + " 所在空间失败：" + err.Error())
	}
	if len(space) == 0 {
		return res, errors.New("文档所在空间不存在！")
	}
	res.spaceId = spaceId
	// check space visit_level
	isVisit, isEditor, _ := self.GetDocumentPrivilege(space)
	if !isVisit {
		return res, errors.New("您没有权限访问该空间！")
	}
	res.isEditor = isEditor

	// get parent documents by document
	parentDocuments, pageFile, err := models.DocumentModel.GetParentDocumentsByDocument(document)
	if err != nil {
		return res, errors.New("查找父文档失败：" + err.Error())
	}
	if len(parentDocuments) == 0 {
		return res, errors.New("父文档不存在！")
	}
	res.pageFile = pageFile
	return res, nil
}
