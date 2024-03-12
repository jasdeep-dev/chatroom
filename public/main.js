// Function to get current_user from the meta tag
function getCurrentUser() {
    var metaTags = document.getElementsByTagName('meta');

    for (var i = 0; i < metaTags.length; i++) {
        if (metaTags[i].getAttribute('name') === "current_user") {
            return metaTags[i].getAttribute('content');
        }
    }

    return null; // Return null if meta tag with specified name not found
}

function getCookie(cookieName) {
    const name = cookieName + "=";
    const decodedCookie = decodeURIComponent(document.cookie);
    const cookieArray = decodedCookie.split(';');
    for (let i = 0; i < cookieArray.length; i++) {
        let cookie = cookieArray[i];
        while (cookie.charAt(0) === ' ') {
            cookie = cookie.substring(1);
        }
        if (cookie.indexOf(name) === 0) {
            let cookieValue = cookie.substring(name.length, cookie.length);
            // Check if the cookie value is surrounded by double quotes
            if (cookieValue.charAt(0) === '"' && cookieValue.charAt(cookieValue.length - 1) === '"') {
                // Remove the surrounding double quotes
                cookieValue = cookieValue.substring(1, cookieValue.length - 1);
            }
            return cookieValue;
        }
    }
    return "";
}

function setTheme(value){
    console.log(value);
    localStorage.setItem('theme', value);
}

window.onload = function() {
    var container = document.getElementById('textchat');
    if (container){
        container.scrollTop = container.scrollHeight;
    }

    let value = localStorage.getItem('theme');
    var htmlElement = document.querySelector('html');
    htmlElement.setAttribute('data-theme', value);

    let nameCookie = getCookie('session_id');
    if (nameCookie != ''){
        let messageField = document.getElementById(nameCookie)

        if(messageField != null){
            messageField.classList.remove('chat-start')
            messageField.classList.add('chat-end')
        }

        document.getElementById('userform')?.classList.add('hidden');
    }

    document.getElementById('message_input')?.focus();

    document.getElementById('messageform').addEventListener('submit', function(event) {
        event.preventDefault();
      
        var message = document.getElementById('message_input');
        if(message.value == ""){
          return
        }
      
        try {
          ws.send(message.value);
          message.value = "";
        } catch (error) {
          console.error('Error while sending WebSocket message:', error);
        }
      });
};
