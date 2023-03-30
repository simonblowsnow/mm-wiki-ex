
$(document).ready(function () {
    
    $('#fileUploads').fileinput({
        minFileSize: null,
        uploadExtraData: function () {
            let pid = $("#tabPanelNew input[name='parent_id']").val();
            let sid = $("#tabPanelNew input[name='space_id']").val();
            return {'parent_id': pid, 'space_id': sid};
        }
    });
    
    $("#fileUploads fileinput-upload").text("上传");

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

