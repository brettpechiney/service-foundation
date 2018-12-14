#!/usr/bin/env bash

set -e
set -x

SQL="/cockroach/cockroach.sh sql"

while ! nc -w 1 -z localhost 26257; do sleep 0.1; done;
/${SQL} -e "
    CREATE DATABASE IF NOT EXISTS ishi;
    CREATE USER IF NOT EXISTS maxroach;
    GRANT ALL ON DATABASE ishi TO maxroach;"