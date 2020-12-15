package lib

import (
	"crypto/tls"
	"crypto/x509"
	"log"

	"google.golang.org/grpc/credentials"
)

var certPEMBlock = []byte(`-----BEGIN CERTIFICATE-----
MIIDNzCCAh+gAwIBAgIJALY2lHCZVr3YMA0GCSqGSIb3DQEBCwUAMBgxFjAUBgNV
BAMMDWZiLnpoZW4td28uY24wHhcNMjAxMjE0MDg1NzQzWhcNMzAxMjEyMDg1NzQz
WjBwMQswCQYDVQQGEwJDTjEQMA4GA1UECAwHQmVpamluZzEQMA4GA1UEBwwHQmVp
amluZzEUMBIGA1UECgwLVW5pdGVkU3RhY2sxDzANBgNVBAsMBkRldm9wczEWMBQG
A1UEAwwNZmIuemhlbi13by5jbjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
ggEBAPLY/Xqj5hQnfIf4/oEBRio7enMH8uVJc9IrkL6WH/YB1YaICKuwqXLJTw7j
y3YylEfervsw6Me3a7AYf0ZklnEpuvZ/WzX4PnvvlJPqb0A/SxqNeS9vxHkHHeBM
rru+X+NNlZEE01C+s8ogiO7QD2BDWOHBvp+Xyq9nzqWOnE3lRPZQd9gLtZOkzdSO
Ls0USwqkswLZPNK3HI+xR9R8TVKec83fpF7DnFDH91fm913ftWs09ot/ZkKF1qBa
OBUXo664VfKzMHVs/jZti4tL0niB+ucsOai/t9ReoXP8PsuVZagrPxUaCrkSXLn0
ov/hTh29MGUrlEIM1GGCM1mlnV0CAwEAAaMsMCowKAYDVR0RBCEwH4INZmIuemhl
bi13by5jboIOZmIuemhlbi13by5jb20wDQYJKoZIhvcNAQELBQADggEBAI6J1rl9
KRx5r5NfVCorj4pct82xKR0tcHcw3brBoxHnasvflLkP1R8izMu4g9fxIUIr2Uht
r3tfdZL15+fDTQHfubNgjpbEdg9UPyxAPMdB1rKHcrSddI1k9Z/K806kfcD8yIin
s4muzeicVIgbWA6aFS+h8dRW3kYpISKf5dj+8rj/WOkIcMo4OmyIEfUEicH1rOjG
ddoBiJaOORu9ePhMJfrsFoKJTzbJL6d0TwEIhe2PlUVUzfhMoHH8FFNUFw7unzwy
kwROycfkEavpDmHWUTIEf7kHv7CKbNhOHf+18FWktmwl4SOWstpjdBRDY0e/9wAb
u0p/KgBJATPlV50=
-----END CERTIFICATE-----
`)

var keyPEMBlock = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA8tj9eqPmFCd8h/j+gQFGKjt6cwfy5Ulz0iuQvpYf9gHVhogI
q7CpcslPDuPLdjKUR96u+zDox7drsBh/RmSWcSm69n9bNfg+e++Uk+pvQD9LGo15
L2/EeQcd4Eyuu75f402VkQTTUL6zyiCI7tAPYENY4cG+n5fKr2fOpY6cTeVE9lB3
2Au1k6TN1I4uzRRLCqSzAtk80rccj7FH1HxNUp5zzd+kXsOcUMf3V+b3Xd+1azT2
i39mQoXWoFo4FRejrrhV8rMwdWz+Nm2Li0vSeIH65yw5qL+31F6hc/w+y5VlqCs/
FRoKuRJcufSi/+FOHb0wZSuUQgzUYYIzWaWdXQIDAQABAoIBAQCAA5ijNJDNYP9J
Yh0u/e/xxUbIKpGFApJWYPa9MMAKW28mqsD/WHIKe0n8jGItnX4C4MUWzvJ2jR7s
Rg2Zmt6fKqNO21XGfmTZyjJlQriAgpzhk2AlfGJydijumx2lBDbhyH0mZAfM0apO
y5XDZdQlJ3tMDmihElAa5LrPFP0aJdGxN+2DqNvhyWlesYiRDS7HOoXGSj2HYnwc
pn4SMIwMRvKP2SdWlKEsbKdO6o8RFIIE1Z20UYK47sbrDj5KDGnMnA39zQXp30R1
Da7CD9k8e2rHflZrVa96Pq4R3WfLVkU/o/XBr98lPBr9MvmHRiqL2zCXro4l6C+p
S2I2Mq/BAoGBAPuQNrd2slLuXkG4Y1t1uhy8D96N+tTOIAGdMi1Q+mR7qeX4CsPx
Yrug/cJpE6pCcIEuVmdrh6q1FVczf5A/9sGOzztX7ZaIjlSzwEcHdYqnwQvsH1CA
rcQo+VS2QMOfM4GN4PYh9jhdcADApus32clMAVxnGZ6Og93pFa3nE/7xAoGBAPch
bPwaImmZV83y5o+Xx0KdgU+5+8oBdoqLNW74VP6ctCsFl6UUuwJ78w3CJnsTRYGT
T98tfV4mr5yKhfoL5j9f1uKdPUctKMbHQdQgX3aX8x0JIuels6iTOtoSXjnzY91+
U4XEkdHj67Bb38/ZC6IAHbakQet4yNWeET8ESJ0tAoGBAJzfHX/iwOj+REDvXuYV
z+1DSRIbr6MstsDK6hNgQASRKS2DNBNkX5Fpn1Swedbef5HO94qef4dwTNKIBrBJ
cvLYv1neRwZsOXWQcgLZH+9LFRL+N7jXxYRhmLm+vTw/9rp/Yx2ZqBUWD1YozO45
cdIZV2/rywoZDRpA04gSZWHBAoGAQsY5WHUHT1krrG4xdiMgqBM+2Xf7XL3Adfbf
XTikXpeg5u7/5o8PaMBtEA6hryep5DUVo8v6z/HMCZQ0VzfX4s/WlCzAXfcJyYwV
cWe946FzAylw0P6o6Ke/gyTraOUm2rZDgyV18SyQhnqMovCWgBNf8/W2ChX8zhuD
tW9G35ECgYA3/sJz73J0a825tq2bsW6mqyGtY16nNLpPFjpggktcIFUv7Igk5jeC
UHmDH6fcNl5HHuzC2PXrQ9h46mrhqBSgXWWSW1uiVG33SDIbX1ZimZxryQdfaB9S
NUyQ5E0nSPv4Hwwp2TzAv7XdsYUqrxXfkoysQ0/YvqVM5pKGiZxJVA==
-----END RSA PRIVATE KEY-----
`)

// ServerTLS server tls
func ServerTLS() credentials.TransportCredentials {
	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		log.Fatal(err)
	}

	return credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})
}

// ClientTLS client tls
func ClientTLS(host string) credentials.TransportCredentials {
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(certPEMBlock) {
		log.Fatal("Credentials: Failed to append certificates")
	}

	return credentials.NewClientTLSFromCert(cp, host)
}

// openssl req -new -sha256 \
//     -key ca.key \
//     -subj "/C=CN/ST=Beijing/L=Beijing/O=UnitedStack/OU=Devops/CN=fb.zhen-wo.cn" \
//     -reqexts SAN \
//     -config <(cat /System/Library/OpenSSL/openssl.cnf \
//         <(printf "[SAN]\nsubjectAltName=DNS:fb.zhen-wo.cn,DNS:fb.zhen-wo.com")) \
//     -out zchd.csr

// openssl x509 -req -days 3650 \
//     -in zchd.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
//     -extfile <(printf "subjectAltName=DNS:fb.zhen-wo.cn,DNS:fb.zhen-wo.com") \
//     -out zchd.crt
