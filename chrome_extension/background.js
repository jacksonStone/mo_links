const searchSites = [
  {host: "www.google.com", path: "/search", query: "q" },
  {host: "duckduckgo.com", path: "/", query: "q" },
  {host: "www.bing.com", path: "/search", query: "q" },
  {host: "search.yahoo.com", path: "/search", query: "p" },
  {host: "www.perplexity.ai", path: "/search", query: "q" },
  {host: "yandex.com", path: "/search", query: "text" },
]

  chrome.webNavigation.onBeforeNavigate.addListener((details) => {
    const url = new URL(details.url);
    for (const site of searchSites) {
      if (url.hostname === site.host && url.pathname === site.path) {
        const searchParams = new URLSearchParams(url.search);
        const searchTerm = searchParams.get(site.query);
        // If the browser attempts to pipe exactly mo or mo/ to the search engine, redirect to the main molink page
        // as this is the user typing mo directly into the URL address bar
        if (searchTerm == "mo") {
          chrome.tabs.update(details.tabId, {url: "https://www.molinks.me/____reserved/hop?next="});
        }
        // If the browser attempts to pipe exactly mo/some-thing to the search engine, redirect to the main molink page
        // as this is the user typing mo directly into the URL address bar
        else if (searchTerm.startsWith("mo/")) {
          chrome.tabs.update(details.tabId, {url: "https://www.molinks.me/____reserved/hop?next=" + encodeURIComponent(searchParams.get(site.query).substring(3))});
        } 
      }
    }
  });