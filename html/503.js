(async () => {
    const response = await fetch(window.location.href);
    const headerList = await response.headers;

    const keys = filterHeaders(headerList);
    const allUpstreams = parseUpstreams(keys, headerList);

    const httpUpstreams = allUpstreams.filter(u => {
        const json = JSON.parse(atob(u.value));
        return (json.protocol === 'http' || json.protocol === 'https') &&
               !json.container.startsWith('docker-proxy-sidecar-');
    });
    const streamUpstreams = allUpstreams.filter(u => {
        const json = JSON.parse(atob(u.value));
        return (json.protocol === 'tcp' || json.protocol === 'udp') &&
               !json.container.startsWith('docker-proxy-sidecar-');
    });

    if (httpUpstreams.length > 0) {
        populateTable(httpUpstreams, '#http-upstream-list');
    } else {
        document.getElementById('http-section').style.display = 'none';
    }

    if (streamUpstreams.length > 0) {
        populateTable(streamUpstreams, '#stream-upstream-list');
    } else {
        document.getElementById('stream-section').style.display = 'none';
    }
})();

function parseUpstreams(keys, headerList) {
    const upstreams = [];
    for (const key of keys) {
        upstreams.push({ key: key, value: headerList.get(key) });
    }
    return upstreams;
}

function populateTable(upstreams, tableSelector) {
    const parent = document.querySelector(tableSelector + ' tbody');

    for (const upstream of upstreams) {
        const child = processHeader(upstream.key, upstream.value);
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

function trim(string, charlist = "\\s") {
    return rtrim(ltrim(string, charlist), charlist);
}
