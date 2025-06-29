FROM elastic/filebeat:7.17.7

# Копируем файл конфигурации в нужное место
COPY filebeat.yml /usr/share/filebeat/filebeat.yml

# Устанавливаем права и владельца
USER root
RUN chown root:root /usr/share/filebeat/filebeat.yml && chmod 644 /usr/share/filebeat/filebeat.yml
