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
    const submitButton = document.getElementById('submit');

    nameInput.addEventListener('input', validateName);

    function validateName() {
        const name = nameInput.value;
        const validRegex = /^[a-zA-Z0-9_-]+$/;
        if (name === '____reserved') {
            document.getElementById("error-text").textContent = "____reserved is... well... reserved.";
            nameGroup.classList.add('error');
            return false;
        }
        if (name === '' || validRegex.test(name)) {
            // insert error message
            nameGroup.classList.remove('error');
            return true;
        } else {
            document.getElementById("error-text").textContent = "Only letters, numbers, _, and - are allowed.";
            nameGroup.classList.add('error');
            return false;
        }
    }

    document.getElementById('submit').addEventListener('click', function () {
        let name = document.getElementById('name').value;
        let url = document.getElementById('url').value;
        // Sanitize name and url
        name = name.trim();
        url = url.trim();
        if (validateName() && name && url) {
            fetch('https://www.molinks.me/____reserved/api/add', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name, url }),
            })
                .then(response => response.json())
                .then(data => {
                    alert('Mo Link created successfully!');
                    // Clear the form
                    document.getElementById('name').value = '';
                    document.getElementById('url').value = '';
                })
                .catch((error) => {
                    console.error('Error:', error);
                    alert('An error occurred. Please try again.');
                });
        } else {
            alert('Please fill in both fields.');
        }
    });
    // Add event listener for the "See Your Mo Links" link
    document.getElementById('see-links').addEventListener('click', function (e) {
        e.preventDefault(); // Prevent the default link behavior
        chrome.tabs.create({ url: 'https://www.molinks.me' });
    });
});