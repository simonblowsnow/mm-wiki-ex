
$(document).ready(function () {
    var UP_FILE_COUNT = 0;  // 上传文件数量
    var 

    $('#fileUploads').fileinput({
        minFileSize: null,
        uploadExtraData: function () {
            let pid = $("#tabPanelNew input[name='parent_id']").val();
            let sid = $("#tabPanelNew input[name='space_id']").val();
            return {'parent_id': pid, 'space_id': sid};
        }
    });

    $("#tabPanelUpload").on('click', ".fileinput-upload-button", function() {
        UP_FILE_COUNT = $("#fileUploads").fileinput("getFilesCount");
        console.log(count);
        // ".file-preview-thumbnails>.file-preview-frame"
    });

    // 文件上传成功
    $('#fileUploads').on("fileuploaded", function (event, data, previewId, index) {   // fileuploaded  filepreupload
        const res = data.response;
        if (res.code == 0) {
            return alert(res.message);
        }
        debugger;
        if (res.code && res.redirect && res.redirect != "") {
            setTimeout(function () {
                window.location.href = res.redirect.url;
            }, res.redirect.sleep);
        }
    })

});

