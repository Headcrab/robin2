const navbar = document.querySelector('.navbar');
const resizeHandle = document.querySelector('.resize-handle');

let isResizing = false;

resizeHandle.addEventListener('mousedown', (e) => {
    isResizing = true;
});

document.addEventListener('mousemove', (e) => {
    if (!isResizing) return;
    navbar.style.width = `${e.clientX}px`;
});

document.addEventListener('mouseup', () => {
    isResizing = false;
});
