FROM clickhouse/clickhouse-server:25.8.18.1

ENV TZ=Asia/Almaty
# Установка cron и инструментов для импорта архивов
RUN apt-get update \
    && apt-get install -y --no-install-recommends tzdata cron mc sshfs python3 \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /var/lib/clickhouse/copyed

# Добавление скрипта
ADD script_download.sh /etc/cron.d/script_download.sh
COPY init/ /docker-entrypoint-initdb.d/
COPY transform_truckscales_json.py /usr/local/bin/transform_truckscales_json.py
# Дать права на исполнение скрипту
RUN chmod +x /etc/cron.d/script_download.sh /usr/local/bin/transform_truckscales_json.py
# Добавление задач в cron
RUN echo "* * * * * root /etc/cron.d/script_download.sh >> /var/log/cron.log 2>&1" >> /etc/cron.d/cron-task
# Ежемесячное удаление логов cron
RUN echo "0 0 1 * * root rm /var/log/import_script.log" >> /etc/cron.d/cron-task
RUN echo "0 0 1 * * root rm /var/log/cron.log" >> /etc/cron.d/cron-task
# Дать права файлу cron-task
RUN chmod 0644 /etc/cron.d/cron-task

# Создание папок
# RUN apt-get update && apt-get install -y sshfs
# RUN mkdir ~/usr0
# RUN sshfs usr0@139.158.31.218:usr0 ~/usr0

# Создать модифицированный entrypoint.sh
RUN echo "#!/bin/bash" > /entrypoint_modified.sh
RUN echo "cron -f &" >> /entrypoint_modified.sh
RUN echo "echo cron started" >> /entrypoint_modified.sh
RUN echo "/entrypoint.sh" >> /entrypoint_modified.sh

# Сделать его исполняемым
RUN chmod +x /entrypoint_modified.sh

# Запуск cron и ClickHouse
ENTRYPOINT ["/entrypoint_modified.sh"]
