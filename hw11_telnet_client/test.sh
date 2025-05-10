#!/usr/bin/env bash
set -xeuo pipefail

go build -o go-telnet

# Запускаем сервер в фоне
(echo -e "Hello\nFrom\nNC\n" && cat 2>/dev/null) | nc -l localhost 4242 >/tmp/nc.out &
NC_PID=$!

# Альтернативная проверка порта без nc -z
for i in {1..10}; do
  # Используем ss для проверки listening-порта
  if ss -tuln | grep -q ':4242 '; then
    break
  fi
  sleep 0.5
done

# Проверяем, что сервер действительно слушает порт
if ! ss -tuln | grep -q ':4242 '; then
  echo "Error: port 4242 not listening"
  kill ${NC_PID} 2>/dev/null || true
  exit 1
fi

# Запускаем клиент
(echo -e "I\nam\nTELNET client\n" && cat 2>/dev/null) | ./go-telnet --timeout=5s localhost 4242 >/tmp/telnet.out &
TL_PID=$!

sleep 5
kill ${TL_PID} 2>/dev/null || true
kill ${NC_PID} 2>/dev/null || true

function fileEquals() {
  local fileData
  fileData=$(cat "$1")
  [ "${fileData}" = "${2}" ] || (echo -e "unexpected output, $1:\n${fileData}" && exit 1)
}

expected_nc_out='I
am
TELNET client'
fileEquals /tmp/nc.out "${expected_nc_out}"

expected_telnet_out='Hello
From
NC'
fileEquals /tmp/telnet.out "${expected_telnet_out}"

rm -f go-telnet
echo "PASS"