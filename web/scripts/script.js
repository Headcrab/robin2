function loadPage(url) {

    p = saveParams();

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
            initialize();
            restoreParams(p);
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
    var url = api + "/data/?tag=" + tag + "&from=" + dateFrom + "&to=" + dateTo + "&count=" + searchCount;
    // + "&group=avg";

    // go to url
    loadPage(url);

    // document.getElementById("searchInput").value = tag;
    // document.getElementById("dateFrom").value = dateFrom;
    // document.getElementById("dateTo").value = dateTo;
    // document.getElementById("searchCount").value = searchCount;

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
    var url = api + "/tags/?like=" + tag;

    // go to url
    loadPage(url);

}

function saveParams() {
    if (document.getElementById("searchInput") != null)
        sessionStorage.setItem("searchInput", document.getElementById("searchInput").value);
    if (document.getElementById("dateFrom") != null)
        sessionStorage.setItem("dateFrom", document.getElementById("dateFrom").value);
    if (document.getElementById("dateTo") != null)
        sessionStorage.setItem("dateTo", document.getElementById("dateTo").value);
    if (document.getElementById("searchCount") != null)
        sessionStorage.setItem("searchCount", document.getElementById("searchCount").value);
}

function restoreParams() {
    if (sessionStorage.getItem("searchInput")) {
        if (document.getElementById("searchInput") != null)
            document.getElementById("searchInput").value = sessionStorage.getItem("searchInput");
        if (document.getElementById("dateFrom") != null)
            document.getElementById("dateFrom").value = sessionStorage.getItem("dateFrom");
        if (document.getElementById("dateTo") != null)
            document.getElementById("dateTo").value = sessionStorage.getItem("dateTo");
        if (document.getElementById("searchCount") != null)
            document.getElementById("searchCount").value = sessionStorage.getItem("searchCount");
    }
}

function getSeason() {
    // Получаем сезон из текущей даты
    var date = new Date();
    var month = date.getMonth();
    var season = Math.floor(month / 3) + 1;
    // winter, spring, summer, fall
    var seasons = ['winter', 'spring', 'summer', 'fall'];
    var seasonName = seasons[season - 1];
    return seasonName;
}

document.addEventListener('DOMContentLoaded', initialize);

function initialize() {
    if (!document.getElementById('robinImage')) {
        return;
    }
    var season = getSeason(); // Вызов вашей функции getSeason
    var imagePath = "../images/robin_" + season + ".png";
    document.getElementById('robinImage').src = imagePath;
}

function loadSwagger() {
    // fetch('/swagger').then(response => {
    //     if (response.ok) {
    //         return response.text();
    //     } else {
    //         throw new Error('Не удалось загрузить Swagger UI');
    //     }
    // }).then(data => {
        if (document.getElementById("content")!=null) {
            document.getElementById("content").innerHTML = '<iframe src="/swagger" style = "text-center" width="800px" height="100%" frameborder="0"></iframe>';
            // Здесь могут быть вызовы для инициализации Swagger UI, если это необходимо
        } else {
            console.error('Element with ID "content" not found');
        }
    // }).catch(error => {
    //     console.error(error);
    // });
}