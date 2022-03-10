#!/bin/sh


if [[ $1 == "-h" ]];then
  echo "Usage: docker run --net host --rm -it gitar -e [external_ip] ...[args]" 
else
    #Generate cert
    openssl req -x509 -newkey rsa:4096 -nodes -out server.crt -keyout server.key -days 365 -subj "/C=FR/O=krkr/OU=Domain Control Validated/CN=*.ariary.io"
    mv server.* ./certs
    #launch
    ./gitar -copy=false -u exchange -d exchange -tls=true -c ./certs ${@}
fi