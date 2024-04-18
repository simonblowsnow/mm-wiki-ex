package controllers

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
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
	parentId   string
	document   map[string]string
}

var FILETYPES = utils.FILETYPES

// ==================================================== 业务代码开始 ===================================================
// document page view
func (this *PageController) View() {

	style := this.GetString("style", "all")
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
	absPath := utils.Document.GetAbsPageFileByPageFile(pageFile)
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		this.ViewError("目录文档不存在！")
		return
	}

	// 未识别的文件，大小超过限制则不读取其内容，大约6M
	// TODO：文件读取上限大小使用配置文件参数
	if !flag && err == nil && fileInfo.Size() > 7000000 {
		flag = true
	}

	// 仅未识别的文件或md格式文件，读取其内容，md大于10M也不再读
	if (!flag || ext == ".md") && fileInfo.Size() < 12000000 {
		fileExt = ext
		dc, err := utils.Document.GetContentByPageFile(pageFile)
		if err != nil {
			this.ErrorLog("查找文档 " + documentId + " 失败：" + err.Error())
			this.ViewError("文档不存在！")
			return
		}
		documentContent = dc
	}

	// 某些已知的非文本型文件不提供编辑功能
	if fileExt == "other" {
		isEditor = false
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
	host := GetHost(this.Ctx)

	this.Data["location"] = "view"
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

	html := If(style == "all", "page/view", "page/preview")
	this.viewLayout(html.(string), "document_page")
}

// page edit
func (this *PageController) Edit() {

	documentId := this.GetString("document_id", "")
	innerFile := this.GetString("innerFile", "")
	category := this.GetString("category", "")
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

	docType, _ := strconv.Atoi(document["type"])
	// 增加编辑压缩包内解压的文件功能
	this.Data["inner_file"] = "0"
	if innerFile != "" {
		c, err := ComCompress(this.Controller, false)
		if err != nil {
			this.Abort(err.Error())
		}

		isGit := docType == models.Document_Type_Git
		if !c.exist && !isGit {
			this.ViewError("该文件未解压，请先执行在线解压操作！")
		}
		pageFile = filepath.Join(If(isGit, utils.GetRepoPageDir(c.pageFile), c.serveRoot).(string), innerFile)
		document["name"] = innerFile
		this.Data["inner_file"] = "1"
	}

	// get document content
	documentContent := "该类型文件不支持在线编辑其内容！"
	t := category
	notTxt := (t == "pkg" || t == "word" || t == "ppt" || t == "pdf" || t == "image" || t == "video" || t == "other")
	if !notTxt {
		documentContent, err = utils.Document.GetContentByPageFile(pageFile)
		if err != nil {
			this.ErrorLog("查找文档 " + documentId + " 失败：" + err.Error())
			this.ViewError("文档不存在！")
		}
	}
	ext := strings.ToLower(path.Ext(pageFile))
	autoFollowDoc := models.ConfigModel.GetConfigValueByKey(models.ConfigKeyAutoFollowdoc, "0")
	sendEmail := models.ConfigModel.GetConfigValueByKey(models.ConfigKeySendEmail, "0")
	this.Data["sendEmail"] = sendEmail
	this.Data["autoFollowDoc"] = autoFollowDoc
	this.Data["page_content"] = documentContent
	this.Data["document"] = document
	this.Data["location"] = "edit"
	this.Data["file_suffix"] = ext
	this.Data["category"] = category
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
	innerFile := this.GetString("inner_file", "0")
	category := this.GetString("category", "category")

	// rm document_page_editor-markdown-doc
	this.Ctx.Request.PostForm.Del("document_page_editor-markdown-doc")

	if documentId == "" {
		this.jsonError("您没有选择文档！")
	}
	if newName == "" {
		this.jsonError("文档名称不能为空！")
	}
	match, err := regexp.MatchString(`[\\\\/:*?\"<>、|]`, newName)
	if err != nil || (innerFile == "0" && match) {
		this.jsonError("文档名称格式不正确！")
	}
	if innerFile == "0" && newName == utils.Document_Default_FileName {
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

	docType, _ := strconv.Atoi(document["type"])
	isGit, isPkg := docType == models.Document_Type_Git, category == "pkg"
	var c *Compressor
	var errC error
	document["category"] = category
	if innerFile == "1" || isPkg || isGit {
		c, errC = ComCompress(this.Controller, false)
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
		// 将平台默认md文件修改为带md后缀文件时，直接修改文件类型，与系统逻辑保持统一
		if ext := strings.ToLower(path.Ext(newName)); docType == models.Document_Type_Page && ext == ".md" {
			docType = models.Document_Type_File
		}

		// 修改压缩包名称及Git仓库名时，应同步修改其对应资源目录名，否则会丢失关联
		if (isGit && innerFile == "0") || (isPkg && c.exist) {
			resFile := If(isGit, utils.GetRepoPageDir(c.pageFile), c.serveRoot).(string)
			name := c.GetName(newName, "", "")
			targetFile := filepath.Join(filepath.Dir(resFile), name)
			flag := utils.Document.PageIsExists(targetFile)
			if flag {
				this.jsonError("操作失败，该文件对应资源路径下存在与目标文件名相同的目录！")
			}
			err = utils.Document.Move(resFile, targetFile, models.Document_Type_File)
			if err != nil {
				this.jsonError(err.Error())
			}
		}
	}

	// update document and file content
	updateValue := map[string]interface{}{
		"name":         newName,
		"edit_user_id": this.UserId,
		"type":         docType,
	}

	// 【新增】仅用于压缩包内部文件修改逻辑
	if innerFile == "1" {
		if errC != nil || (!c.exist && !isGit) {
			this.jsonError("修改失败，该文件可能未被解压！")
		}
		pageFile := filepath.Join(If(isGit, utils.GetRepoPageDir(c.pageFile), c.serveRoot).(string), newName)
		err = models.DocumentModel.UpdateFile(pageFile, documentContent, updateValue)
		if err != nil {
			this.ErrorLog("修改文档 " + documentId + "," + newName + " 失败：" + err.Error())
			this.jsonError("修改文档失败！")
		}
		this.InfoLog("修改文档 " + documentId + "," + newName + " 成功")
		this.jsonSuccess("文档修改成功！", nil, "/document/index?document_id="+documentId)
		return
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

	this.Data["doc_id"] = documentId
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
	_, fPath := GetInnerFilePath(info.pageFile, innerFile)

	c, err := ComCompress(this.Controller, false)
	if err != nil {
		this.Abort(err.Error())
	}
	// 若有在线解压，则请求在线解压文件夹中的文件，否则请求单个解压的临时文件
	if c.exist {
		fPath = filepath.Join(c.serveRoot, innerFile)
	}
	if c.docType == models.Document_Type_Git {
		fPath = filepath.Join(utils.GetRepoPageDir(c.pageFile), innerFile)
	}

	RequestFile(this, fPath, info)

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
	self.Data["file_suffix"] = ext
	self.Data["file_url"] = host + "/file/" + pageFile
	self.Data["page_content"] = documentContent
}

func GetHost(ctx *context.Context) string {
	data := ctx.Request.Host
	scheme := strings.ToLower(strings.Split(ctx.Request.Proto, "/")[0])
	host := scheme + "://" + data
	return host
}

// DataController PageController 的共同祖先类 beego.Controller
func ComCompress(self beego.Controller, extract bool) (*Compressor, error) {
	documentId := self.GetString("document_id", "")
	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil || len(document) == 0 {
		return nil, errors.New("文件空间查找错误！")
	}
	// pageFile := self.GetString("file", "")	// 最好还是不依赖于前端传输
	_, pageFile, err := models.DocumentModel.GetParentDocumentsByDocument(document)
	if err != nil {
		return nil, err
	}

	c := CreateCompressor(pageFile)
	// Git转储仓储目录无需解压
	c.docType, _ = strconv.Atoi(document["type"])
	if c.docType == models.Document_Type_Git {
		extract = false
	}
	err = c.InitCompress(document["space_id"], extract)
	if extract && err == nil {
		_, err = c.GetFileList(true)
	}
	return c, err
}

// func CompressToDoc(self beego.Controller, extract bool) (*Compressor, error) {
// 	c, err := ComCompress(this.Controller, false)
// 	if err != nil {
// 		this.Abort(err.Error())
// 	}
// 	if !c.exist {
// 		this.Abort("该文件未在线解压，无需清除！")
// 	}
// }

// 在线解压
func (this *DataController) Decompress() {
	if _, err := ComCompress(this.Controller, true); err != nil {
		this.Abort(err.Error())
	}
	this.ServeJSON()
}

// 将本地文件夹及内部所有转存为在线文档
func (this *DataController) FolderToDoc(folder string) error {
	var files FileList

	// idx := len(folder)
	// fs := FileList{}

	logs.Info("Test")

	// err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
	// 	if path != folder {
	// 		fs.setValueD(info, path[idx+1:])

	// 		logs.Info(":::", path)

	// 	}
	// 	return nil
	// })

	files, err := GetLocalFileList(folder)

	if err != nil {
		this.Abort(err.Error())
	}

	this.Data["json"] = files

	return nil
}

// 在线解压转存
func (this *DataController) DecompressToDoc() {
	c, err := ComCompress(this.Controller, false)
	if err != nil {
		this.Abort(err.Error())
	}
	// Git仓库不支持此操作
	if !c.exist || c.docType == models.Document_Type_Git {
		this.Abort("该文件未在线解压，请先点击在线解压按钮！")
	}
	// 检查当前文件夹下是否有同名文件夹或文件
	parent, _ := filepath.Split(c.path)
	folder := filepath.Join(parent, c.folder)
	// flag, _ := utils.File.PathIsExists(folder)
	// if flag {
	// 	this.Abort("操作失败：当前目录下已存在同名文件夹或文件！")
	// }
	if err := utils.Document.CreateFolder(folder); err != nil {
		this.Abort("操作失败！")
	}
	// logs.Error(c.fileRoot)

	this.FolderToDoc(c.fileRoot)

	// logs.Error(c.pageFile)
	// logs.Error(c.name)
	// logs.Error(c.path)
	// logs.Error(c.folder)
	// logs.Error(c.fileRoot)
	// logs.Error(parent)
	// logs.Error(folder)
	// logs.Error(flag)

	// this.Data["json"] = map[string]string{"ext": c.ext, "exist": c.serveRoot, "path": c.path}

	// fmt.Print()
	// fmt.Print()

	this.ServeJSON()
}

// 为解决文本复制问题，在请求压缩包时提前请求一次服务地址，以pre区分
// 获取在线解压后的服务地址
func (this *DataController) GetServeUrl() {
	pre := this.GetString("pre", "")
	c, err := ComCompress(this.Controller, false)
	if err != nil {
		this.Abort(err.Error())
	}
	isGit := c.docType == models.Document_Type_Git
	exist, pageFile := strconv.FormatBool(c.exist), c.serveRoot

	// 在请求查看压缩包文件时，可以未经解压，但取服务地址时需检查异常
	if !c.exist && pre == "" && !isGit {
		this.Abort("该文件未解压，请先执行在线解压操作！")
	}
	if isGit {
		exist, pageFile = "true", utils.GetRepoPageDir(c.pageFile)
	}
	host := GetHost(this.Ctx)
	url := strings.ReplaceAll(host+"/file/"+pageFile, "\\", "/")

	this.Data["json"] = map[string]string{"url": url + "/", "exist": exist}
	this.ServeJSON()
}

// 在线解压清除
func (this *DataController) DelCompress() {
	c, err := ComCompress(this.Controller, false)
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

func GetLocalFileList(folder string) (FileList, error) {
	idx := len(folder)
	fs := FileList{}
	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if path != folder {
			fs.setValueD(info, path[idx+1:])
		}
		return nil
	})
	return fs, err
}

// 提供压缩包文件预览功能	DataController PageController
func (this *DataController) ViewPkg() {
	pageFile := this.GetString("filePath", "")
	docType, _ := this.GetInt("docType")
	c := CreateCompressor(pageFile)

	var files FileList
	var err error
	if docType == models.Document_Type_Git {
		pageFile = utils.GetRepoAbsPath(pageFile)
		files, err = GetLocalFileList(pageFile)
	} else {
		files, err = c.GetFileList(false)
	}

	if err != nil {
		this.Abort(err.Error())
	}
	this.Data["json"] = files
	this.ServeJSON()
}

// 提供压缩包内部文件内容预览功能
// 默认解压到同目录下_temp文件夹下，若该压缩包已被解压，则请求解压后的文件
func (this *DataController) ViewPkgFile() {
	c, err := ComCompress(this.Controller, false)
	if err != nil {
		this.Abort(err.Error())
	}
	innerFile := this.GetString("innerFile", "")
	pageFile, absPath := c.pageFile, c.path
	iFile, _ := GetInnerFilePath(pageFile, innerFile)

	// TODO：暂不检验请求路径合法性
	// fs := strings.Split(strings.ReplaceAll(innerFile, "\\", "/"), "/")
	innerPath, name := filepath.Split(iFile)
	dstPath := filepath.Join(filepath.Dir(absPath), "_temp", c.name, innerPath)
	dstName := filepath.Join(dstPath, name)
	isGit := c.docType == models.Document_Type_Git
	if isGit {
		repoPath := utils.GetRepoAbsPath(pageFile)
		this.Data["json"] = filepath.Join(repoPath, innerFile)
	}

	// 文件已存在或有在线解压则不再重复解压
	flag, _ := utils.File.PathIsExists(dstName)
	if flag || c.exist || isGit {
		this.ServeJSON()
		return
	}
	ok, _ := utils.File.PathIsExists(dstPath)
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

	// c = CreateCompressor(pageFile)
	err = c.ExtractInnerFile(absPath, innerFile, dst)
	dst.Close()

	if err != nil {
		this.Abort(err.Error())
	}
	this.Data["json"] = dstName
	this.ServeJSON()
}

// 将已下载好的云资源，迁移到相应位置并入库
func (this *BaseController) CloneCloud() {
	info, err := GetDirInfo(this)
	if err != nil {
		this.jsonError(err.Error())
	}

	url := this.GetString("url", "")
	folder := this.GetString("folder", "")
	filename := this.GetString("name", "")

	isFile, err := this.GetInt("isFile", -1)
	if isFile == -1 || err != nil {
		this.jsonError("参数有误！")
	}
	if isFile == 1 && filename == "" {
		this.jsonError("文件名不能为空！")
	}
	root := beego.AppConfig.String("document::root_file")
	if root == "" {
		this.jsonError("云存储路径未配置，请联系管理员！")
	}

	baseName := path.Base(url)
	ext := path.Ext(baseName)
	if isFile == 0 {
		if ext != ".git" {
			this.jsonError("这似乎不是一个规范的Git仓库地址！")
		}
		filename = baseName[:len(baseName)-4]
	} else {
		folder += ".tmp"
	}

	match, err := regexp.MatchString(`[\\\\/:*?\"<>、&=|]`, filename)
	if err != nil || filename == "" || match {
		this.jsonError("文档名称格式不正确，名称不能含有特殊字符！")
	}

	// 新文件或目录的路径
	doc := map[string]string{
		"parent_id": info.parentId, "space_id": info.spaceId, "name": "", "type": "4",
		"path": info.document["path"] + "," + info.parentId,
	}
	_, pageFile, _ := models.DocumentModel.GetParentDocumentsByDocument(doc)
	fd, _ := filepath.Split(pageFile)
	newPageFile := filepath.Join(fd, filename)
	absPath := utils.Document.GetAbsPageFileByPageFile(newPageFile)
	if flag, _ := utils.File.PathIsExists(absPath); flag {
		this.jsonError("目标文件已存在，请检查！")
	}

	// 对于仓库类资源，不在原文件体系内转储，置于统一目录下管理，需同时检查是否存在文件体系及服务体系
	if isFile == 0 {
		newPageFile, _ = utils.GetRepoPageFile(newPageFile, true)
		absPath = utils.Document.GetAbsPageFileByPageFile(newPageFile)
		if flag, _ := utils.File.PathIsExists(absPath); flag {
			this.jsonError("目标资源已存在，请检查，或删除后重试！")
		}
	}

	resPath := filepath.Join(root, folder)
	flag, err := utils.File.PathIsExists(resPath)
	if !flag || err != nil {
		this.jsonError("云资源未就绪，或内部存储地址有误！")
	}

	// 移动
	err = utils.Document.Move(resPath, newPageFile, models.Document_Type_Git)
	if err != nil {
		this.jsonError(err.Error())
	}

	// 入库
	docType := If(isFile == 0, models.Document_Type_Git, models.Document_Type_File).(int)
	insertDocument := map[string]interface{}{
		"parent_id":      info.parentId,
		"space_id":       info.spaceId,
		"name":           filename,
		"type":           docType,
		"path":           info.document["path"] + "," + info.parentId,
		"create_user_id": this.UserId,
		"edit_user_id":   this.UserId,
	}
	documentId, err := models.DocumentModel.Insert(insertDocument)
	if err != nil {
		this.ErrorLog("云转储操作失败：" + err.Error())
		this.jsonError("操作失败")
	}
	this.InfoLog("云转储文档 " + utils.Convert.IntToString(documentId, 10) + " 成功")
	this.jsonSuccess("转储文档成功", nil, "/document/index?document_id="+utils.Convert.IntToString(documentId, 10))
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
