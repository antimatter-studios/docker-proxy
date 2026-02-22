(async () => {
    const response = await fetch(window.location.href);
    const headerList = response.headers;

    const keys = filterHeaders(headerList);

    if (keys.length === 0) {
        document.getElementById('upstream-list').style.display = 'none';
        document.getElementById('no-services').style.display = 'block';
        document.getElementById('service-count').textContent = 'No services registered';
    } else {
        const label = keys.length === 1 ? 'service' : 'services';
        document.getElementById('service-count').textContent = keys.length + ' ' + label + ' registered';
        populateTable(keys, headerList);
    }
})();

function populateTable(keys, headerList) {
    const parent = document.querySelector('#upstream-list tbody');

    for (const key of keys) {
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
    const json = JSON.parse(value);
    json.host = ltrim(json.host, "~^\/");
    json.path = ltrim(json.path, "~^\/");

    const url = rtrim(`${json.protocol}://${json.host}/${json.path}`, "\/");

    const template = document.querySelector('#upstream-template');
    const clone = template.content.cloneNode(true);
    const link = clone.querySelector('.url a');
    link.href = url;
    link.textContent = url;
    clone.querySelector('.container').textContent = json.container;

    return clone;
}

function ltrim(string, charlist = "\\s") {
    return string.replace(new RegExp("^[" + charlist + "]+"), "");
}

function rtrim(string, charlist = "\\s") {
    return string.replace(new RegExp("[" + charlist + "]+$"), "");
}
