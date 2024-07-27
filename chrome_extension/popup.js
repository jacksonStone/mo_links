const root = "https://www.molinks.me"

function validateUrl(url) {
    // can't be longer than 2048 characters
    if (url.length > 2048) {
        surfaceError("url must be 2048 characters or less");
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
    if (name === '_ping') {
        surfaceError("_ping is reserved.");
        return false;
    }
    if (name === '' || validRegex.test(name)) {
        return true;
    } else {
        surfaceError("Only letters, numbers, _, and - are allowed in the mo/ path.");
        return false;
    }
}
function surfaceError(message) {
    document.getElementById("error-text").textContent = message;
    nameGroup.classList.add('error');
}

document.addEventListener('DOMContentLoaded', function () {
    const nameInput = document.getElementById('name');
    const nameGroup = document.getElementById('name-group');
    const urlInput = document.getElementById('url');
    const hideError = () => {
        if (validateName(nameInput.value) && validateUrl(urlInput.value)) {
            document.getElementById("error-text").textContent = "";
            nameGroup.classList.remove('error');
        }
    }

    nameInput.addEventListener('input', hideError);
    urlInput.addEventListener('input', hideError);

    // Get the current tab's URL
    chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
        var currentTab = tabs[0];
        var currentUrl = currentTab.url;
        if (!currentUrl) {
            // Some pages like about:blank don't have a URL
            return;
        }
        // Set the URL input field's value
        urlInput.value = currentUrl;
    });



    document.getElementById('submit').addEventListener('click', function () {
        let name = nameInput.value;
        let url = urlInput.value;
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
                .then(response => {
                    if (!response.ok) {
                        // Body will be a text string
                        return response.text().then(text => {
                            throw new Error(text);
                        });
                    }
                })
                .then(() => {
                    alert('mo/' + name + " created successfully!");
                    // Clear the form
                    document.getElementById('name').value = '';
                    document.getElementById('url').value = '';
                })
                .catch((error) => {
                    surfaceError(error.message);
                });
        } else {
            surfaceError("Both fields must not be empty");
        }
    });
    // Add event listener for the "See Your Mo Links" link
    document.getElementById('see-links').addEventListener('click', function (e) {
        e.preventDefault(); // Prevent the default link behavior
        chrome.tabs.create({ url: root });
    });
});