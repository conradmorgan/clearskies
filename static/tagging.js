$(document).ready(function() {
    var mousePos = {
        x: 0,
        y: 0
    }
    var alignmentPoints = [];
    var objects = [];
    var drawCircle = function(radius, id) {
        if (radius < 7) {
            radius = 7;
        }
        var canvas = document.createElement("canvas");
        canvas.setAttribute("id", id);
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
        $('#'+id).css("position", "absolute")
            .css("pointer-events", "none");
        return $('#'+id);
    };
    var retrievedTags = [];
    $.ajax({
        method: "GET",
        url: "/tags/" + $('#uploadId').text(),
        dataType: "json",
        success: function(tags) {
            retrievedTags = tags;
            for (var i = 0; i < tags.length; i++) {
                var w = $('#image').width();
                var radius = w * tags[i].Radius;
                var tag = drawCircle(radius, "tagCircle"+i);
                var x = w * tags[i].X;
                var y = w * tags[i].Y;
                tag.hide()
                   .css("left", x - radius)
                   .css("top", y - radius)
                   .width(radius * 2)
                   .height(radius * 2);
                var nx = Math.cos(tags[i].LabelAngle / 180 * Math.PI);
                var ny = Math.sin(tags[i].LabelAngle / 180 * Math.PI);
                var ox = radius * nx;
                var oy = radius * ny;
                var label = document.createElement("p");
                label.setAttribute("id", "tagLabel"+i);
                $('#image').append(label);
                label = $('#tagLabel'+i);
                $('#tagLabel'+i).hide()
                    .text(tags[i].Tag)
                    .css("position", "absolute")
                    .css("color", "#0F0")
                    .css("word-break", "break-word")
                    .css("text-shadow", "-1px -1px 0 #000, 1px -1px 0 #000, -1px 1px 0 #000, 1px 1px 0 #000")
                    .css("font-weight", "bold")
                    .css("z-index", 100)
                    .css("left", x + (nx-1)*$('#tagLabel'+i).width()/2 + ox*1.01)
                    .css("top", y + (ny-1)*$('#tagLabel'+i).height()/2 + oy*1.01);
                tags[i].Circle = tag;
                tags[i].Label = $('#tagLabel'+i);
            }
        }
    });
    var tags = [];
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
        $.ajax({
            method: "POST",
            url: "/savetags/" + $('#uploadId').text(),
            data: JSON.stringify(tags),
            success: function() {
                location.reload();
            }
        });

    });
    var taggingMode = "none";
    $('#addTag').click(function() {
        taggingMode = "selecting";
        var objectId = prompt("Enter name of object to tag.", "");
        if (!objectId) {
            return;
        }
        var mouseText = document.createElement("p");
        mouseText.setAttribute("id", "mouseText");
        $('#image').append(mouseText);
        $('#mouseText').text(objectId)
            .css("position", "absolute")
            .css("color", "#0F0")
            .css("word-break", "break-word")
            .css("text-shadow", "-1px -1px 0 #000, 1px -1px 0 #000, -1px 1px 0 #000, 1px 1px 0 #000")
            .css("font-weight", "bold")
            .css("cursor", "crosshair")
            .css("z-index", 100)
            .css("pointer-events", "none");
        $('#image').css("cursor", "crosshair");
        var circle = drawCircle(25, "circle");
        circle.css("cursor", "crosshair");
        var c = createCrosshair();
        $('body').mousemove(function(e) {
            var imgX = $('#image').offset().left;
            var imgY = $('#image').offset().top;
            $(c.hline).css("top", e.pageY - imgY);
            $(c.vline).css("left", e.pageX - imgX);
            mousePos.x = e.pageX;
            mousePos.y = e.pageY;
            $('#mouseText').css("top", e.pageY - imgY + 15);
            $('#mouseText').css("left", e.pageX - imgX + 15);
            circle.css("top", e.pageY - imgY - 25);
            circle.css("left", e.pageX - imgX - 25);
        });
        var dragPosX = 0;
        var dragPosY = 0;
        $('#image').click(function() {
            if (taggingMode == "selecting") {
                taggingMode = "resizing";
                $(c.hline).remove();
                $(c.vline).remove();
                $('body').unbind("mousemove");
                $('body').mousemove(function(e) {
                    mousePos.x = e.pageX;
                    mousePos.y = e.pageY;
                    var x = tags[tags.length-1].x*$('#image').width();
                    var y = tags[tags.length-1].y*$('#image').width();
                    var ox = e.pageX - ($('#image').offset().left + x);
                    var oy = e.pageY - ($('#image').offset().top + y);
                    var radius = Math.sqrt(ox*ox + oy*oy);
                    circle.remove();
                    circle = drawCircle(radius, "circle");
                    circle.css("left", x - radius);
                    circle.css("top", y - radius);
                    var nx = ox / radius;
                    var ny = oy / radius;
                    $('#mouseText').css("left", x + (nx-1)*$('#mouseText').width()/2 + ox*1.01)
                                   .css("top", y + (ny-1)*$('#mouseText').height()/2 + oy*1.01);
                    tags[tags.length-1].labelAngle = Math.atan2(oy, ox) / Math.PI * 180;
                    tags[tags.length-1].radius = radius / $('#image').width();
                });
                tags.push({
                    tag: objectId,
                    x: (mousePos.x - $('#image').offset().left) / $('#image').width(),
                    y: (mousePos.y - $('#image').offset().top) / $('#image').width()
                });
            } else if (taggingMode == "resizing") {
                dragPosX = mousePos.x - $('#image').offset().left - tags[tags.length-1].x*$('#image').width();
                dragPosY = mousePos.y - $('#image').offset().top - tags[tags.length-1].y*$('#image').width();
                taggingMode = "moving";
                $('body').unbind("mousemove");
                $('body').mousemove(function(e) {
                    mousePos.x = e.pageX;
                    mousePos.y = e.pageY;
                    var imgX = $('#image').offset().left;
                    var imgY = $('#image').offset().top;
                    var x = e.pageX - imgX - dragPosX;
                    var y = e.pageY - imgY - dragPosY;
                    var radius = $('#image').width()*tags[tags.length-1].radius;
                    circle.remove();
                    circle = drawCircle(radius, "circle");
                    circle.css("left", x - radius);
                    circle.css("top", y - radius);
                    var ox = radius*Math.cos(tags[tags.length-1].labelAngle / 180 * Math.PI);
                    var oy = radius*Math.sin(tags[tags.length-1].labelAngle / 180 * Math.PI);
                    var nx = ox / radius;
                    var ny = oy / radius;
                    $('#mouseText').css("left", x + (nx-1)*$('#mouseText').width()/2 + ox*1.01)
                                   .css("top", y + (ny-1)*$('#mouseText').height()/2 + oy*1.01);
                    tags[tags.length-1].labelAngle = Math.atan2(oy, ox) / Math.PI * 180;
                    tags[tags.length-1].radius = radius / $('#image').width();
                    tags[tags.length-1].x = x / $('#image').width();
                    tags[tags.length-1].y = y / $('#image').width();
                });
            } else if (taggingMode == "moving") {
                $('#saveTags').show();
                $('#mouseText').attr("id", "manualTagName" + (tags.length-1));
                circle.attr("id", "manualTagHighlight" + (tags.length-1));
                $('#image').unbind("click");
                $('body').unbind("mousemove");
                if (alignmentPoints.length >= 2) {
                    console.log(alignmentPoints);
                }
            }
        });
    });
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
                        var tag = drawCircle(dim, "tag"+i);
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
        var circle = drawCircle(25, "circle");
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
    $(window).resize(function() {
        var w = $('#image').width();
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
        for (var i = 0; i < tags.length; i++) {
            var tag = $('#manualTagHighlight'+i);
            var x = w * tags[i].x;
            var y = w * tags[i].y;
            var radius = w * tags[i].radius;
            tag.css("left", x - tag.width()/2)
               .css("top", y - tag.height()/2)
               .width(radius * 2)
               .height(radius * 2);
            var nx = Math.cos(tags[i].labelAngle / 180 * Math.PI);
            var ny = Math.sin(tags[i].labelAngle / 180 * Math.PI);
            var ox = radius * nx;
            var oy = radius * ny;
            var label = $('#manualTagName'+i);
            label.css("left", x + (nx-1)*label.width()/2 + ox*1.01)
                 .css("top", y + (ny-1)*label.height()/2 + oy*1.01);
        }
        for (var i = 0; i < retrievedTags.length; i++) {
            var x = w * retrievedTags[i].X;
            var y = w * retrievedTags[i].Y;
            var radius = w * retrievedTags[i].Radius;
            retrievedTags[i].Circle.css("left", x - radius)
                                   .css("top", y - radius)
                                   .width(radius * 2)
                                   .height(radius * 2);
            var nx = Math.cos(retrievedTags[i].LabelAngle / 180 * Math.PI);
            var ny = Math.sin(retrievedTags[i].LabelAngle / 180 * Math.PI);
            var ox = radius * nx;
            var oy = radius * ny;
            var label = retrievedTags[i].Label;
            label.css("left", x + (nx-1)*label.width()/2 + ox*1.01)
                 .css("top", y + (ny-1)*label.height()/2 + oy*1.01);
        }
    });
    $('#image').mouseenter(function() {
        for (var i = 0; i < retrievedTags.length; i++) {
            retrievedTags[i].Circle.show();
            retrievedTags[i].Label.show();
        }
    });
    $('#image').mouseleave(function() {
        for (var i = 0; i < retrievedTags.length; i++) {
            retrievedTags[i].Circle.hide();
            retrievedTags[i].Label.hide();
        }
    });
});
