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

function fetchStatus() {
    if (!document.getElementById('apiserver')) {
        return;
    }
    api = document.getElementById('apiserver').textContent
    fetch(api+'/api/status/')
        .then(response => response.json())
        .then(data => {
            document.getElementById('dbserver').textContent = data.dbserver;
            document.getElementById('dbversion').textContent = data.dbversion;
            document.getElementById('dbuptime').textContent = data.dbuptime;
            document.getElementById('appuptime').textContent = data.appuptime;

            const statusElement = document.getElementById('dbstatus');
            statusElement.className = 'status ' + data.dbstatus;
            statusElement.textContent = '';
        })
        .catch(error => console.error('Ошибка:', error));
}

setInterval(fetchStatus, 1000);
fetchStatus();
