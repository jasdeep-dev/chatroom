function updateMessages(jsondata) {

    const container = document.getElementById('textchat');

    var currentUser = getCurrentUser();
    var chatContent = "";

    if(currentUser == jsondata.Name){
        chatContent = `
        <div id="${jsondata.Name}" class="bg-base-200 rounded p-2 my-2">
            <div>
                Me: ${jsondata.Text}
            </div>
        </div>`
    }else{
        chatContent = `
        <div id="${jsondata.Name}" class="bg-base-100 rounded p-2 my-2">
            <div>
                ${jsondata.Name}: ${jsondata.Text}
            </div>
        </div>`
    }

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
        status?.classList.add("badge-success");
    }else{
        status?.classList.remove("badge-success");
    }
}