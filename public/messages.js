function updateMessages(jsondata) {

    const container = document.getElementById('textchat');

    className = (currentUser == jsondata.Name) ? 'chat-end' : 'chat-start'
    
    const chatContent = `
        <div class="chat ${className}" id="${jsondata.Name}">
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

function updateUsers(jsondata){
    if(jsondata.Name == ""){
        return
    }
    var user = document.getElementById("user_"+jsondata.ID);
    if(user == null){

        // Assuming you have the user's name in a variable
        var name = jsondata.Name; // This is where you dynamically set the name

        // Create the <li> element
        var li = document.createElement('li');
        li.className = 'py-2 p-4 rounded-lg mb-2 bg-base-100 capitalize';
        li.id = 'user_' + name; // Dynamic ID based on user name

        // Create the <span> element for the badge
        var span = document.createElement('span');
        span.className = 'indicator-item badge badge-xs badge-success';
        span.id = 'status_' + name; // Dynamic ID based on user name

        // Append the <span> to the <li>
        li.appendChild(span);

        // Set the text content for the <li>. Note: This text node is added after the <span>
        var textNode = document.createTextNode(name);
        li.appendChild(textNode);

        // Assuming there's an existing <ul> or <ol> element with an ID 'userList' in your document
        var userList = document.getElementById('users_list');
        userList.appendChild(li);

    }


    var status = document.getElementById("status_"+jsondata.Name);
    if(jsondata.IsOnline){

        status.classList.add("badge-success");
    }else{
        status.classList.remove("badge-success");
    }

}