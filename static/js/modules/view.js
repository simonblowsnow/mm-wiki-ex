

function getFileUrl () {
    var path = $("#filePath").val();
    return window.location.origin + "/file/" + path;
}

// 【修改·原核心逻辑，仅支持文本型文件查看】
function TextViewInit () {
    editormd.katexURL = {
        js  : "/static/plugins/editor.md/lib/katex/katex.min",
        css : "/static/plugins/editor.md/lib/katex/katex.min"
    };
    editormd.markdownToHTML("document_page_view", {
        path : '/static/plugins/editor.md/lib/',
        //htmlDecode      : true,       // 开启 HTML 标签解析，为了安全性，默认不开启
        htmlDecode      : "style,script,iframe",  // you can filter tags decode
        //toc             : false,
        tocm            : true,    // Using [TOCM]
        //tocContainer    : "#custom-toc-container", // 自定义 ToC 容器层
        //gfm             : false,
        //tocDropdown     : true,
        //markdownSourceCode : true, // 是否保留 Markdown 源码，即是否删除保存源码的 Textarea 标签
        emoji           : false,
        taskList        : true,
        tex             : true,  // 默认不解析
        flowChart       : true,  // 默认不解析
        sequenceDiagram : true,  // 默认不解析
    });

    // 将所有的 a 标签 target="_parent"
    $("#document_page_view").find("a").each(function () {
        var target = $(this).attr("target") || "_parent";
        $(this).attr("target", target);
    })

    // 生成目录树
    new Toc("document_page_view", {
        clazz: "page-view-dir-toc",
        level: 7,
        top: -2,
        targetId: "page-view-dir-toc"
    });

    // 按钮目录按钮永久显示, 滑动页面时, 固定显示
    window.onscroll = function(){
        var scrollHeight = document.documentElement.scrollTop || document.body.scrollTop;
        var dirPreviewBox = $(".dir-preview-box");
        if (scrollHeight >  120) {
            dirPreviewBox.css({
                position: "fixed",
                top: "20px"
            })
        } else {
            dirPreviewBox.css({
                position: "absolute",
                top: "0"
            })
        }
    };

    // 目录显示与隐藏
    $(".dir-preview-btn").click(function () {
        var pageViewDirToc = $(".page-view-dir-toc");
        if (this.innerText === "展开目录") {
            this.innerText = "隐藏目录";
            pageViewDirToc.show();
            pageViewDirToc.css({
                position: "absolute",
                top: "25px",
                right: 0
            })
        } else {
            this.innerText = "展开目录";
            pageViewDirToc.hide();
        }
    })
}


/*==========================================Class FileTree · 将文件路径列表转为目录树==================================================
    * 原始class写法，以免与整个项目风格不搭。
    * 原始数据样例：["zipTest/fs/", "zipTest/fs/a.txt", "zipTest/zipTest/", "zipTest/zipTest/files/", "zipTest/zipTest/区块链浏览器1.png"]
    *       (注：原始数据来自于压缩包的API输出)
    * Author：chuixue
    */
function FileTree(filelist) {
    this.T = {'children': []};
    this.cache = {};
    // this.Create(filelist);
}
// 无论文件与目录，text初始末尾斜杠先处理掉，否则影响keyLen判断
FileTree.prototype.CreateElement = function (text, path, isDir, i, keyLen) {
    if (i < 0) return this.T;
    // 非末尾路径必定为目录
    var dir = i < (path.length - 1) ? 1 : isDir;
    // key为是否继续下探的依据
    var key = text.substr(0, keyLen);
    if (key in this.cache) return this.cache[key];
    // 递归找到父级，没有则依次重建，直至根节点
    var last = this.CreateElement(text, path, isDir, i - 1, keyLen - path[i].length - 1);
    last.children.push({'text': path[i], 'type': ['file', 'folder'][dir], 'children': []});
    // 缓存记录该节点
    this.cache[key] = last.children[last.children.length - 1];
    return this.cache[key];
}
// 原始Names文件列表中末尾为斜杠的应该是目录，但附加Types数据更严谨些
FileTree.prototype.Create = function (data) {
    data.Names.forEach((e, i) => {
        e = e.replaceAll("\\", "/")
        // 此句防止目录或文件名带斜杠转义之类情况
        var fs = e.split("/");
        // 末尾空字符串是由于原字符串末尾带斜杠
        var flag = fs[fs.length - 1] == "";
        var idx = fs.length - (flag ? 2 : 1);
        var keyLen = flag ? e.length - 1 : e.length;
        this.CreateElement(e, fs, data.Types[i], idx, keyLen);
    });
    return this.T;
};
function FileList2Tree (data) {
    return (new FileTree()).Create(data).children;
}
// ===========================================End Class==============================================

function showUrl (url) {
    layer.open({
        type: 1, skin: Layers.skin, shadeClose: true, shade : 0.6, maxmin: true, 
        title: '<strong>文件地址</strong>', area: ["450px", "200px"],
        content: "<div><div style='height: 40px; margin-top: 25px'>已复制到剪切板，您也可" + 
            "<a target='_blank' href='"+ url + "'>直接点击访问</a>或" + 
            '<a onclick="DownloadFile(' + "'" + url + "'" + ')">下载</a></div>' + 
            "<div><input type='text' style='width: 400px; height: 35px' value='" + url + "'/></div></div>",
        padding: "10px"
    });
}

function setCopy(id, url, callback, noClick) {
    $(id).attr("data-clipboard-text", url);
    var clipboard = new ClipboardJS(id);
    $(id).tipsy({fade: true, gravity: 'n', html: true, title: function () {
        return "<div style='text-align: left; color: #ccc'>点击复制网络地址</div>" + callback();
    }});
    if (noClick) return;
    $(id).click(function () {
        showUrl(callback());
    });
}

// 【增加·文件地址复制功能】
function initView() {
    var fileExt = $("#fileExt").val();
    var filePath = $("#filePath").val();
    var url = getFileUrl();
    // 复制文件地址
    setCopy("#clipBtn", url, function(){
        return getFileUrl();
    });

    // html文件则显示打开按钮
    var ps = url.split(".");
    if (ps.length > 0 && (ps[ps.length - 1] == "html" || ps[ps.length - 1] == "htm")) {
        $("#visit").show();
        $("#visit").on("click", function () {
            window.open(url);
        });
    }

    if ($('#fileView').length > 0) loadFiletree(filePath);
    if ($("#document_page_view").length > 0 && fileExt != ".html") TextViewInit();       
};

function htmlInNextLevel(data) {
    if (data.length > 1) return null;
    var fs = data[0].children;
    for (var i = 0; i < fs.length; i ++) {
        if (fs[i].type == 'file' && (fs[i].text == "index.html" || fs[i].text == "index.htm")) return data[0].text;
    }
    return null;
}

// 用于加载压缩包内的文件树
function loadFiletree(filePath) {
    var docType = $("#fileType").val();
    ajaxData('/ViewPkg', {'filePath': filePath, 'docType': docType}, function (data) {
        var treeData = FileList2Tree(data);
        $('#fileView').jstree({
            'core' : {
                'data': treeData
            },
            "themes": { "stripes": true },
            'types' : {
                'default' : { 'icon' : 'fa fa-folder' },
                'f-open' : { 'icon' : 'fa fa-folder-open fa-fw' },
                'f-closed' : { 'icon' : 'fa fa-folder fa-fw' },
                'folder' : { 'icon' : 'fa fa-folder' },
                'file' : { 'icon' : 'fa fa-file-text' }
            },
            "plugins": ["changed", "types"]
        });
        // 提前获取压缩包的网络服务地址，为复制按钮做准备
        InnerIsServer = htmlInNextLevel(treeData);
        getServeUrl("pre");
    });

    $('#fileView').on("touchstart", ".jstree-anchor", function(e) {
        let node = $(e.currentTarget).parent();
        console.log(node);
        if ($(node).hasClass("jstree-closed")) {
            return $('#fileView').jstree('open_node', node);
        }
        if ($(node).hasClass("jstree-open")) {
            return $('#fileView').jstree('close_node', node);
        }
    });
    
    $("#fileView").on('open_node.jstree', function (event, data) {
        data.instance.set_type(data.node,'f-open');
    });
    $("#fileView").on('close_node.jstree', function (event, data) {
        data.instance.set_type(data.node,'f-closed');
    });
    // 选中节点
    $('#fileView').on("activate_node.jstree", function (obj, e) {
        var fType = e.node.type;
        if (fType != "file") {
            if (!$("#editFile").hasClass("disabled")) $("#editFile").addClass("disabled");
            return
        }

        var paths = [];
        var text = e.node.text;
        e.node.parents.forEach(function(d) {
            var p = $('#fileView').jstree(true).get_node(d);
            if (p.id === "#") return;
            paths.push(p.text);
        });
        paths.reverse();
        paths.push(text);
        var ifPath = paths.join("/");
        var docId = $("#fileId").val();
        
        $("#innerFilePath").val(ifPath);
        $("#editFile").removeClass("disabled");
        viewPkgFile(filePath, docId, ifPath);   
    });
}

// 请求压缩包内的单个文件内容，先请求解压，再请求解压后的文件内容
function viewPkgFile(file, docId, innerFilePath) {
    var param = {'file': file, 'innerFile': innerFilePath, 'document_id': docId};
    ajaxData('/ViewPkgFile', param, function (data) {
        openPkgContent(param);
    });
}

// TODO：使用文件ID代替文件路径，理论上讲相对安全些
function openPkgContent(param) {
    // var content = "/page/ViewPkgCom?file=" + param.file + "&innerFile=" + param.innerFile;
    var fileId = $("#fileId").val();
    var content = "/page/ViewPkgCom?document_id=" + fileId + "&innerFile=" + param.innerFile;
    var sz = ["90%", "90%"];

    layer.open({
        type: 2,
        skin: Layers.skin,
        title: '<strong>文件预览</strong>',
        shadeClose: true,
        shade: 0.6,
        maxmin: true,
        area: sz,
        content: content,
        padding: "10px"
    });
}

function decompressFile() {
    var filePath = $("#filePath").val();
    var fileId = $("#fileId").val();
    var param = {'file': filePath, 'document_id': fileId};
    ajaxData('/Decompress', param, function (data) {
        ShowMessage("解压成功！", 1);
        $("#webUrl,#clearPkg").removeClass("disabled");
    });
}

// 将压缩包解压为文档
function decompressToDoc() {
    let msg = "若压缩包内小文件较多，可能导致执行时间较长！";
    var filePath = $("#filePath").val();
    var fileId = $("#fileId").val();
    var param = {'file': filePath, 'document_id': fileId};
    ajaxData('/Decompress', param, function (data) {
        ShowMessage("迁移成功！", 1);
    });
}

// 获取在线解压后形成的服务地址
// TODO：按钮状态设置==================================================================================
function getServeUrl(pre) {
    var url = $("#fileWebUrl").val();
    var status = $("#fileWebStatus").val();
    var filePath = $("#filePath").val();
    var fileId = $("#fileId").val();
    var param = {'file': filePath, 'document_id': fileId, 'pre': pre || ''};
    ajaxData('/GetServeUrl', param, function (data) {
        var url = data.url;
        // 兼顾原型文件压缩包多一层情况
        if (InnerIsServer) url += InnerIsServer + "/";
        $("#fileWebUrl").val(url);
        if (pre) {
            if (data.exist == "false") {
                $("#webUrl,#clearPkg").addClass("disabled");
            }
            setCopy("#webUrl", url, function () { return url }, true);
        } else {
            $("#fileWebStatus").val("1");
            showUrl(url);
        }
    });
}

function DelCompress() {
    var filePath = $("#filePath").val();
    var fileId = $("#fileId").val();
    var param = {'file': filePath, 'document_id': fileId};
    ajaxData('/DelCompress', param, function (data) {
        ShowMessage("清除成功，服务地址已失效！", 1);
        $("#webUrl,#clearPkg,#editFile").addClass("disabled");
    });
}

function EditInnerFile () {
    var fileId = $("#fileId").val();
    var innerFile = $("#innerFilePath").val();
    if (innerFile == "") return ShowMessage("未选择压缩包内可编辑文件", 0);
    var url = "/page/edit?document_id=" + fileId + "&innerFile=" + innerFile;
    window.location.href = url;
    //alert("保存成功，请注意：该结果仅存储在云端，不会写入压缩包，若点击清除，则所有改动将被清空！");
}

function ajaxData(url, param, callbackSuc, callbackErr) {
    $("#loading").show();
    $.ajax({
        url: url, data: param, type:"get", dataType : "json",      
        success: function(data) {
            if (callbackSuc) callbackSuc(data);
        },
        error: function(err) {
            if (callbackErr) callbackErr(err);
            alert(err.responseText);
        },
        complete: function() {
            $("#loading").hide();
        }
    });
}