

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
