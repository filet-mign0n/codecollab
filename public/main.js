$(function() {
    $("button").attr("disabled", true);
    
    $("#inputname").on("keyup", function(e) {
        if ($(this).val().length != 0) {
                $("button").attr("disabled", false)
        } else {
            $("button").attr("disabled", true);
        }
    });
    
    $("#inputname").on("keydown", function(e) {
        console.log("input keydown")
        if (e.keyCode == 13 && !$("button").attr("disabled")) {
            $("button").click();
        } else if ($("button").attr("disabled")) {
            $("input").attr("placeholder","    YOUR NAME!");
        }
    });

});

function join(username) {
    if (!window["WebSocket"]) {
        return;
    }
    
    $(".modal").hide();
    
    populate();

    function populate() {

        var textarea = document.getElementById("lines");
        var str = '';
        for (var i=1;i < 100;i++) {
            str = str + (i +'\r\n');
        }
        textarea.value = str;
    };

    var content = $("#content");
    var conn = new WebSocket('ws://' + window.location.host + '/ws');

    // Textarea is editable only when socket is opened.
    conn.onopen = function(e) {
        conn.send("C"+username)

        content.attr("disabled", false);
    };

    conn.onclose = function(e) {
        content.attr("disabled", true);
    };

    // Whenever we receive a message, update textarea
    conn.onmessage = function(e) {
        console.log(e, content.val())
        msgKey = e.data.substring(0, 1)
        msgData = e.data.substring(1)
        
        switch(msgKey) {
            case "W":
                writting(msgData)
                break;
            case "M":
                content.val(msgData)
                break;
            case "C":
                user(true, msgData) 
                break;
            case "D":
                user(false, msgData)
                break;
            default:
                console.log("invalid msgKey", e.data)
        }
    };

    var timeoutId = null;
    var typingTimeoutId = null;
    var isTyping = false;

    content.on("keydown", function(e) {
        isTyping = true;
        
        conn.send("W")
        window.clearTimeout(typingTimeoutId);

        // allow natural tab behavior in html textarea
        if (e.keyCode === 9) {
            var val = this.value,
                start = this.selectionStart,
                end = this.selectionEnd;
            
            this.value = val.substring(0, start) + '\t' + val.substring(end);
            this.selectionStart = this.selectionEnd = start + 1;
            return false;
        }
    });

    content.on("keyup", function() {
        console.log("keyup")
        typingTimeoutId = window.setTimeout(function() {
            isTyping = false;
        }, 300);

        window.clearTimeout(timeoutId);
        timeoutId = window.setTimeout(function() {
            if (isTyping) return;
            console.log("keyup is not typing")
            conn.send("M"+content.val());
        }, 300);
    });

    function writting(name) {
        user_div = $("ul#"+name+".userinfo")
        if (user_div.length) {
            user_div.empty().append("[ "+name+" <span style='font-size:11px'>✍</span> ]")
        } 
        window.setTimeout(function() {
            user_div.empty().append("[ "+name+" <span style='color:lightgreen'>◆</span> ]")
        }, 1000);
    };

    function user(bool, name) {
        user_div = $("ul#"+name+".userinfo")
        if (bool && !user_div.length) {
            $("div #info").append("<ul class='userinfo' id='"+name+"'>[ "+name+" <span style='color:lightgreen'>◆</span> ]</ul>")
        } else if (!bool && user_div.length) {
            user_div.empty().append("<ul class='userinfo' id='"+name+"'>[ <span style='color:#FF0000; font-size:1em;'>"+name+" ◆</span> ]</ul>")
            window.setTimeout(function() {
                user_div.remove()
            }, 500);
        }
    };
};
