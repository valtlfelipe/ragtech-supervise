#!/usr/bin/env bash

mkdir /opt/supervise/log
touch /opt/supervise/log/supsvc.log /opt/supervise/log/device.log /opt/supervise/log/serialhid.log

# Set the monit DB to WAL to allow concurrent readers
sqlite3 /opt/supervise/monit.db "PRAGMA journal_mode=WAL" >/dev/null

# Setup the database in the data directory
ln -sf /data/monit.db /opt/supervise/monit.db
ln -sf /data/monit.db-wal /opt/supervise/monit.db-wal
ln -sf /data/monit.db-shm /opt/supervise/monit.db-shm

# Start supsvc (forks if parent is not PID 1)
echo "[init] starting supsvc"
/opt/supervise/supsvc 2>&1 | sed 's/^/\[main\] /' & main=$!

# Start Prometheus exporter
echo "[init] starting exporter"
/opt/exporter/exporter 2>&1 | sed 's/^/\[exporter\] /' &

# Tail all output log streams with their respective tag
echo "[init] starting log streams"
tail -f /opt/supervise/log/supsvc.log | sed 's/^/\[supsvc\] /' &
tail -f /opt/supervise/log/device.log | sed 's/^/\[device-manager\] /' &
tail -f /opt/supervise/log/serialhid.log | sed 's/^/\[serialhid\] /' &

# Signal handling
trap exit EXIT

# Wait for main process to exit, then kill all other processes
wait $main; retcode=$?
echo "[init] supsvc process terminated, status code: $retcode"
jobs -p | xargs --no-run-if-empty -- kill
wait
exit $retcode