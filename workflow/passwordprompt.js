#!/usr/bin/osascript
ObjC.import('stdlib')

function run(argument){
    var username = $.getenv('username')
    if (argument == "Code") {
        dialogtext = "OTP token"
        hidden = "false"
    } else {
        dialogtext = `Enter ${argument} for ${username}:`
        hidden = "true"
    }

    var app = Application.currentApplication()
    app.includeStandardAdditions = true

    var response = app.displayDialog(dialogtext, {
        defaultAnswer: "",
        withIcon: Path("./icon.png"),
        buttons: ["Cancel", "OK"],
        defaultButton: "OK",
        cancelButton: "Cancel",
        givingUpAfter: 120,
        hiddenAnswer: hidden
    })

    password = response.textReturned
    return password
}
