const root = "https://www.molinks.me"

function validateUrl(url) {
    // can't be longer than 1024 characters
    if (url.length > 1024) {
        surfaceError("url must be 1024 characters or less");
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
function surfaceError(message) {
    document.getElementById("error-text").textContent = message;
    nameGroup.classList.add('error');
}
async function store(key, str) {
    return localStorage.setItem(key, str);
}
async function get(key) {
    return localStorage.getItem(key);
}
let emailVerified;
async function getLoginInfo() {
    const token = await get("token");
    if (!token) {
        throw new Error("Must login");
    }
    return fetch(root + '/____reserved/api/test_cookie?token=' + encodeURIComponent(token)).then(response => response.json()).then(data => {
        if (data.verifiedEmail) {
            emailVerified = true;
        } else {
            emailVerified = false;
            throw new Error("Email not verified");
        }
    });
}
async function refreshLoginInfo() {
    const token = await get("token");
    if (!token) {
        throw new Error("Must login");
    }
    return fetch(root + '/____reserved/api/refresh_token?token=' + encodeURIComponent(token))
    .then(response => response.text()).then(data => {
        store("token", data);
        return getLoginInfo();
    });
}

function login(e) {
    e.preventDefault();
    const email = document.getElementById("email").value
    const password = document.getElementById("password").value
    fetch(root + "/____reserved/api/login?get_token=true", {
        method: "POST",
        body: JSON.stringify({ email, password }),
    })
        .then((res) => {
            if (res.status === 200) {
                console.log("Login successful")
                return res.text().then(text => {
                    document.getElementById("login-content").style.display = "none";
                    document.getElementById("main-content").style.display = "block";
                   return store("token", text);
                }).then(renderPage);
            } else {
                alert("Invalid email or password");
            }
        });
}
function signup(e) {
    e.preventDefault();
    const email = document.getElementById("email").value
    const password = document.getElementById("password").value
    fetch(root + "/____reserved/api/signup?get_token=true", {
        method: "POST",
        body: JSON.stringify({ email, password }),
    })
        .then((res) => {
            if (res.status === 200) {   
                console.log("Login successful")
                return res.text().then(text => {
                    document.getElementById("login-content").style.display = "none";
                    document.getElementById("main-content").style.display = "block";
                    return store("token", text);
                }).then(renderPage);
            } else {
                alert(res.message);
            }
        });
}
// mo/test this is a test link
function initializeMainContent() {
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



    document.getElementById('submit').addEventListener('click', async function () {
        let name = nameInput.value;
        let url = urlInput.value;
        // Sanitize name and url
        name = name.trim();
        url = url.trim();
        const token = await get("token");
        if (validateName(name) && validateUrl(url) && name && url) {
            hideError()
            fetch(root + '/____reserved/api/add?token=' + encodeURIComponent(token), {
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
}

function renderPage() {
    refreshLoginInfo().then(initializeMainContent).catch(() => {
        document.getElementById("main-content").style.display = "none";
        document.getElementById("login-content").style.display = "block";
        if (emailVerified === undefined) {
            document.getElementById("see-links").textContent = "Signup Here";
        } else if(emailVerified === false) {
            document.getElementById("see-links").textContent = "Email not yet verified, check email for verification link";
            document.getElementById("main-content").style.display = "none";
            document.getElementById("login-content").style.display = "none";
        }
    });
}

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('login').addEventListener('click', login);
    document.getElementById('signup').addEventListener('click', signup);
    document.getElementById('see-links').addEventListener('click', async function (e) {
        e.preventDefault();
        chrome.tabs.create({ url: root + "/____reserved/api/give_me_cookie?token=" + encodeURIComponent(await get("token")) });
    });
    renderPage();
});