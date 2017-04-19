openssl genrsa -out ServerKey.pem 1024/2048
openssl req -new -x509 -key ServerKey.pem -out ServerCert.pem -days 1095 -subj "/CN=localhost"

openssl genrsa -out ClientKey.pem 1024/2048
openssl req -new -x509 -key ClientKey.pem -out ClientCert.pem -days 1095 -subj "/CN=localhost"
