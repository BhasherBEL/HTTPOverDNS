package httpoverdns

import (
	"bufio"
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
var history map[string]int = make(map[string]int)
var sendQueue map[string][]string = make(map[string][]string)

func (e HTTPOverDNS) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	domain := r.Question[0].Name

	if state.QType() == dns.TypeTXT && strings.HasSuffix(domain, ".l.") {

		_, inHistory := history[domain]

		if inHistory {
			return 0, nil
		} else {
			history[domain] = 0
		}

		parts := strings.Split(strings.TrimSuffix(domain, ".l."), ".")

		msg := new(dns.Msg)
		msg.SetReply(r)

		header := dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}

		fmt.Println(parts)

		if len(parts) == 2 {
			queue, exists := sendQueue[parts[0]]

			if !exists {
				txtRecord := &dns.TXT{Hdr: header, Txt: []string{"1."}}
				msg.Answer = []dns.RR{txtRecord}
			} else {
				if len(queue) == 1 {
					txtRecord := &dns.TXT{Hdr: header, Txt: []string{"1." + queue[0]}}
					msg.Answer = []dns.RR{txtRecord}
					delete(sendQueue, parts[0])
				} else {
					txtRecord := &dns.TXT{Hdr: header, Txt: []string{"0." + queue[0]}}
					msg.Answer = []dns.RR{txtRecord}
					sendQueue[parts[0]] = sendQueue[parts[0]][1:]
				}
			}
		} else {

			uid := parts[0]
			isLastChunk := parts[1] == "1"
			chunk := parts[2]

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
				var err error

				text, err = decodeAndGetAnswer(chunks)

				if err != nil {
					text = []byte(err.Error())
				}
			}

			rchunks := splitText(base64.RawStdEncoding.EncodeToString(text), 250)

			partBit := "1"

			if len(rchunks) > 1 {
				sendQueue[uid] = rchunks[1:]
				partBit = "0"
			}

			txtRecord := &dns.TXT{Hdr: header, Txt: []string{partBit + "." + rchunks[0]}}
			msg.Answer = []dns.RR{txtRecord}
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

func decodeAndGetAnswer(encoded string) ([]byte, error) {
	bdecoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(strings.TrimSuffix(encoded, ".l."), "_", "="))

	if err != nil {
		return nil, err
	}

	decoded := string(bdecoded)

	req, err := ParseHTTPRequest2(decoded)
	// req, err = http.NewRequest("GET", "http://httpforever.com", nil)

	if err != nil {
		return nil, err
	}

	client := &http.Client{
		// Transport: &http2.Transport{},
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}
func ParseHTTPRequest2(rawRequest string) (*http.Request, error) {
	reader := bufio.NewReader(strings.NewReader(rawRequest))

	req, err := http.ReadRequest(reader)

	if err != nil {
		return nil, err
	}

	clientReq := &http.Request{
		Method: req.Method,
		URL:    req.URL,
		Header: req.Header,
		Body:   req.Body,
	}

	return clientReq, nil
}
