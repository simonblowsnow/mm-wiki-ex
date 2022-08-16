/**
 * Form.js 表单提交类
 * 依赖 jquery.form.js
 */

var Form = {

    /**
     * 提示 div
     */
    failedBox: '#failedBox',

    /**
     * 是否在弹框中
     */
    inPopup: false,

    /**
     * ajax submit
     * @param element
     * @param inPopup
     * @returns {boolean}
     */
    ajaxSubmit: function(element, inPopup) {

        if (inPopup) {
            Form.inPopup = true;
        }

        /**
         * 成功信息条
         * @param message
         * @param data
         */
        function successBox(message, data) {
            Common.successBox(Form.failedBox, message)
        }

        /**
         * 失败信息条
         * @param message
         * @param data
         */
        function failed(message, data) {
            Common.errorBox(Form.failedBox, message)
        }

        /**
         * request success
         * @param result
         */
        function response(result) {
            //console.log(result)
            if (result.code == 0) {
                failed(result.message, result.data);
            }
            if (result.code == 1) {
                successBox(result.message, result.data);
            }
            $("body,html").animate({scrollTop:0},300);
            if (result.redirect.url) {
                var sleepTime = result.redirect.sleep || 3000;
                setTimeout(function() {
                    if (Form.inPopup) {
                        parent.location.href = result.redirect.url;
                    } else {
                        location.href = result.redirect.url;
                    }
                }, sleepTime);
            }
        }

        var options = {
            dataType: 'json',
            success: response
        };

        $(element).ajaxSubmit(options);

        return false;
    }
};

// 新增
Form.clickTab = function (idx) {
    
    if (idx == 0) {
        $("#tabPanelNew").show();
        $("#tabPanelUpload").hide();
        $("#tabPanelFork").hide();
        $("#tabNew").addClass("tabActive");
        $("#tabUpload").removeClass("tabActive");
        $("#tabFork").removeClass("tabActive");
    } else if (idx == 1) {
        $("#tabPanelNew").hide();
        $("#tabPanelUpload").show();
        $("#tabPanelFork").hide();
        $("#tabNew").removeClass("tabActive");
        $("#tabUpload").addClass("tabActive");
        $("#tabFork").removeClass("tabActive");
    } else {
        $("#tabPanelNew").hide();
        $("#tabPanelUpload").hide();
        $("#tabPanelFork").show();
        $("#tabNew").removeClass("tabActive");
        $("#tabUpload").removeClass("tabActive");
        $("#tabFork").addClass("tabActive");
    }
}

function checkUrl (url, notice) {
    var reg = /^[hH][tT]{2}[pP][sS]?:\/\/.+\..{1,5}$/;
    var flag = reg.test(url);
    if (!flag && notice) alert("该Url不符合规范，请检查：\r\n" + url);
    return flag;
}
function checkName (name) {
    return name.search(/\?|&|=/) == -1;
}
// 检查网络文件路径，末尾文件名不能含有?&=三种符号
function getUrlName (url) {
    if (!checkUrl(url)) return false;
    url = url.replace(/\\/g, '/');
    var ps = url.split("/");
    var name = ps[ps.length - 1];
    if (!checkName(name)) return false;
    return name;
}


Form.changeUrl = function (url) {
    var name = getUrlName(url);
    // if (name && $("#fileName").val().trim() == "")
    if (name) {
        $("#fileName").val(name);
    }

}

// 请求下载服务
Form.submitClone = function () {
    var sel = $("#group1 input:radio:checked").val();
    var isFile = (sel == "option1") ? 1 : 0;
    var url = isFile ? $("#fileUrl").val() : $("#gitUrl").val();
    var name = $("#fileName").val();
    if (isFile){
        if (name.trim() == "") return alert("名称不能为空，请检查！");
        if (!checkName(name)) return alert("名称不规范，请检查！");
    }
    if (url == "") return alert("地址不能为空，请检查！");
    if (!checkUrl(url, true)) return; 
    $("#progress").html("Waiting...");
    $("#btnSubmit").attr("disabled","disabled");
    
    var server = "/filecloud/cloneRes";
    var params = {'isFile': isFile, 'url': url};
    $.post(server, params, function (res) {
        if (res.error) return alert(res.message);
        var jobId = res.data.jobId;
        var urlR = "/filecloud/getStatus?jobId=" + jobId;
        createRequest(urlR, isFile, name); 
    }).error(function () {
        alert("网络错误，或服务未启动！");
    });
}

// 请求下载状态
function createRequest (url, isFile, name) {
    var timer = setInterval(function() {
        console.log("request");
        var errorCount = 0;
        $.ajax({url: url, success: function (res) {
            // 检测错误次数，超过5次则终止
            if (res.error) {
                errorCount++;
                if (errorCount > 5) {
                    clearInterval(timer);
                    alert(res.message);
                }
                return;
            } else if (res.data.status == -1) {
                clearInterval(timer);
                $("#btnSubmit").removeAttr("disabled");
                alert("任务失败，请重试，可能是网络超时");
            }

            var status = res.data.status;
            var progress = res.data.progress.replace(",", "，");
            var infos = ["Failed", "Waiting", "Running...", "Finished！"];
            if (progress != "") progress += "， \t ";
            progress += infos[status + 1];
            $("#progress").html(progress);
            if (status == 2) {
                clearInterval(timer);
                console.log(res);
                cloneCloud(res.data.folder, res.data.url, isFile, name);
            }
        }});
    }, 2000);
}

function cloneCloud (folder, url, isFile, name) {
    var spaceId = $("#spaceId").val();
    var parentId = $("#parentId").val();
    var params = {"folder": folder, "url": url, "isFile": isFile, "name": name, "space_id": spaceId, "parent_id": parentId};
    $.post("/CloneCloud", params, function (res) {
        $("#btnSubmit").removeAttr("disabled");
        if (res.code == 0) {
            return alert(res.message);
        }
        if (res.code && res.redirect && res.redirect != "") {
            setTimeout(function () {
                window.location.href = res.redirect.url;
            }, res.redirect.sleep);
        }
    });
}