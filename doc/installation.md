# Installation

You can install the Antivirus Scan Service (AV) using `helm` tool. For this case, you can find the chart in the [`charts`](/charts) directory.
    
    ```bash
    helm install av -n $AV_NAMESPACE $AV_CHART_LOCATION 
    ```
    where `av` - modifiable name of helm release.

All available parameters are available in the [`values.yaml`](charts/av-scan-service/values.yaml) file.

Helm chart creates AV deployment and service. Then you can use the service to access the [Antivirus API](/doc/openapi.yaml).

## Chart Configuration

It is possible to pass additional configuration to AV chart.
For the full list of supported options and their defaults, see [`values.yaml`](/charts/av-scan-service/values.yaml).

### Tls in AV

AV API can be configured with tls. It can be done manually or using cert-manager/openshift integration. The steps for every possible approach are provided below.

#### Manually Created tls certificates

1. Create configuration file for SSL certificate:  

```bash
cat <<EOF > server.conf 
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
prompt = no

[req_distinguished_name]
CN = av-scan-service

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = @alt_names
[alt_names]
IP.1 = 127.0.0.1
DNS.1 = <AV_SCAN_SERVICE_FULL_NAME>
DNS.2 = <AV_SCAN_SERVICE_FULL_NAME>.<AV_SCAN_SERVICE_NAMESPACE>
DNS.3 = <AV_SCAN_SERVICE_FULL_NAME>.<AV_SCAN_SERVICE_NAMESPACE>.svc
EOF
```

Where:

* `AV_SCAN_SERVICE_FULL_NAME` is a full name of av-scan-cervice. Default value is `<helm release name>-av-scan-service` (can be overriden with `fullnameOverride` helm parameter);
* `AV_SCAN_SERVICE_NAMESPACE` is a namespace, where av-scan-cervice should be installed;

**Note**: Do not forget to specify any other IP addresses and DNS names that you plan to use to connect to the av-scan-cervice. For this, specify the additional `DNS.#` and `IP.#` fields.

2. Create the CA certificate:

```bash
openssl req -days 730 -nodes -new -x509 -keyout ca.key -out ca.crt -subj "/CN=av-scan-service"
```

3. Create KEY for the av-scan-service:

```bash
openssl genrsa -out av-scan-service.key 2048
```

4. Create CRT file for av-scan-service:

```bash
openssl req -new -key av-scan-service.key -subj "/CN=av-scan-service" -config server.conf | \
openssl x509 -req -days 730 -CA ca.crt -CAkey ca.key -CAcreateserial -out av-scan-service.crt -extensions v3_req -extfile server.conf
```

5. Deploy av-scan-service with following parameters:

```yaml
tls:
    enabled: true
    ca: |   # value from ca.crt
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
    crt: |  # value from av-scan-service.crt
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
    key: |  # value from av-scan-service.key
        -----BEGIN RSA PRIVATE KEY-----
        ...
        -----END RSA PRIVATE KEY-----
```

**Important**: AV scan service handles changes in provided certificate, so if it is changed, AV-scan-service applies the new certificate and continue working with it (without restarts). But if provided certificate is wrong, for example, has extra symbols, AV scan service **do not fail** and continue work with old certificate, until further update.

#### Existed tls secret

**Prerequisites:**
av-scan-service namespace contains the secret with certificates via:

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: <some name>
  namespace: AV_SCAN_SERVICE_NAMESPACE
data:
  ca.crt: <base64 encoded ca certificate>
  tls.crt: <base64 encoded public key>
  tls.key: <base64 encoded private key>
type: kubernetes.io/tls
```

Where:

* `AV_SCAN_SERVICE_NAMESPACE` is a namespace, where av-scan-cervice should be installed;

**Deploy parameters**:

```yaml
tls:
    enabled: true
    secretName: <the name of secret with certificates>
```

**Important**: AV scan service handles changes in the provided certificate. Hence, if it is changed, AV-scan-service applies new certificate and continue working with it (without restarts). But if the provided certificate is wrong, for example, has extra symbols, AV scan service **do not fail** and continue work with old certificate, until further update.

#### Cert-manager integration

**Prerequisites**:

* Cert-manager is installed in your cluster;

**Deploy parameters**:

```yaml
tls:
    enabled: true
    generateCerts:
        executor: cert-manager
        enabled: true
        clusterIssuerName: <the cluster issuer name for certificate>
```

If you do not have cluster issuer, you can generate self-signed certificate with following settings (recommended for test instances only):

```yaml
tls:
    enabled: true
    generateCerts:
        executor: cert-manager
        enabled: true
```

#### Openshift integration

**Prerequisites**:

* Deploy on openshift cluster;

**Deploy parameters**:

```yaml
tls:
    enabled: true
    generateCerts:
        executor: openshift
        enabled: true
```
