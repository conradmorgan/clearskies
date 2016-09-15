$(document).ready(function() {
    var mousePos = {
        x: 0,
        y: 0
    }
    var drawCircle = function(radius) {
        if (radius < 7) {
            radius = 7;
        }
        var canvas = document.createElement("canvas");
        canvas.width = radius * 2;
        canvas.height = radius * 2;
        var context = canvas.getContext("2d");
        context.beginPath();
        context.fillStyle = "#0F0";
        context.arc(radius, radius, radius-5, 0, 2 * Math.PI, false);
        context.arc(radius, radius, radius-7, 0, 2 * Math.PI, true);
        context.lineWidth = 3;
        context.strokeStyle = "#000";
        context.stroke();
        context.fill();
        $('#image').append(canvas);
        $(canvas).css("position", "absolute")
                 .css("pointer-events", "none");
        return canvas;
    };
    var Tag = function(radius, label) {
        this.radius = radius;
        this.circle = drawCircle(radius);
        this.label = document.createElement("p");
        $(this.label).text(label);
        $('#image').append(this.label);
        $(this.label).css("position", "absolute")
                     .css("color", "#0F0")
                     .css("word-break", "break-word")
                     .css("text-shadow", "-1px -1px 0 #000, 1px -1px 0 #000, -1px 1px 0 #000, 1px 1px 0 #000")
                     .css("font-weight", "bold")
                     .css("z-index", 100);
        this.labelAngle = 0;
        this.x = 0;
        this.y = 0;
        $(window).resize(this._updateRadius);
        this.realX = function() {
            return this.x * $('#image').width();
        }
        this.realY = function() {
            return this.y * $('#image').width();
        }
        this.getNormalizedRadius = function() {
            return this.radius / $('#image').width();
        };
        this.setPosition = function(x, y) {
            this.x = x;
            this.y = y;
            this._updatePosition();
        }
        this.setRealPosition = function(x, y) {
            this.x = x / $('#image').width();
            this.y = y / $('#image').width();
            this._updatePosition();
        };
        this.setLabel = function(text) {
            $(this.label).text(text);
            this._updatePosition();
        };
        this.setRadius = function(radius) {
            this.radius = radius;
            this._updateRadius();
        };
        this.setLabelAngle = function(angle) {
            this.labelAngle = angle;
            this._updatePosition();
        };
        this._updatePosition = function() {
            var nx = Math.cos(this.labelAngle / 180 * Math.PI);
            var ny = Math.sin(this.labelAngle / 180 * Math.PI);
            var ox = this.radius * nx;
            var oy = this.radius * ny;
            $(this.circle).css("left", this.realX() - $(this.circle).height()/2)
                          .css("top", this.realY() - $(this.circle).width()/2);
            $(this.label).css("left", this.realX() + (nx-1)*$(this.label).width()/2 + ox*1.01)
                         .css("top", this.realY() + (ny-1)*$(this.label).height()/2 + oy*1.01);
        };
        this._updateRadius = function() {
            var id = $(this.circle).attr("id");
            var crosshair = ($(this.circle).css("cursor") == "crosshair");
            $(this.circle).remove();
            this.circle = drawCircle(this.radius, id);
            if (crosshair) {
                $(this.circle).css("cursor", "crosshair");
            }
            this._updatePosition();
        }
        return this;
    };
    var rawTags = JSON.parse(document.getElementById("tags").innerHTML);
    var tags = [];
    for (var i = 0; i < rawTags.length; i++) {
        var tag = new Tag(rawTags[i].Radius*$('#image').width(), rawTags[i].Tag);
        $(tag.circle).hide();
        $(tag.label).hide();
        tag.setPosition(rawTags[i].X, rawTags[i].Y);
        tag.setLabelAngle(rawTags[i].LabelAngle);
        tags.push(tag);
    }
    var createCrosshair = function() {
        var obj = {
            hline: document.createElement("div"),
            vline: document.createElement("div")
        };
        $('#image').append(obj.hline);
        $('#image').append(obj.vline);
        $(obj.hline).width("100%")
                    .height("1px")
                    .css("background", "white")
                    .css("position", "absolute")
        $(obj.vline).width("1px")
                    .height("100%")
                    .css("background", "white")
                    .css("position", "absolute")
                    .css("top", 0);
        return obj;
    };
    $('#clearTags').click(function() {
        if (confirm("Confirm removal of tags.")) {
            $.ajax({
                method: "POST",
                url: "/cleartags/" + $('#uploadId').text(),
                success: function() {
                    location.reload();
                }
            });
        }
    });
    $('#saveTags').click(function() {
        var tagsToSend = [];
        for (var i = 0; i < userTags.length; i++) {
            tagsToSend.push({
                Tag: $(userTags[i].label).text(),
                Radius: userTags[i].radius / $('#image').width(),
                LabelAngle: userTags[i].labelAngle,
                X: userTags[i].x,
                Y: userTags[i].y
            });
        }
        $.ajax({
            method: "POST",
            url: "/savetags/" + $('#uploadId').text(),
            data: JSON.stringify(tagsToSend),
            success: function() {
                location.reload();
            }
        });

    });
    var userTags = [];
    $('#addTag').click(function() {
        var objectId = prompt("Enter name of object to tag.", "");
        if (!objectId) {
            return;
        }
        var tag = new Tag(25, objectId);
        $(tag.circle).css("cursor", "crosshair");
        $(tag.label).css("cursor", "crosshair");
        var c = createCrosshair();
        $('body').mousemove(function(e) {
            var imgX = $('#image').offset().left;
            var imgY = $('#image').offset().top;
            mousePos.x = e.pageX - imgX;
            mousePos.y = e.pageY - imgY;
            $(c.hline).css("top", mousePos.y);
            $(c.vline).css("left", mousePos.x);
            tag.setRealPosition(mousePos.x, mousePos.y);
        });
        var dragPosX = 0;
        var dragPosY = 0;
        taggingMode = "positioning";
        $('#image').click(function() {
            switch (taggingMode) {
            case "positioning":
                taggingMode = "resizing";
                $(c.hline).remove();
                $(c.vline).remove();
                $('body').unbind("mousemove");
                $('body').mousemove(function(e) {
                    mousePos.x = e.pageX - $('#image').offset().left;
                    mousePos.y = e.pageY - $('#image').offset().top;
                    var ox = mousePos.x - tag.realX();
                    var oy = mousePos.y - tag.realY();
                    tag.setRadius(Math.sqrt(ox*ox + oy*oy));
                    tag.setLabelAngle(Math.atan2(oy, ox) / Math.PI * 180);
                });
                break;
            case "resizing":
                taggingMode = "placing";
                dragPosX = mousePos.x - tag.realX();
                dragPosY = mousePos.y - tag.realY();
                $('body').unbind("mousemove");
                $('body').mousemove(function(e) {
                    mousePos.x = e.pageX - $('#image').offset().left;
                    mousePos.y = e.pageY - $('#image').offset().top;
                    tag.setRealPosition(mousePos.x - dragPosX, mousePos.y - dragPosY);
                });
                break;
            case "placing":
                $('body').unbind("mousemove");
                $('#image').unbind("click");
                $('#saveTags').show();
                userTags.push(tag);
                break;
            }
        });
    });
    $('#image').mouseenter(function() {
        for (var i = 0; i < tags.length; i++) {
            $(tags[i].circle).show();
            $(tags[i].label).show();
        }
    });
    $('#image').mouseleave(function() {
        for (var i = 0; i < tags.length; i++) {
            $(tags[i].circle).hide();
            $(tags[i].label).hide();
        }
    });
    /*
    $(window).resize(function() {
        for (var i = 0; i < alignmentPoints.length; i++) {
            var text = $('#text'+(i+1));
            var circle = $('#circle'+(i+1));
            text.css("left", w * alignmentPoints[i].point.x + 15);
            text.css("top", w * alignmentPoints[i].point.y + 15);
            circle.css("left", w * alignmentPoints[i].point.x - 25);
            circle.css("top", w * alignmentPoints[i].point.y - 25);
        }
        for (var i = 0; i < objects.length; i++) {
            $('#tag'+i).css("left", w * objects[i].Point.X - 25);
            $('#tag'+i).css("top", w * objects[i].Point.Y - 25);
            $('#name'+i).css("left", w * objects[i].Point.X + 15);
            $('#name'+i).css("top", w * objects[i].Point.Y + 15);
        }
    });
    */
    /*
    var alignmentPoints = [];
    var objects = [];
    $('#autoTagger').click(function() {
        if (alignmentPoints.length >= 2) {
            json = JSON.stringify({
                aspectRatio: $('#image').width() / $('#image').height(),
                points: alignmentPoints,
                isMirrored: $('#isMirrored').prop("checked")
            });
            $.ajax({
                method: "POST",
                url: "/generatetags",
                data: json,
                dataType: "json",
                success: function(data) {
                    objects = data;
                    console.log(data);
                    for (var i = 1; i <= 2; i++) {
                        $('#circle'+i).remove();
                        $('#text'+i).remove();
                    }
                    for (var i = 0; i < data.length; i++) {
                        if (data[i].Name == "V* S And") {
                            continue;
                        }
                        var dim = 25;
                        if (data[i].Dim) {
                            dim = data[i].Dim*$('#image').width() / 2;
                        }
                        var tag = $(drawCircle(dim, "tag"+i));
                        tag.css("top", $('#image').width() * data[i].Point.Y - tag.height()/2);
                        tag.css("left", $('#image').width() * data[i].Point.X - tag.width()/2);
                        var text = document.createElement("p");
                        text.setAttribute("id", "name"+i);
                        $('#image').append(text);
                        text = $('#name'+i);
                        text.text(data[i].Name);
                        text.css("position", "absolute");
                        text.css("color", "#0F0");
                        text.css("word-break", "break-word");
                        text.css("text-shadow", "-1px -1px 0 #000, 1px -1px 0 #000, -1px 1px 0 #000, 1px 1px 0 #000");
                        text.css("font-weight", "bold");
                        text.css("top", $('#image').width() * data[i].Point.Y + 15);
                        text.css("left", $('#image').width() * data[i].Point.X + 15);
                        var mag = Number($('#magnitude').val());
                        if (data[i].Magnitude <= mag) {
                            tag.show();
                            text.show();
                        } else {
                            tag.hide();
                            text.hide();
                        }
                    }
                }
            });
            return;
        }
        var objectId = prompt("Enter name of alignment object " + (alignmentPoints.length+1) + " of 2.", "");
        if (!objectId) {
            return;
        }
        switch (alignmentPoints.length) {
            case 0:
                $('#autoTagger').text("Add Second Alignment Point");
                break;
            case 1:
                $('#autoTagger').text("Generate Tags");
                break;
        }
        var mouseText = document.createElement("p");
        mouseText.setAttribute("id", "mouseText");
        $('#image').append(mouseText);
        $('#mouseText').text(objectId);
        $('#mouseText').css("position", "absolute");
        $('#mouseText').css("color", "#0F0");
        $('#mouseText').css("word-break", "break-word");
        $('#mouseText').css("text-shadow", "-1px -1px 0 #000, 1px -1px 0 #000, -1px 1px 0 #000, 1px 1px 0 #000");
        $('#mouseText').css("font-weight", "bold");
        $('#mouseText').css("cursor", "crosshair");
        $('#image').css("cursor", "crosshair");
        var circle = $(drawCircle(25, "circle"));
        circle.css("cursor", "crosshair");
        $('body').mousemove(function(e) {
            mousePos.x = e.pageX;
            mousePos.y = e.pageY;
            $('#mouseText').css("top", e.pageY - $('#image').offset().top + 15);
            $('#mouseText').css("left", e.pageX - $('#image').offset().left + 15);
            circle.css("top", e.pageY - $('#image').offset().top - 25);
            circle.css("left", e.pageX - $('#image').offset().left - 25);
        });
        $('#image').click(function() {
            alignmentPoints.push({
                objectId: objectId,
                point: {
                    x: (mousePos.x - $('#image').offset().left) / $('#image').width(),
                    y: (mousePos.y - $('#image').offset().top) / $('#image').width()
                }
            });
            $('#mouseText').attr("id", "text" + alignmentPoints.length);
            circle.attr("id", "circle" + alignmentPoints.length);
            $('#image').unbind("click");
            $('body').unbind("mousemove");
            if (alignmentPoints.length >= 2) {
                console.log(alignmentPoints);
            }
        });
    });
    $('#magnitude').change(function() {
        var mag = Number($('#magnitude').val());
        $('#magText').text(mag);
        for (var i = 0; i < objects.length; i++) {
            if (objects[i].Magnitude <= mag) {
                $('#tag'+i).show();
                $('#name'+i).show();
            } else {
                $('#tag'+i).hide();
                $('#name'+i).hide();
            }
        }
    });
    */
});
