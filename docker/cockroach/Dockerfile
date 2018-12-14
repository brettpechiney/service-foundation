FROM cockroachdb/cockroach

LABEL name="ishi-cockroach"
LABEL description="Initializes CockroachDB with the database and users Ishi wants"

ENV COCKROACH_DB=ishi

RUN apt-get update && apt-get install -y netcat

EXPOSE 8080
EXPOSE 26257

COPY init.sh /cockroach
ENTRYPOINT ["/cockroach/init.sh"]
