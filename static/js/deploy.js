$(document).ready(function() {

    var uuid = $("#text-uuid").html();

    // 部署按钮
    $("#btn-deploy").click(function () {
        ajaxShell("../deploy", {uuid: uuid}, function() {
        });
    });

    // 重启按钮
    $("#btn-restart").click(function () {
        ajaxShell("../restart", {uuid: uuid}, function() {
        });
    });

    // 停止按钮
    $("#btn-stop").click(function () {
        ajaxShell("../stop", {uuid: uuid}, function() {
        });
    });

    $(".exe").click(function () {
        $.ajax({
            type: "get",
            dataType: "json",
            contentType: "application/json;charset=utf-8",
            url: "/deploy/Execute",
            success: function (result) {
                console.log("大" + result)
            },
          error: function (XMLHttpRequest, textStatus, errorThrown) {
              alert(errorThrown);
              console.log(errorThrown)
          },
        async: false
        });
    })

    function guid2() {
        function S4() {
            return (((1 + Math.random()) * 0x10000) | 0).toString(16).substring(1);
        }
        return (S4() + S4() + "-" + S4() + "-" + S4() + "-" + S4() + "-" + S4() + S4() + S4());
    }

    // 查看日志
    $(".btn-showlog").click(function () {
        var uuid = guid2()

        var url = $(this).attr("data-wsurl");
        url=url+"?uuid="+uuid
        var websocket = new WebSocket(url);
        $.ajax({
            type: "get",
            contentType: "application/json;charset=utf-8",
            url: "/deploy/Execute?uuid="+uuid,
            async: true
        })
        websocket.onmessage = function(event) {
            var msg = event.data;
            $("#layer-modal .modal-content>div").append(msg);
            console.log(msg)
        };

        $("#layer-modal .modal-content").html("<div style='text-align: left;overflow-x: scroll;overflow:scroll;overflow-y: scroll;scrollbar-face-color: #f0eeef'></div>");
        $("#layer-modal").openModal({
            dismissible: false,
            complete: function () {
                websocket.close();
            }
        });

    });

    $(".btn-clear-log").click(function () {
        $("#layer-modal .modal-content>div").html("")
    })
    
    /**
     *加载分支
     * */
    $("#bra").click(function () {
        $("#bra").html("");
        $.ajax({
                type: "get",
                dataType: "json",
                contentType: "application/json;charset=utf-8",
                url: "/deploy/getBranches",
                success: function (result) {
                    $.each(result, function (index, value) {
                        $("#bra").append("<option value='" + value.Value + "'>" + value.Name + "</option>");
                    })
                },
                error: function (XMLHttpRequest, textStatus, errorThrown) {
                    alert(errorThrown);
                },
                async: false
            }
        );
    })
    /**
     * ajax请求后台运行脚本
     */
    function ajaxShell(url, postData, successCallback) {
        $("#loader-modal").openModal({dismissible: false});
        $.ajax({
            url: url,
            type: "POST",
            data: postData,
            cache: false,
            dataType: "text",
            success: function (data) {
                $("#loader-modal").closeModal();
                $("#layer-modal .modal-content").html(data.replace(/\n/g,"<br>"));
                $("#layer-modal").openModal({dismissible: false});
                if(successCallback) {
                    successCallback();
                }
            },
            error: function () {
                $("#loader-modal").closeModal();
                layerAlert("发生异常，请重试！");
            }
        });
    }


    /**
     * 用于代替alert
     * @param text
     */
    function layerAlert(text) {
        $("#alert-modal .text-alert").html(text);
        $("#alert-modal").openModal({dismissible: false});
    }

});