function updateChatInnerHTML(jsondata) {

    jsondata = JSON.parse(jsondata)
    const container = document.getElementById('textchat');

    const chatContent = `
        <div class="chat chat-start" id="${jsondata.Name}">
            <div class="chat-header">
                <span class="capitalize">${jsondata.Name}</span>
                <!-- Implement formatTime function for TimeStamp and uncomment if needed -->
                <!-- <time class="text-xs opacity-50">${jsondata.TimeStamp}</time> -->
            </div>
            <div class="chat-bubble">${jsondata.Text}</div>
            <!-- <div class="chat-footer opacity-50">Seen</div> -->
        </div>
    `;
    container.innerHTML += chatContent;

    container.scrollTop = container.scrollHeight;
}