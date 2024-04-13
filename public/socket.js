var ws = new WebSocket("ws://localhost:3000/ws");

document.addEventListener("DOMContentLoaded", function() {
    ws.onopen = function() {
        console.log("Connected to the WebSocket server.");

        var user_name = document.getElementById("user_name")
        user_name.style.color = "green";
    };
    ws.onmessage = function(evt) {
        var jsondata = JSON.parse(evt.data);

        if(jsondata.User){
            console.log("Message from server USER: ", evt.data);
            updateUsers(jsondata.User);
        }else if(jsondata.Message){
            console.log("Message from server MESSAGE: ", evt.data);
            updateMessages(jsondata.Message);
        }else if(jsondata.Status){
            console.log("==>", jsondata.Status)
        }
    };
    ws.onclose = function() {
        var user_name = document.getElementById("user_name")
        user_name.style.color = "white";

        console.log("Disconnected from the WebSocket server.");
    };
    ws.onerror = function(err) {
        console.log("WebSocket error: ", err);
    };
});
