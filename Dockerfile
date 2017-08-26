FROM debian

RUN apt update && \
    apt -y install \
    curl \
    git

RUN L=/usr/local/bin/flynn && \
    curl -sSL -A "`uname -sp`" https://dl.flynn.io/cli | \
    zcat >$L && chmod +x $L

COPY dist/backup_cluster dist/runner /usr/bin/

CMD ["runner"]
