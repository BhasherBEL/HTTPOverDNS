# HttpOverDNS

HttpOverDNS is an open-source project that enables tunneling HTTP traffic over DNS using a CoreDNS plugin and a local proxy forwarder. This solution allows users to bypass network restrictions and access HTTP services even in environments where direct access is restricted or monitored.

## Features

- **HTTP Tunneling:** Tunnel HTTP traffic over DNS to bypass network restrictions.
  
- **CoreDNS Integration:** Seamless integration with CoreDNS for intercepting and processing DNS requests.

- **Local Proxy Forwarder:** Handles the encoding and decoding of HTTP traffic between the client and the server.

## Installation

### Building and Installation

1. Clone the CoreDNS repository:

```bash
git clone https://github.com/coredns/coredns.git
```

2. Clone the HttpOverDNS repository:

```bash
git clone https://github.com/BhasherBEL/HttpOverDNS.git
```

2. Copy the HttpOverDNS plugin into CoreDNS:

```bash
cp HttpOverDNS/coredns-http-overdns-plugin coredns/plugin
```

3. Build CoreDNS with the HttpOverDNS plugin:

```bash
go build
```

4. Update your CoreDNS configuration to include the HttpOverDNS plugin.

5. Start the local proxy:
```bash
cd local-proxy
python proxy-get.py
```

## License

HttpOverDNS is licensed under the [MIT License](LICENSE).
