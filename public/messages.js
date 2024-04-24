function UserAdded(event){
    event.parentNode.setAttribute("hidden", true)
}
function RemoveUser(event){
    event.parentNode.parentNode.parentNode.setAttribute('hidden', true)
}
function CreateGroupFunction() {
    my_modal_5.showModal()
    document.getElementById("groupForm")?.addEventListener('htmx:beforeSwap', function(evt) {
        if (evt.detail.isError) {
            classList = document.getElementById("createGroupError").classList
            classList.remove("hidden");
            classList.add("text-error");
            document.querySelector("#createGroupError .errorMsg").innerText = evt.detail.xhr.response;
            setTimeout(function() {
                classList.add("hidden");
            }, 5000);
        }else{
            my_modal_5.close()
            window.location.href = evt.detail.xhr.response;
        }
    });
}

function ScrollToTop(){
    var cont = document.querySelectorAll('#textchat .messageList')[0]
    if(cont){
        cont.scrollTop = cont.scrollHeight;
    }
    document.getElementById("GroupSelection")?.addEventListener('htmx:beforeSwap', function(evt) {
        var cont = document.querySelectorAll('#textchat .messageList')[0]
        cont.scrollTop = cont.scrollHeight;
    });
}

function updateUsers(jsondata){
    // if(jsondata.id == ""){
    //     return
    // }
    // var user = document.getElementById("user_"+jsondata.id);
    // if(user == null){

    //     var name = jsondata.first_name;

    //     var li = document.createElement('li');
    //     li.className = 'bg-base-100 rounded p-2 my-2';
    //     li.id = 'user_' + name;

    //     var span = document.createElement('span');
    //     span.className = 'indicator-item badge badge-xs badge-success';
    //     span.id = 'status_' + name;
    
    //     li.appendChild(span);
    
    //     var textNode = document.createTextNode(name);
    //     li.appendChild(textNode);
    
    //     var userList = document.getElementById('users_list');
    //     userList.appendChild(li);
    }
function updateMessages(jsondata) {
    const timestampStr = jsondata.TimeStamp;
    const timestamp = new Date(timestampStr);

    const hour = timestamp.getHours().toString().padStart(2, "0");
    const minute = timestamp.getMinutes().toString().padStart(2, "0");
    const timeFormatted = `${hour}:${minute}`;

    var currentUser = getCurrentUser();
    var cont = document.getElementById("item-list-"+jsondata.GroupID);

    if(cont == null){

        var list = document.querySelector("#li"+jsondata.GroupID+ " span")
        list.classList.remove("hidden")
        var currentCount = parseInt(list.innerText);
        var newCount = currentCount + 1;
        list.innerText = newCount;
    }else{
        var chatContent = `<div id='msg_7' class='user_message_${currentUser == jsondata.Email}'>
                <div class='chat-header'>
                    ${jsondata.Name}
                    <time class='text-xs opacity-50'>${timeFormatted}</time>
                </div>
                <div class='chat-image avatar placeholder'>
                    <div class='bg-neutral text-neutral-content rounded-full w-10'>
                        <span class='text-center'>${jsondata.Name[0]}</span>
                    </div>
                </div>
                <div class='chat-bubble'>${jsondata.Text}</div>
            </div>`
        cont.innerHTML += chatContent;
        cont.scrollTop = cont.scrollHeight;
    }
}