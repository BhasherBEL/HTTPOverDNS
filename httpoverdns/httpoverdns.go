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

func (e HTTPOverDNS) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	domain := r.Question[0].Name

	if state.QType() == dns.TypeTXT && strings.HasSuffix(domain, ".l.") {

		fmt.Println(domain)

		bdecoded, err := base64.Encoding.Strict(*base64.StdEncoding).DecodeString(strings.ReplaceAll(strings.TrimSuffix(domain, ".l."), "_", "="))

		text := []byte("")

		if err != nil {
			text = []byte("1." + err.Error())
		} else {
			// decoded := strings.TrimSpace(string(bdecoded))
			decoded := string(bdecoded)

			fmt.Println("decoded: " + decoded)

			req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(decoded)))

			if err != nil {
				text = []byte("2." + err.Error())
			} else {

				client := &http.Client{}

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
					}
				}
			}
		}

		msg := new(dns.Msg)
		msg.SetReply(r)

		header := dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}

		chunks := splitText(strings.ReplaceAll(base64.StdEncoding.EncodeToString(text), "=", "_"), 255)

		for _, chunk := range chunks {
			txtRecord := &dns.TXT{Hdr: header, Txt: []string{chunk}}
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
