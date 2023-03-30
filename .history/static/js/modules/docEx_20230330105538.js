
$(document).ready(function () {
    var UP_FILE_COUNT = 0;  // 上传文件数量
    var RETURN_FILE_COUNT = 0;  // 上传返回结果数量

    $('#fileUploads').fileinput({
        minFileSize: null,
        uploadExtraData: function () {
            let pid = $("#tabPanelNew input[name='parent_id']").val();
            let sid = $("#tabPanelNew input[name='space_id']").val();
            return {'parent_id': pid, 'space_id': sid};
        }
    });

    $("#tabPanelUpload").on('click', ".fileinput-upload-button", function() {
        // TODO：有没有可能网络请求先返回，后执行这里的代码？
        RETURN_FILE_COUNT = 0;
        UP_FILE_COUNT = $("#fileUploads").fileinput("getFilesCount");
        console.log("Count:" + UP_FILE_COUNT);
    });

    // 文件上传成功
    $('#fileUploads').on("fileuploaded", function (event, data, previewId, index) {   // fileuploaded  filepreupload
        RETURN_FILE_COUNT += 1;
        console.log(RETURN_FILE_COUNT);
        const res = data.response;
        if (res.code == 0) {
            return alert(res.message);
        }
        if (res.code && res.redirect && res.redirect != "") {
            // 防止在多文件上传时因跳转中断后续上传
            if (RETURN_FILE_COUNT < UP_FILE_COUNT) return;
            setTimeout(function () {
                window.location.href = res.redirect.url;
            }, res.redirect.sleep);
        }
    })

});

