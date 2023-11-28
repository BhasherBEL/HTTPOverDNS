package httpoverdns

import (
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

		bdecoded, err := base64.Encoding.Strict(*base64.RawStdEncoding).DecodeString(strings.TrimSuffix(domain, ".l."))

		fmt.Println(err)

		text := ""

		if err != nil {
			text = err.Error()
		} else {
			decoded := string(bdecoded)

			fmt.Println(decoded)

			resp, err := http.Get(decoded)

			fmt.Println(resp.StatusCode)

			if err != nil {
				text = err.Error()
			} else {
				body, err := io.ReadAll(resp.Body)

				// fmt.Println(string(body))

				if err != nil {
					text = err.Error()
				} else {
					text = base64.RawStdEncoding.EncodeToString(body)
				}
			}
		}

		msg := new(dns.Msg)
		msg.SetReply(r)

		header := dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}

		chunks := splitText(text, 255)

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
