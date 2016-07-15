FROM alpine
ADD ca-certificates.crt /etc/ssl/certs/


COPY baryon /
COPY scripts/baryon.sh /baryon.sh
RUN chmod 755 /baryon.sh /baryon

EXPOSE 8888
ENTRYPOINT exec /baryon.sh
