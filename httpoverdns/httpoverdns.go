package httpoverdns

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
			var err error

			text, err = decodeAndGetAnswer(chunks)

			if err != nil {
				text = []byte(err.Error())
			}
		}

		msg := new(dns.Msg)
		msg.SetReply(r)

		header := dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}

		if len(text) > 350 {
			text = text[:350]
		}

		rchunks := splitText(base64.RawStdEncoding.EncodeToString(text), 255)

		for _, rchunk := range rchunks {
			txtRecord := &dns.TXT{Hdr: header, Txt: []string{rchunk}}
			msg.Answer = append(msg.Answer, txtRecord)
		}

		w.WriteMsg(msg)

		// fmt.Println("Answer lengths:" + len(msg.Answer))
		fmt.Println("Answer length: " + strconv.Itoa(len(msg.Answer)))

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
