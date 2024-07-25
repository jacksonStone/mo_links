// const root = "localhost:3003"
const root = "https://www.molinks.me"
document.addEventListener('DOMContentLoaded', function () {
    // Get the current tab's URL
    chrome.tabs.query({ active: true, currentWindow: true }, function (tabs) {
        var currentTab = tabs[0];
        var currentUrl = currentTab.url;
        if (!currentUrl) {
            // Some pages like about:blank don't have a URL
            return;
        }

        // Set the URL input field's value
        document.getElementById('url').value = currentUrl;
    });
    const nameInput = document.getElementById('name');
    const nameGroup = document.getElementById('name-group');
    const urlInput = document.getElementById('url');
    nameInput.addEventListener('input', hideError);
    urlInput.addEventListener('input', hideError);
    function surfaceError(message) {
        document.getElementById("error-text").textContent = message;
        nameGroup.classList.add('error');
    }
    function hideError() {
        if (validateName(nameInput.value) && validateUrl(urlInput.value)) {
            document.getElementById("error-text").textContent = "";
            nameGroup.classList.remove('error');
        }
    }
    function validateUrl(url) {
        if (url === "") {
            surfaceError("url must not be empty");
            return false
        }
        if (!url.startsWith("http://") && !url.startsWith("https://")) {
            surfaceError("url must start with http:// or https:// or mo/");
            return false
        }
        // can't be longer than 2000 characters
        if (url.length > 2000) {
            surfaceError("url must be 2000 characters or less");
            return false
        }
        return true;
    }
    function validateName(name) {
        const validRegex = /^[a-zA-Z0-9_-]+$/;
        if (name.length > 255) {
            surfaceError("Name must be 255 characters or less.");
            return false;
        }
        if (name === '____reserved') {
            surfaceError("____reserved is... well... reserved.");
            return false;
        }
        if (name === '' || validRegex.test(name)) {
            return true;
        } else {
            surfaceError("Only letters, numbers, _, and - are allowed in the mo/ path.");
            return false;
        }
    }

    document.getElementById('submit').addEventListener('click', function () {
        let name = document.getElementById('name').value;
        let url = document.getElementById('url').value;
        // Sanitize name and url
        name = name.trim();
        url = url.trim();
        if (validateName(name) && validateUrl(url) && name && url) {
            hideError()
            fetch(root + '/____reserved/api/add', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name, url }),
            })
                .then(() => {
                    alert('mo/' + name + " created successfully!");
                    // Clear the form
                    document.getElementById('name').value = '';
                    document.getElementById('url').value = '';
                })
                .catch((error) => {
                    console.error('Error:', error);
                    alert('An error occurred. Please try again: ' + error.message);
                });
        } else {
            alert('Please fill in both fields.');
        }
    });
    // Add event listener for the "See Your Mo Links" link
    document.getElementById('see-links').addEventListener('click', function (e) {
        e.preventDefault(); // Prevent the default link behavior
        chrome.tabs.create({ url: root });
    });
});