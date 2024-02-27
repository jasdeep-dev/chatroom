var ws = new WebSocket("ws://localhost:3000/ws");

document.addEventListener("DOMContentLoaded", function() {
    // var ws = new WebSocket("ws://localhost:3000/ws");
    ws.onopen = function() {
        console.log("Connected to the WebSocket server.");
        ws.send("Hello from the client!");
    };
    ws.onmessage = function(evt) {
        console.log("Message from server: ", evt.data);
    };
    ws.onclose = function() {
        console.log("Disconnected from the WebSocket server.");
    };
    ws.onerror = function(err) {
        console.log("WebSocket error: ", err);
    };
});
