$(document).ready(function() {
    var mousePos = {
        x: 0,
        y: 0
    }
    var alignmentPoints = [];
    var objects = [];
    var drawCircle = function(radius, id) {
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
        $('#'+id).css("position", "absolute");
        return $('#'+id);
    };
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
        circle.click(function() {
            alignmentPoints.push({
                objectId: objectId,
                point: {
                    x: (mousePos.x - $('#image').offset().left) / $('#image').width(),
                    y: (mousePos.y - $('#image').offset().top) / $('#image').width()
                }
            });
            $('#mouseText').attr("id", "text" + alignmentPoints.length);
            circle.attr("id", "circle" + alignmentPoints.length);
            circle.unbind("click");
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
    });
});
