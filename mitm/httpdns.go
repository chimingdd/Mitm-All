package mitm

import (
	"bytes"
	"encoding/json"
	yaklog "github.com/yaklang/yaklang/common/log"
	"io"
	"net/http"
	"socks2https/pkg/color"
	"socks2https/pkg/dnsutils"
	"socks2https/setting"
)

// todo 并发锁
var (
	Domain2IP = make(map[string]map[string]struct{})
	IP2Domain = make(map[string][]string)
)

var DNSRequest = ModifyRequest(func(req *http.Request, ctx *Context) (*http.Request, *http.Response) {
	if _, ok := Domain2IP[req.Host]; !ok {
		ipv4s, err := dnsutils.DNS2IPv4(req.Host, setting.Config.DNS)
		if err != nil {
			yaklog.Warnf(color.SetColor(color.MAGENTA_COLOR_TYPE, err))
			return req, nil
		}
		for _, ipv4 := range ipv4s {
			if _, ok = Domain2IP[req.Host][ipv4]; !ok {
				Domain2IP[req.Host] = make(map[string]struct{})
				Domain2IP[req.Host][ipv4] = struct{}{}
				IP2Domain[ipv4] = append(IP2Domain[ipv4], req.Host)
			}
		}
	}
	return req, nil
})

type BaiduHTTPDNS struct {
	Clientip string `json:"clientip"`
	Data     map[string]struct {
		IPv4 struct {
			Ip  []string `json:"ip"`
			Ttl int      `json:"ttl"`
			Msg string   `json:"msg"`
		} `json:"ipv4"`
		IPv6 struct {
			Ip  []string `json:"ip"`
			Ttl int      `json:"ttl"`
			Msg string   `json:"msg"`
		} `json:"ipv6"`
	} `json:"data"`
	Msg      string `json:"msg"`
	Serverip struct {
		Ipv4 []string `json:"ipv4"`
	} `json:"serverip"`
	Timestamp int `json:"timestamp"`
}

var HTTPDNSResponse = ModifyResponse(func(resp *http.Response, ctx *Context) *http.Response {
	switch ctx.Request.Host {
	case "httpdns.baidubce.com":
		baiduDNS := &BaiduHTTPDNS{}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			yaklog.Warnf("read Response Body failed : %v", err)
		}
		resp.Body = io.NopCloser(bytes.NewReader(body))
		if err = json.Unmarshal(body, baiduDNS); err != nil {
			yaklog.Warnf("parse Response Body failed : %v", err)
		}
		for domain, ip := range baiduDNS.Data {
			for _, ipv4 := range ip.IPv4.Ip {
				if _, ok := Domain2IP[domain][ipv4]; !ok {
					Domain2IP[domain] = make(map[string]struct{})
					Domain2IP[domain][ipv4] = struct{}{}
					IP2Domain[ipv4] = append(IP2Domain[ipv4], domain)
				}
			}
		}
	}
	return resp
})
