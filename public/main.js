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


function removeCookie(cookieName) {
    document.cookie = cookieName + "=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;";
}