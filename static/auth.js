function hasher(password, hexSalt) {
    var salt = sjcl.codec.hex.toBits(hexSalt);
    var hash = sjcl.misc.pbkdf2(password, salt, 1024, 128);
    return sjcl.codec.hex.fromBits(hash);
}

$(document).ready( function() {
    $('#signupButton').click(function ( event ) {
        event.preventDefault();
        var form = {
            username: $('#username').val(),
            email:    $('#email').val(),
            fullName: $('#fullName').val(),
            passcode: ""
        };
        var password = $('#password').val();
        if (password.length < 8) {
            alert("Password must be at least 8 characters long!");
            return
        }
        var confirmPassword = $('#confirmPassword').val();
        if (password != confirmPassword){
            alert("Passwords do not match!");
            return;
        }
        var saltFields = {
            usernameOrEmail: form.username
        }; 
        $.ajax({
            method: "POST",
            url: "/salt",
            data: saltFields,
            success: function(data) {
                // Hash passwords client-side for privacy.
                // Passcodes are subsequently hashed server-side.
                form.passcode = hasher(password, data);
                $.ajax({
                    method: "POST",
                    url: "/signup",
                    data: form,
                    success: function() {
                        console.log("*** ^SUCCESS^ ***");
                        window.location = "/login";
                    },
                    error: function(data) {
                        if (data.responseText == "email exists") {
                            alert("Email already exists!");
                        } else if (data.responseText == "username exists") {
                            alert("Username already exists!");
                        } else {
                            alert("Signup failure");
                        }
                        console.log("*** /FAILURE/ ***");
                    }
                });
            }
        });
    });
    $('#loginButton').click(function ( event ) {
        event.preventDefault();
        var form = {
            usernameOrEmail: $('#usernameOrEmail').val(),
        };
        var password = $('#password').val();
        $.ajax({
            method: "POST",
            url: "/salt",
            data: form,
            success: function(data) {
                form.passcode = hasher(password, data);
                $.ajax({
                    method: "POST",
                    url: "/login",
                    data: form,
                    success: function() {
                        console.log("*** ^SUCCESS^ ***");
                        window.location = "/account";
                    },
                    error: function(data) {
                        if (data.responseText == "multiple unverified emails") {
                            alert("Please use your username (not your email) to log in.");
                        } else {
                            alert("Incorrect username or password.");
                        }
                        console.log("*** /FAILURE/ ***");
                    }
                });
            },
            error: function(data) {
                if (data.responseText == "multiple unverified emails") {
                    alert("There are multiple unverified accounts associated with this email. Please use your username instead to log in. You can use your email address to log in once it has been verified.");
                } else {
                    alert("Handshake error.");
                }
                console.log("*** /FAILURE/ ***");
            }
        });
    });
    $('#changePasswordButton').click(function ( event ) {
        event.preventDefault();
        var form = {
            username: $('#username').val(),
        };
        if ($('#resetToken').length) {
            form.resetToken = $('#resetToken').val();
        }
        var newPassword = $('#newPassword').val();
        if( newPassword.length < 8 ){
            alert("Password must be at least 8 characters long!");
            return
        }
        var confirmNewPassword = $('#confirmNewPassword').val();
        if( newPassword != confirmNewPassword){
            alert("Passwords do not match!");
            return;
        }
        var saltFields = {
            usernameOrEmail: form.username
        };
        $.ajax({
            method: "POST",
            url: "/salt",
            data: saltFields,
            success: function(data) {
                if ($('#oldPassword').length) {
                    form.oldPasscode = hasher($('#oldPassword').val(), data);
                }
                form.newPasscode = hasher(newPassword, data);
                $.ajax({
                    method: "POST",
                    url: "/changepassword",
                    data: form,
                    success: function() {
                        console.log("*** ^SUCCESS^ ***");
                        if ($('#resetToken').length) {
                            window.location = "/login";
                        } else {
                            window.location = "/account";
                        }
                    },
                    error: function(data) {
                        console.log("*** /FAILURE/ ***");
                        if (data.responseText == "incorrect password") {
                            alert("Wrong password!");
                        } else {
                            alert("Something went wrong!");
                        }
                    }
                });
            },
            error: function() {
                alert("Something went wrong!");
            },
        });
    });
    $('#username').keyup(function() {
        var usernameMatcher = /^[0-9A-Za-z_-]{1,30}$/;
        if (!usernameMatcher.exec($('#username').val())) {
            $('#username').css("background-color", "#ffcccc");
        } else {
            $('#username').css("background-color", "#fff");
        }
    });
    $('#confirmPassword').keyup(function() {
        if ($('#confirmPassword').val() != $('#password').val()) {
            $('#confirmPassword').css("background-color", "#ffcccc");
        } else {
            $('#confirmPassword').css("background-color", "#fff");
        }
    });
    $('#confirmNewPassword').keyup(function() {
        if ($('#confirmNewPassword').val() != $('#newPassword').val()) {
            $('#confirmNewPassword').css("background-color", "#ffcccc");
        } else {
            $('#confirmNewPassword').css("background-color", "#fff");
        }
    });
});
