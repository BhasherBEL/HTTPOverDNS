from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib import request
import shutil
import dns.resolver
import base64
import random

PORT = 1080

resolver = dns.resolver.Resolver(configure=False)
resolver.nameservers = ['127.0.0.1']

def split_string_into_chunks(input_string, chunk_size):
    return [input_string[i:i + chunk_size] for i in range(0, len(input_string), chunk_size)]


class HTTPProxy(BaseHTTPRequestHandler):
    def get_raw_request(self):
        # raw_request = self.raw_requestline
        # raw_request += b'\r\n' + self.headers.as_bytes()
        # raw_request += b'\r\n\r\n' + self.rfile.read(int(self.headers['Content-Length']))

        raw_request = self.raw_requestline + b'\r\n' + self.headers.as_bytes()

        content_length = self.headers.get('Content-Length')
        if content_length:
            raw_request += b'\r\n\r\n' + self.rfile.read(int(content_length))

        return raw_request


    def do_GET(self):
        self.handle_request()

    def do_POST(self):
        self.handle_request()

    def do_DELETE(self):
        self.handle_request()

    def do_PUT(self):
        self.handle_request()

    def handle_request(self):
        print("RAW:", self.get_raw_request())
        content = base64.b64encode(self.get_raw_request()).decode('utf-8').replace('=', '_')

        rid = random.randint(0, 99999)

        chunks = split_string_into_chunks(content, 53)

        try:
            for i, chunk in enumerate(chunks):
                domain = f'{rid}.{int(i == len(chunks)-1)}.{chunk}.l' 
                r = resolver.resolve(domain, 'TXT')
            
            rep = ''

            for i in r.response.answer:
                for j in i.items:
                    rep += j.to_text().replace('_', '=')

            response = base64.b64decode(rep)

            self.send_response(200)
            self.end_headers()
            self.wfile.write(response)

        except Exception as e:
            self.send_error(500, str(e))

httpd = HTTPServer(('', PORT), HTTPProxy)
print("Now serving at", PORT)
httpd.serve_forever()
