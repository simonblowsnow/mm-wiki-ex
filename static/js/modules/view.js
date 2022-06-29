

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
