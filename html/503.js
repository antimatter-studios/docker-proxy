(async () => {
    const response = await fetch(window.location.href);
    const headerList = await response.headers;
    
    updatePage(headerList);
})();

function updatePage(headerList) {
    const parent = document.querySelector('#upstream-list tbody');

    const keys = filterHeaders(headerList);

    for(const key of keys) {
        const child = processHeader(key, headerList.get(key));
        parent.appendChild(child);
    }
}

function filterHeaders(headerList) {
    const keys = Object.keys(Object.fromEntries(headerList));
    keys.sort((a, b) => a.localeCompare(b, undefined, { numeric: true, sensitivity: 'base' }));
    
    return keys.filter(key => key.startsWith('x-proxy-upstream'));
}

function processHeader(key, value) {
    value = atob(value);
    json = JSON.parse(value);
    json.host = ltrim(json.host, "~^\/");
    json.path = ltrim(json.path, "~^\/");

    const url = rtrim(`${json.protocol}://${json.host}/${json.path}`, "\/");

    const template = document.querySelector('#upstream-template');
    const clone = template.content.cloneNode(true);
    clone.querySelector('.url').textContent = url;
    clone.querySelector('.container').textContent = json.container;

    return clone;
}

function ltrim(string, charlist = "\\s") {
    return string.replace(new RegExp("^[" + charlist + "]+"), "");
}

function rtrim(string, charlist = "\\s") {
    return string.replace(new RegExp("[" + charlist + "]+$"), "");
}

function trim(string, charlist = "\\s") {
    return rtrim(ltrim(string, charlist), charlist);
}
