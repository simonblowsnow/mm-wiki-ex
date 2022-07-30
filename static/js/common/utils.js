
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