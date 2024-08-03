document.addEventListener("DOMContentLoaded", function () {
    let currentUrl = window.location.origin + window.location.pathname
    if (!currentUrl.endsWith("/")) {
        currentUrl += "/"
    }
    fetch(currentUrl + "____reserved/_ping")
})