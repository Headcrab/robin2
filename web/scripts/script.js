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
    fetch(api + '/api/status/')
        .then(response => response.json())
        .then(data => {
            document.getElementById('dbserver').textContent = data.dbserver;
            document.getElementById('dbtype').textContent = data.dbtype;
            document.getElementById('dbversion').textContent = data.dbversion;
            document.getElementById('dbuptime').textContent = data.dbuptime;
            document.getElementById('appuptime').textContent = data.appuptime;

            const statusElement = document.getElementById('dbstatus');
            statusElement.className = 'status ' + data.dbstatus;
            statusElement.textContent = '';
        })
        .catch(error => console.error('Ошибка:', error));
}

fetchStatus();
setInterval(fetchStatus, 60000);

function getTagOnDate() {
    // Получаем значения полей ввода
    var tag = document.getElementById("searchInput").value;
    var dateFrom = document.getElementById("dateFrom").value;
    var dateTo = document.getElementById("dateTo").value;
    var searchCount = document.getElementById("searchCount").value;

    // Формируем URL с параметрами
    if (!document.getElementById('apiserver')) {
        return;
    }
    api = document.getElementById('apiserver').textContent
    var url = api+"/data/?tag=" + tag + "&from=" + dateFrom + "&to=" + dateTo + "&count=" + searchCount;
    // + "&group=avg";

    // go to url
    loadPage(url);

    // Отправляем GET-запрос
    // fetch(url)
    //     .then(function (response) {
    //         return response.text();
    //     })
    //     .then(function (data) {
    //         // Выводим результат на страницу
    //         document.getElementById("results").textContent = data;
    //     })
    //     .catch(function (error) {
    //         console.log("Произошла ошибка: " + error);
    //     });
}

function getTagList() {
    // Получаем значения полей ввода
    var tag = document.getElementById("searchInput").value;

    // Формируем URL с параметрами
    if (!document.getElementById('apiserver')) {
        return;
    }
    api = document.getElementById('apiserver').textContent
    var url = api+"/tags/?like=" + tag;

    // go to url
    loadPage(url);

}