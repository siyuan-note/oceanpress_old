function hashAndUpdate(e) {
    scrollIntoView(e.newURL);
}

function scrollIntoView(url) {
    const hash = url.split("#").pop();
    const target = document.querySelector(`[data-block-id="${hash}"]`);
    if (target) {
        var highlightClassName = [ 'hash_selected', 'hash_selected-highlight' ];
        target.classList.add(...highlightClassName);
        target.scrollIntoView();
        setTimeout(() => {
            target.classList.remove(...highlightClassName);

        }, 2000);
    }
}
window.addEventListener('hashchange', hashAndUpdate);
window.addEventListener("load", () => scrollIntoView(location.href));