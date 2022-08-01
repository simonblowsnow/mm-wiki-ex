
/**
 * 弹出消息提示框，采用浏览器布局，位于整个页面中央，默认显示3秒
 * 后面的消息会覆盖原来的消息
 * @param message：待显示的消息
 * @param type：消息类型，0：错误消息，1：成功消息
 */
 function ShowMessage(message, type) {
    let messageJQ= $("<div class='showMessage'>" + message + "</div>");
    if (type == 0) {
        messageJQ.addClass("showMessageError");
    } else if (type == 1) {
        messageJQ.addClass("showMessageSuccess");
    }
    // 先将原始隐藏，然后添加到页面，最后以400毫秒的速度下拉显示出来
    messageJQ.hide().appendTo("body").slideDown(400);
    // 4秒之后自动删除生成的元素
    window.setTimeout(function() {
        messageJQ.show().slideUp(400, function() {
            messageJQ.remove();
        })
    }, 4000);
}

// 根据URL下载文件
function DownloadFile(url, name) {
    if (name == undefined) {
        var ps = url.split(/[/\\]/);
        name = ps[ps.length - 1];
    }
    var xhr = new XMLHttpRequest();
    xhr.open('GET', url, true); //get请求，请求地址，是否异步
    xhr.responseType = "blob";  // 返回类型blob
    xhr.onload = function () {  // 请求完成处理函数
        if (this.status === 200) {
            var blob = this.response;   // 获取返回值
            var a = document.createElement('a');
            a.download = name;
            a.href=window.URL.createObjectURL(blob);
            a.click();
        }
    };
    // 发送ajax请求
    xhr.send();
}