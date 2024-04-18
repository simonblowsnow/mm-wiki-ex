package controllers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/simonblowsnow/mm-wiki-ex/app/services"

	"github.com/simonblowsnow/mm-wiki-ex/app/models"
	"github.com/simonblowsnow/mm-wiki-ex/app/utils"
	// "github.com/astaxie/beego/logs"
)

type DocumentController struct {
	BaseController
}

// document index
func (this *DocumentController) Index() {

	documentId := this.GetString("document_id", "")
	if documentId == "" {
		this.ViewError("页面参数错误！", "/space/index")
	}

	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("查找空间文档 " + documentId + " 失败：" + err.Error())
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
	isVisit, _, _ := this.GetDocumentPrivilege(space)
	if !isVisit {
		this.ViewError("您没有权限访问该空间下的文档！")
	}

	// get default space document
	spaceDocument, err := models.DocumentModel.GetSpaceDefaultDocument(spaceId)
	if err != nil {
		this.ErrorLog("查找文档 " + documentId + " 失败：" + err.Error())
		this.ViewError("查找文档失败！")
	}
	if len(spaceDocument) == 0 {
		this.ViewError(" 空间首页文档不存在！")
	}

	// get space all document
	documents, err := models.DocumentModel.GetAllSpaceDocuments(spaceId)
	if err != nil {
		this.ErrorLog("查找文档 " + documentId + " 所在空间失败：" + err.Error())
		this.ViewError("查找文档失败！")
	}

	// get space privilege
	_, isEditor, isDelete := this.GetDocumentPrivilege(space)

	this.Data["is_editor"] = isEditor
	this.Data["is_delete"] = isDelete
	this.Data["documents"] = documents
	this.Data["default_document_id"] = documentId
	this.Data["space"] = space
	this.Data["space_document"] = spaceDocument
	this.viewLayout("document/index", "document")
}

// add document
func (this *DocumentController) Add() {

	spaceId := this.GetString("space_id", "0")
	parentId := this.GetString("parent_id", "0")

	if spaceId == "0" {
		this.ViewError("没有选择空间！")
	}
	if parentId == "0" {
		this.ViewError("没有选择上级！")
	}
	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("添加文档失败：" + err.Error())
		this.ViewError("添加文档失败！")
	}
	if len(space) == 0 {
		this.ViewError("空间不存在！")
	}

	// check space document privilege
	_, isEditor, _ := this.GetDocumentPrivilege(space)
	if !isEditor {
		this.ViewError("您没有权限在该空间下创建文档！")
	}

	parentDocument, err := models.DocumentModel.GetDocumentByDocumentId(parentId)
	if err != nil {
		this.ErrorLog("添加文档 " + parentId + " 失败：" + err.Error())
		this.ViewError("添加文档失败！")
	}
	if len(parentDocument) == 0 {
		this.ViewError("父文档不存在！")
	}
	path := parentDocument["path"] + "," + parentId
	// get parent documents by path
	parentDocuments, err := models.DocumentModel.GetParentDocumentsByPath(path)
	if err != nil {
		this.ErrorLog("查找父文档失败：" + err.Error())
		this.ViewError("查找父文档失败！")
	}
	if len(parentDocuments) == 0 {
		this.ViewError("父文档不存在！")
	}

	this.Data["parent_documents"] = parentDocuments
	this.Data["parent_id"] = parentId
	this.Data["space_id"] = spaceId
	this.viewLayout("document/form", "default")

}

// save document
func (this *DocumentController) Save() {

	if !this.IsPost() {
		this.ViewError("请求方式有误！", "/main/index")
	}
	spaceId := strings.TrimSpace(this.GetString("space_id", "0"))
	parentId := strings.TrimSpace(this.GetString("parent_id", "0"))
	docType, _ := this.GetInt("type", models.Document_Type_Page)
	name := strings.TrimSpace(this.GetString("name", ""))

	if spaceId == "0" {
		this.jsonError("没有选择空间！")
	}
	if parentId == "0" {
		this.jsonError("没有选择父文档！")
	}
	if name == "" {
		this.jsonError("文档名称不能为空！")
	}
	match, err := regexp.MatchString(`[\\\\/:*?\"<>、|]`, name)
	if err != nil {
		this.jsonError("文档名称格式不正确！")
	}
	if match {
		this.jsonError("文档名称格式不正确！")
	}
	if name == utils.Document_Default_FileName {
		this.jsonError("文档名称不能为 " + utils.Document_Default_FileName + " ！")
	}

	// 在线创建非md文件，需要在线创建新文件
	isCreateFile := docType == models.Document_Type_Page
	if docType == models.Document_Type_Page && strings.Contains(name, ".") {
		docType = models.Document_Type_File
	}

	if docType < 1 || docType > 4 {
		this.jsonError("文档类型错误！")
	}

	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("创建保存文档失败：" + err.Error())
		this.jsonError("创建文档失败！")
	}
	if len(space) == 0 {
		this.jsonError("空间不存在！")
	}

	// check space document privilege
	_, isEditor, _ := this.GetDocumentPrivilege(space)
	if !isEditor {
		this.jsonError("您没有权限在该空间下创建文档！")
	}

	parentDocument, err := models.DocumentModel.GetDocumentByDocumentId(parentId)
	if err != nil {
		this.ErrorLog("创建保存文档失败：" + err.Error())
		this.jsonError("创建文档失败！")
	}
	if len(parentDocument) == 0 {
		this.jsonError("父文档不存在！")
	}
	if parentDocument["type"] != fmt.Sprintf("%d", models.Document_Type_Dir) {
		this.jsonError("父文档不是目录！")
	}

	// check document name
	document, err := models.DocumentModel.GetDocumentByNameParentIdAndSpaceId(name, parentId, spaceId, docType)
	if err != nil {
		this.ErrorLog("创建保存文档失败：" + err.Error())
		this.jsonError("创建文档失败！")
	}
	if len(document) != 0 {
		this.jsonError("该文档名称已经存在！")
	}

	insertDocument := map[string]interface{}{
		"parent_id":      parentId,
		"space_id":       spaceId,
		"name":           name,
		"type":           docType,
		"path":           parentDocument["path"] + "," + parentId,
		"create_user_id": this.UserId,
		"edit_user_id":   this.UserId,
		"isCreateFile":   isCreateFile,
	}
	documentId, err := models.DocumentModel.Insert(insertDocument)
	if err != nil {
		this.ErrorLog("创建文档失败：" + err.Error())
		this.jsonError("创建文档失败")
	}
	this.InfoLog("创建文档 " + utils.Convert.IntToString(documentId, 10) + " 成功")
	this.jsonSuccess("创建文档成功", nil, "/document/index?document_id="+utils.Convert.IntToString(documentId, 10))
}

// document history
func (this *DocumentController) History() {

	page, _ := this.GetInt("page", 1)
	documentId := this.GetString("document_id", "0")
	number, _ := this.GetRangeInt("number", 10, 10, 100)
	limit := (page - 1) * number

	if documentId == "0" {
		this.ViewError("没有选择文档目录！")
	}

	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("查看文档 " + documentId + " 修改历史失败：" + err.Error())
		this.ViewError("查看文档修改历史失败！")
	}
	if len(document) == 0 {
		this.jsonError("文档不存在！")
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
	isVisit, _, _ := this.GetDocumentPrivilege(space)
	if !isVisit {
		this.ViewError("您没有权限查看该空间修改历史！")
	}

	logDocuments, err := models.LogDocumentModel.GetLogDocumentsByDocumentIdAndLimit(documentId, limit, number)
	if err != nil {
		this.ErrorLog("查看文档 " + documentId + " 修改历史失败：" + err.Error())
		this.ViewError("查看文档修改历史失败！")
	}
	count, err := models.LogDocumentModel.CountLogDocumentsByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("查看文档 " + documentId + " 修改历史失败：" + err.Error())
		this.ViewError("查看文档修改历史失败！")
	}

	userIds := []string{}
	for _, logDocument := range logDocuments {
		userIds = append(userIds, logDocument["user_id"])
	}
	users, err := models.UserModel.GetUsersByUserIds(userIds)
	if err != nil {
		this.ErrorLog("查看文档 " + documentId + " 修改历史失败：" + err.Error())
		this.ViewError("查看文档修改历史失败！")
	}
	for _, logDocument := range logDocuments {
		logDocument["username"] = ""
		for _, user := range users {
			if logDocument["user_id"] == user["user_id"] {
				logDocument["username"] = user["username"]
				break
			}
		}
	}

	this.Data["logDocuments"] = logDocuments
	this.SetPaginator(number, count)
	this.viewLayout("document/history", "default")
}

// move document
func (this *DocumentController) Move() {

	documentId := this.GetString("document_id", "0")
	targetId := this.GetString("target_id", "0")
	moveType := this.GetString("move_type", "") // 同层文件排序

	if documentId == "0" {
		this.jsonError("没有选择文档节点！")
	}
	if targetId == "0" {
		this.jsonError("没有选择目标文档节点！")
	}

	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("查找移动文档失败：" + err.Error())
		this.jsonError("移动文档失败！")
	}
	if len(document) == 0 {
		this.jsonError("文档不存在！")
	}
	if moveType != "next" && moveType != "prev" {
		if document["type"] == fmt.Sprintf("%d", models.Document_Type_Dir) {
			this.jsonError("不能移动文档目录！")
		}
	}

	targetDocument, err := models.DocumentModel.GetDocumentByDocumentId(targetId)
	if err != nil {
		this.ErrorLog("查找目标文档失败：" + err.Error())
		this.jsonError("移动文档失败！")
	}
	if len(targetDocument) == 0 {
		this.jsonError("目标文档不存在！")
	}
	if document["space_id"] != targetDocument["space_id"] {
		this.jsonError("文档和目标文档不在同一空间！")
	}
	if moveType != "next" && moveType != "prev" {
		if targetDocument["type"] != fmt.Sprintf("%d", models.Document_Type_Dir) {
			this.jsonError("目标文档必须是目录！")
		}
	}

	spaceId := document["space_id"]
	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("移动文档失败：" + err.Error())
		this.jsonError("移动文档失败！")
	}
	if len(space) == 0 {
		this.jsonError("文档空间不存在！")
	}
	// check space document privilege
	_, isEditor, _ := this.GetDocumentPrivilege(space)
	if !isEditor {
		this.jsonError("您没有权限移动该空间下的文档！")
	}

	// 排序逻辑：next-移动到目标文档之后 prev-移动到目标文档之前
	if moveType == "next" || moveType == "prev" {
		this.updateDocSequence(moveType, document, targetDocument)
		return
	}

	_, oldPageFile, err := models.DocumentModel.GetParentDocumentsByDocument(document)
	if err != nil {
		this.ErrorLog("移动文档 " + documentId + " 失败：" + err.Error())
		this.jsonError("移动文档失败！")
	}
	newDocument := map[string]string{
		"space_id":  document["space_id"],
		"parent_id": targetId,
		"name":      document["name"],
		"type":      document["type"],
		"path":      targetDocument["path"] + "," + targetId,
	}

	_, newPageFile, err := models.DocumentModel.GetParentDocumentsByDocument(newDocument)
	if err != nil {
		this.ErrorLog("移动文档 " + documentId + " 失败：" + err.Error())
		this.jsonError("移动文档失败！")
	}

	// update database and move document file
	updateValue := map[string]interface{}{
		"parent_id":    targetId,
		"path":         targetDocument["path"] + "," + targetId,
		"edit_user_id": this.UserId,
	}
	// Git仓库移动逻辑
	docType, _ := strconv.Atoi(document["type"])
	if docType == models.Document_Type_Git {
		// 后续没有真正使用到该字段，故使用目录逻辑来代替
		document["type"] = "2"
		oldRepo, _ := utils.GetRepoPageFile(oldPageFile, false)
		newRepo, _ := utils.GetRepoPageFile(newPageFile, false)
		newFolder := filepath.Dir(newRepo)
		if utils.Document.PageIsExists(newFolder) {
			this.jsonError("移动文档失败，目标资源存在同名仓库！")
		}

		newParent := filepath.Dir(newFolder)
		newAbsParent := utils.Document.GetAbsPageFileByPageFile(newParent)
		if flag, _ := utils.File.PathIsExists(newAbsParent); !flag {
			utils.Document.CreateFolder(newAbsParent)
		}
		utils.Document.Move(oldRepo, newRepo, models.Document_Type_Dir)
	}
	_, err = models.DocumentModel.MoveDBAndFile(documentId, spaceId, updateValue,
		oldPageFile, newPageFile, document["type"], "移动文档到 "+targetDocument["name"])
	if err != nil {
		this.ErrorLog("移动文档 " + documentId + " 失败：" + err.Error())
		this.jsonError("移动文档失败！")
	}

	this.InfoLog("移动文档 " + documentId + " 成功")
	this.jsonSuccess("移动文档成功", nil, "/document/index?document_id="+documentId)
}

// 移动文档排序
func (this *DocumentController) updateDocSequence(moveType string, document map[string]string, targetDocument map[string]string) {

	sequence := utils.Convert.StringToInt(targetDocument["sequence"])
	spaceId := targetDocument["space_id"]
	targetDocumentId := targetDocument["document_id"]
	movedDocumentId := document["document_id"]

	updateSequence := sequence
	if moveType == "next" {
		updateSequence = sequence + 1
	}

	// 批量修改序号
	_, err := models.DocumentModel.MoveSequenceBySpaceIdAndGtSequence(spaceId, updateSequence, 1)
	if err != nil {
		this.ErrorLog("移动文档 " + movedDocumentId + "到目标文档 " + targetDocumentId + " " + moveType + " 失败：" + err.Error())
		this.jsonError("移动文档失败！")
	}

	// 修改当前文档的序号
	updateValue := map[string]interface{}{
		"sequence":     updateSequence,
		"edit_user_id": this.UserId,
	}
	_, err = models.DocumentModel.Update(movedDocumentId, updateValue, fmt.Sprintf("移动文档"), spaceId)
	if err != nil {
		this.ErrorLog("移动文档 " + movedDocumentId + "到目标文档 " + targetDocumentId + " " + moveType + " 失败：" + err.Error())
		this.jsonError("移动文档失败！")
	}
	this.jsonSuccess("移动文档成功", "", "/document/index?document_id="+movedDocumentId)
}

// delete document
func (this *DocumentController) Delete() {

	documentId := this.GetString("document_id", "0")

	if documentId == "0" {
		this.jsonError("没有选择文档！")
	}
	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("删除文档失败：" + err.Error())
		this.jsonError("删除文档失败！")
	}
	if len(document) == 0 {
		this.jsonError("文档不存在！")
	}
	docType, _ := strconv.Atoi(document["type"])

	if docType == models.Document_Type_Dir {
		childDocs, err := models.DocumentModel.GetDocumentsByParentId(document["document_id"])
		if err != nil {
			this.ErrorLog("删除文档失败：" + err.Error())
			this.jsonError("删除文档失败！")
		}
		if len(childDocs) > 0 {
			this.jsonError("请先删除或移动目录下所有文档！")
		}
	}
	spaceId := document["space_id"]
	space, err := models.SpaceModel.GetSpaceBySpaceId(document["space_id"])
	if err != nil {
		this.ErrorLog("删除文档失败：" + err.Error())
		this.jsonError("删除文档失败！")
	}
	if len(space) == 0 {
		this.jsonError("文档空间不存在！")
	}
	// check space document privilege
	_, _, isManager := this.GetDocumentPrivilege(space)
	if !isManager {
		this.jsonError("您没有权限删除该空间下的文档！")
	}

	_, pageFile, err := models.DocumentModel.GetParentDocumentsByDocument(document)
	if err != nil {
		this.ErrorLog("删除文档 " + documentId + " 失败：" + err.Error())
		this.jsonError("删除文档失败！")
	}

	// 如果是压缩包，则删除临时缓存
	tp := utils.GetFileType(pageFile)
	if docType == models.Document_Type_File && tp == "pkg" {
		tempPath := utils.GetTempFilePath(pageFile)
		flag, _ := utils.File.PathIsExists(tempPath)
		if flag {
			os.RemoveAll(tempPath)
		}
		// 在线服务的临时文件目录
		c, _ := ComCompress(this.Controller, false)
		if c.exist {
			absRoot := utils.Document.GetAbsPageFileByPageFile(c.serveRoot)
			if flag, _ := utils.File.PathIsExists(absRoot); flag {
				os.RemoveAll(absRoot)
			}
		}
	}
	// 如果是转储仓库，则同步删除资源
	if docType == models.Document_Type_Git {
		tempPath := utils.GetRepoAbsPath(pageFile)
		os.RemoveAll(tempPath)
	}

	err = models.DocumentModel.DeleteDBAndFile(documentId, spaceId, this.UserId, pageFile, document["type"])
	if err != nil {
		this.ErrorLog("删除文档 " + documentId + " 失败：" + err.Error())
		this.jsonError("删除文档失败！")
	}

	// delete attachment
	err = models.AttachmentModel.DeleteAttachmentsDBFileByDocumentId(documentId)
	if err != nil {
		this.ErrorLog("删除文档 " + documentId + " 附件失败：" + err.Error())
	}

	// 删除文档索引
	go func(documentId string) {
		services.DocIndexService.ForceDelDocIdIndex(documentId)
	}(documentId)

	this.InfoLog("删除文档 " + documentId + " 成功")
	this.jsonSuccess("删除文档成功", "", "/document/index?document_id="+document["parent_id"])
}

//=================================================================================
// 注·九成以上代码来自于Save()方法
// upload file
func (this *DocumentController) Upload() {
	if !this.IsPost() {
		this.ViewError("请求方式有误！", "/main/index")
	}
	info, err := GetDirInfo(&this.BaseController)
	if err != nil {
		this.ViewError(err.Error())
	}

	f, h, err := this.GetFile("file1")
	if err != nil {
		this.jsonError("获取上传文件失败！")
	}
	defer f.Close()
	filename := h.Filename

	// check document name "Readme"
	document, err := models.DocumentModel.GetDocumentByNameParentIdAndSpaceId(filename, info.parentId, info.spaceId, 3)
	if err != nil {
		this.ErrorLog("创建保存文档" + filename + "失败：" + err.Error())
		this.jsonError("创建文档失败！")
	}
	if len(document) != 0 {
		this.jsonError("上传失败，该目录下文档名称'" + filename + "'已经存在，请检查！")
	}

	// 获取文件目录绝对路径
	doc := map[string]string{
		"parent_id": info.parentId, "space_id": info.spaceId, "name": "", "type": "1",
		"path": info.document["path"] + "," + info.parentId,
	}
	_, pageFile, _ := models.DocumentModel.GetParentDocumentsByDocument(doc)
	absFilePath := utils.Document.GetAbsPageFileByPageFile(pageFile)
	folder, _ := filepath.Split(absFilePath)

	// TODO：是否需要d.lock.Lock()？
	this.SaveToFile("file1", folder+filename)

	// 入库
	insertDocument := map[string]interface{}{
		"parent_id":      info.parentId,
		"space_id":       info.spaceId,
		"name":           filename,
		"type":           models.Document_Type_File,
		"path":           info.document["path"] + "," + info.parentId,
		"create_user_id": this.UserId,
		"edit_user_id":   this.UserId,
	}

	documentId, err := models.DocumentModel.Insert(insertDocument)
	if err != nil {
		this.ErrorLog("上传文档失败：" + err.Error())
		this.jsonError("上传文档失败")
	}
	this.InfoLog("上传文档 " + utils.Convert.IntToString(documentId, 10) + " 成功")
	this.jsonSuccess("上传文档成功", nil, "/document/index?document_id="+utils.Convert.IntToString(documentId, 10))
}

func (this *DocumentController) UpdateFile() {
	if !this.IsPost() {
		this.ViewError("请求方式有误！", "/main/index")
	}
	fType := this.GetString("fType", "xlsx")
	info, err := GetDocInfo(this.BaseController)
	if err != nil {
		this.ViewError(err.Error())
	}

	absFilePath := utils.Document.GetAbsPageFileByPageFile(info.pageFile)
	folder, name := filepath.Split(absFilePath)
	// json文件是附属文件，无需写入日志，直接返回，且未返回信息给前端
	this.InfoLog("fType: " + fType)
	this.InfoLog("folder: " + absFilePath)

	if fType == "json" {
		this.SaveToFile("file1", folder+"/"+name+".json")
		return
	}

	this.SaveToFile("file1", folder+"/"+name)
	// 日志相关
	updateValue := map[string]interface{}{
		"name": name, "edit_user_id": this.UserId,
	}
	_, err = models.DocumentModel.UpdateDBAndFile(info.documentId, info.spaceId, info.document, "", updateValue, "")
	if err != nil {
		this.ErrorLog("修改文档 " + info.documentId + " 失败：" + err.Error())
		this.jsonError("修改文档失败！")
	}

	this.InfoLog("修改文档 " + info.documentId + " 成功")
	this.jsonSuccess("修改文档成功")
}

func GetDirInfo(this *BaseController) (DocInfo, error) {
	res := DocInfo{}
	spaceId := strings.TrimSpace(this.GetString("space_id", "0"))
	parentId := strings.TrimSpace(this.GetString("parent_id", "0"))

	if spaceId == "0" || parentId == "0" {
		return res, errors.New("没有选择空间或父文档！")
	}
	space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	if err != nil {
		this.ErrorLog("创建保存文档失败：" + err.Error())
		return res, err
	}
	if len(space) == 0 {
		return res, errors.New("空间不存在！")
	}

	// check space document privilege
	_, isEditor, _ := this.GetDocumentPrivilege(space)
	if !isEditor {
		return res, errors.New("您没有权限在该空间下创建文档！")
	}
	d := models.DocumentModel
	parentDocument, err := d.GetDocumentByDocumentId(parentId)
	if err != nil {
		this.ErrorLog("创建保存文档失败：" + err.Error())
		return res, err
	}
	if len(parentDocument) == 0 {
		return res, errors.New("父文档不存在！")
	}
	if parentDocument["type"] != fmt.Sprintf("%d", models.Document_Type_Dir) {
		return res, errors.New("父文档不是目录！")
	}
	res.spaceId, res.parentId, res.document = spaceId, parentId, parentDocument

	return res, nil
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

func FileIsDir(fi os.FileInfo) int {
	if fi == nil {
		return 0
	}
	if fi.IsDir() {
		return 1
	}
	return 0
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

// 将本地文件夹及内部所有转存为在线文档
func (this *DataController) FolderToDoc(folder string) error {
	// var files FileList
	// check document name "Readme"

	/*
		document, err := models.DocumentModel.GetDocumentByNameParentIdAndSpaceId(filename, info.parentId, info.spaceId, 3)
		if err != nil {
			this.ErrorLog("创建保存文档" + filename + "失败：" + err.Error())
			this.jsonError("创建文档失败！")
		}
		if len(document) != 0 {
			this.jsonError("上传失败，该目录下文档名称'" + filename + "'已经存在，请检查！")
		}


	*/

	// fs := FileList{}
	// id := 0
	// logs.Info("Test")
	// err := WalkFolder(folder, id, func(name string, isDir int, parentId int) int {

	// 	return 0
	// })

	// idx := len(folder)
	// err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
	// 	if path != folder {
	// 		// fs.setValueD(info, path[idx+1:])
	// 		name := path[idx+1:]
	// 		isDir := FileIsDir(info)
	// 		logs.Info(":::", name, isDir)

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

	// this.FolderToDoc(c.fileRoot)

	pageFolder, _ := filepath.Split(c.pageFile)
	pageFile := filepath.Join(pageFolder, c.folder)
	logs.Info(pageFile)

	// this.InsertCompressRoot(c.fileRoot, c.pageFile)

	logs.Error(c.pageFile)
	logs.Error(c.name)
	logs.Error(c.path)
	logs.Error(c.folder)
	logs.Error(c.fileRoot)
	logs.Error(parent)
	logs.Error(folder)
	// logs.Error(flag)

	// this.Data["json"] = map[string]string{"ext": c.ext, "exist": c.serveRoot, "path": c.path}

	// fmt.Print()
	// fmt.Print()

	this.ServeJSON()
}

func (this *DataController) InsertCompressRoot(root string, pageFile string) (err error) {
	documentId := this.GetString("document_id", "")
	if documentId == "" {
		this.Abort("页面参数错误！")
	}

	document, err := models.DocumentModel.GetDocumentByDocumentId(documentId)
	if err != nil {
		this.Abort("查找空间文档 " + documentId + " 失败：" + err.Error())
	}

	if len(document) == 0 {
		this.Abort("文档不存在！")
	}

	logs.Error(pageFile)
	return

	// 创建readme.md======================================================
	err = utils.Document.Create(pageFile)
	if err != nil {

	}

	logs.Info(document)

	userId := document["user_id"]
	// 入库
	insertDocument := map[string]interface{}{
		"parent_id":      document["parentId"],
		"space_id":       document["space_id"],
		"name":           root,
		"type":           models.Document_Type_Dir,
		"path":           document["path"],
		"create_user_id": userId,
		"edit_user_id":   userId,
	}
	logs.Info(insertDocument)

	// space, err := models.SpaceModel.GetSpaceBySpaceId(spaceId)
	// if err != nil {
	// 	logs.Error("查找文档 " + documentId + " 所在空间失败：" + err.Error())
	// 	this.Abort("查找文档失败！")
	// }
	// if len(space) == 0 {
	// 	this.Abort("文档所在空间不存在！")
	// }
	// TODO：此处预留重大漏洞，不检查空间权限

	// get space privilege
	// _, isEditor, isDelete := this.GetDocumentPrivilege(space)

	// this.Data["is_editor"] = isEditor
	// this.Data["is_delete"] = isDelete
	// this.Data["documents"] = documents
	// this.Data["default_document_id"] = documentId
	// this.Data["space"] = space
	// this.Data["space_document"] = spaceDocument
	return
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
