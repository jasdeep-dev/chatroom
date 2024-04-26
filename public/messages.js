function UserAdded(event){
    event.parentNode.setAttribute("hidden", true)
}

function RemoveUser(event){
    event.parentNode.parentNode.parentNode.setAttribute('hidden', true)
}

function CreateGroupFunction() {
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

function CreateDirectChat(event) {
    event.target?.addEventListener('htmx:afterRequest', function(evt) {
        window.location.href = evt.detail.xhr.responseURL
    });
}

function groupChanged(event){
    prev = event.target.parentNode.parentNode.getElementsByClassName("bg-base-100")[0]
    prev.classList.remove("bg-base-100")
    prev.classList.add("bg-base-300")
    event.target.parentNode.classList.remove('bg-base-300')
    event.target.parentNode.classList.add('bg-base-100')
    document.getElementById('textchat').innerHTML = `<div class="skeleton h-full"></div>`
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

function updateUsers(jsondata){}

function updateMessages(jsondata) {
    const timestampStr = jsondata.TimeStamp;
    const timestamp = new Date(timestampStr);

    const hour = timestamp.getHours().toString().padStart(2, "0");
    const minute = timestamp.getMinutes().toString().padStart(2, "0");
    const timeFormatted = `${hour}:${minute}`;

    var currentUser = getCurrentUser();
    var cont = document.getElementById("item-list-"+jsondata.GroupID);

    if(cont == null){

        var list = document.getElementById("li"+jsondata.GroupID).parentElement.children[1]
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