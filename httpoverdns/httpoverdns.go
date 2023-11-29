package httpoverdns

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

type HTTPOverDNS struct {
	Next plugin.Handler
}

var cache map[string]string = make(map[string]string)

func (e HTTPOverDNS) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	domain := r.Question[0].Name

	if state.QType() == dns.TypeTXT && strings.HasSuffix(domain, ".l.") {

		parts := strings.Split(strings.TrimSuffix(domain, ".l."), ".")

		uid := parts[0]
		isLastChunk := parts[1] == "1"
		chunk := parts[2]

		fmt.Println(parts)

		chunks, exists := cache[uid]
		if !exists {
			chunks = ""
		}

		chunks += chunk

		cache[uid] = chunks

		var text []byte

		if !isLastChunk {
			text = []byte("OK")
		} else {
			bdecoded, err := base64.Encoding.Strict(*base64.StdEncoding).DecodeString(strings.ReplaceAll(strings.TrimSuffix(chunks, ".l."), "_", "="))

			if err != nil {
				text = []byte("1." + err.Error())
			} else {
				decoded := string(bdecoded)

				fmt.Println(decoded)

				req, err := ParseHTTPRequest(decoded)

				fmt.Println(req)

				if err != nil {
					text = []byte("2." + err.Error())
				} else {

					client := &http.Client{
						// Transport: &http2.Transport{},
					}

					resp, err := client.Do(req)

					fmt.Println(resp.StatusCode)

					if err != nil {
						text = []byte("3." + err.Error())
					} else {
						defer resp.Body.Close()

						body, err := io.ReadAll(resp.Body)

						if err != nil {
							text = []byte("4." + err.Error())
						} else {
							text = body
							fmt.Println("Response")
						}
					}
				}
			}
		}

		msg := new(dns.Msg)
		msg.SetReply(r)

		header := dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}

		rchunks := splitText(strings.ReplaceAll(base64.StdEncoding.EncodeToString(text), "=", "_"), 255)

		for _, rchunk := range rchunks {
			txtRecord := &dns.TXT{Hdr: header, Txt: []string{rchunk}}
			msg.Answer = append(msg.Answer, txtRecord)
		}

		w.WriteMsg(msg)

		return 0, nil

	} else {
		return plugin.NextOrFailure(e.Name(), e.Next, ctx, w, r)
	}

}

func (e HTTPOverDNS) Name() string { return "HTTPOverDNS" }

func splitText(text string, n int) []string {
	var chunks []string
	for i := 0; i < len(text); i += n {
		end := i + n
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}
	return chunks
}

func ParseHTTPRequest(rawRequest string) (*http.Request, error) {
	reader := bufio.NewReader(strings.NewReader(rawRequest))

	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	parts := strings.Fields(firstLine)
	method := parts[0]
	url := parts[1]

	headers := make(http.Header)
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" {
			break
		}
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			key := strings.TrimSpace(headerParts[0])
			value := strings.TrimSpace(headerParts[1])
			headers.Add(key, value)
		}
	}

	bodyBytes, err := reader.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(bodyBytes)

	// Create a new request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Set the headers
	req.Header = headers

	return req, nil
}
