#!/bin/bash

# Массив шаблонов имен файлов и соответствующих таблиц ClickHouse
declare -A patterns=(
    ["result_*.json.gz"]="runtime.history (TagName, DateTime, Value)"
    ["hs_*.json.gz"]="runtime.history (TagName, DateTime, Value)"
    ["av_*.json.gz"]="truckscales.stat (DateTime, CertNum, Tare, Brutto, DriverName, VanNum, CargoType)"
)

# Путь к файлам
path="/var/lib/clickhouse/copyed/"

# Лог-файл
log_file="/var/log/import_script.log"

# Конфигурация клиента ClickHouse
clickhouse_client_config="--password password123 -u admin"

# Перебор всех шаблонов
for pattern in "${!patterns[@]}"; do
    # Проверяем, есть ли файлы, соответствующие шаблону
    files=($(find $path -maxdepth 1 -name "$pattern"))
    if [ ${#files[@]} -gt 0 ]; then
        # Файлы найдены, выполнение запроса INSERT
        for file in "${files[@]}"; do
            clickhouse-client $clickhouse_client_config --query="INSERT INTO ${patterns[$pattern]} FROM INFILE '${file}' COMPRESSION 'gzip' FORMAT JSONEachRow"
            if [ $? -eq 0 ]; then
                # Если запрос выполнен успешно, удаляем файл
                rm "$file"
            else
                echo "Ошибка при импорте файла $file" >>$log_file
            fi
        done
    fi
done

# Перенаправляем все сообщения скрипта в лог-файл
exec 1>>$log_file 2>&1
