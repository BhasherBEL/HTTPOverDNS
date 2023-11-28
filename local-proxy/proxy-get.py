from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib import request
import shutil
import dns.resolver
import base64

PORT = 1080

resolver = dns.resolver.Resolver(configure=False)
resolver.nameservers = ['127.0.0.1']

class HTTPProxy(BaseHTTPRequestHandler):
    def do_GET(self):
        url = self.path
        # print(self.raw_requestline)
        domain = base64.b64encode(bytes(url, 'utf-8')).decode('utf-8').replace('=', '') + '.l'
        print(domain)
        try:
            r = resolver.query(domain, 'TXT')

            rep = ''

            for i in r.response.answer:
                for j in i.items:
                    rep += j.to_text()

            response = base64.b64decode(rep)

            self.send_response(200)
            self.end_headers()
            self.wfile.write(response)
        except Exception as e:
            self.send_error(500, str(e))


httpd = HTTPServer(('', PORT), HTTPProxy)
print("Now serving at", PORT)
httpd.serve_forever()
