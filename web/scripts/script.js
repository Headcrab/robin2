function loadPage(url) {
    const loaderElement = document.getElementById('loader');
    
    if (!loaderElement) {
        console.error('Element with ID "loader" not found.');
        return;
    }

    // Показываем анимацию загрузки
    loaderElement.style.display = 'flex';

    // Загружаем новую страницу
    fetch(url)
        .then(response => response.text())
        .then(data => {
            // Скрываем анимацию загрузки и отображаем новую страницу
            loaderElement.style.display = 'none';
            document.body.innerHTML = data;

            // Обновляем URL в адресной строке
            history.pushState(null, '', url);
        })
        .catch(error => {
            console.error('Error:', error);
        });
}
