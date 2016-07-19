FROM scratch
ADD ca-certificates.crt /etc/ssl/certs/
COPY baryon /

EXPOSE 8888
ENTRYPOINT ["/baryon", "-p 8888"]
