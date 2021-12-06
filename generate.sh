openssl req -x509 -newkey rsa:4096 -nodes -out server.crt -keyout server.key -days 365 -subj "/C=FR/O=krkr/OU=Domain Control Validated/CN=*.ariary.io"
mkdir -p $HOME/.gitar/certs
cp server.crt $HOME/.gitar/certs && cp server.key $HOME/.gitar/certs
