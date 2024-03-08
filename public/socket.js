var ws = new WebSocket("ws://localhost:3000/ws");

document.addEventListener("DOMContentLoaded", function() {
    ws.onopen = function() {
        console.log("Connected to the WebSocket server.");
        ws.send("Hello from the client!");
    };
    ws.onmessage = function(evt) {
        console.log("Message from server: ", evt.data);
        var jsondata = JSON.parse(evt.data);
        if(jsondata.User){
            updateUsers(jsondata.User);
        }else if(jsondata.Message){
            updateMessages(jsondata.Message);
        }
    };
    ws.onclose = function() {
        classList = document.getElementById("active_button").classList
        classList.remove("badge-success")
        removeCookie("session_id")
        console.log("session")
        console.log("Disconnected from the WebSocket server.");
    };
    ws.onerror = function(err) {
        console.log("WebSocket error: ", err);
    };
});
