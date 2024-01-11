package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
	"golang.org/x/net/publicsuffix"
)

type Session struct {
	client  *http.Client
	headers http.Header
}

func proxyDialer(proxyURL string) (proxy.ContextDialer, error) {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
	}

	var cntx_dialer proxy.ContextDialer = proxy.Direct

	switch parsedURL.Scheme {
	case "socks5":
		username := parsedURL.User.Username()
		passwd, _ := parsedURL.User.Password()
		auth := &proxy.Auth{
			User:     username,
			Password: passwd,
		}
		dialer, err := proxy.SOCKS5("tcp", parsedURL.Host, auth, proxy.Direct)
		if err != nil {
			ErrorLogger.Printf("failed to create socks5 proxy dialer: %v", err.Error())
			return nil, err
		}
		cntx_dialer = dialer.(proxy.ContextDialer)
	case "http", "https":
		dialer, err := proxy.FromURL(parsedURL, proxy.Direct)
		if err != nil {
			ErrorLogger.Printf("failed to create http proxy dialer: %v", err.Error())
			return nil, err
		}
		cntx_dialer = dialer.(proxy.ContextDialer)
	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %s", parsedURL.Scheme)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy dialer: %w", err)
	}
	return cntx_dialer, nil
}

type Response struct {
	body   []byte
	status int
	err    error
}

func (res *Response) String() string {
	return string(res.body)
}
func (res *Response) Error() string {
	return res.err.Error()
}

func custm_err(text string, err error) error {
	return fmt.Errorf("%v: %v", text, err.Error())
}
func CreateSession() Session {
	session := Session{}
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	client := &http.Client{Jar: jar}
	session.client = client
	session.headers = generate_headers()
	return session
}

func read_body(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

func get_reader(datastr string) *bytes.Reader {
	data := []byte(datastr)
	return bytes.NewReader(data)
}

func Check_localhost(proxy string) *tls.Config {
	if strings.Contains(proxy, "127.0.0.1") || strings.Contains(proxy, "localhost") {
		return &tls.Config{InsecureSkipVerify: true}
	} else {
		return nil
	}
}

func (session *Session) Set_proxy(proxy string) error {
	proxy = strings.Replace(proxy, "\r", "", -1)
	dialer, err := proxyDialer(proxy)
	if err != nil {
		return err
	}
	tr := &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: Check_localhost(proxy)}
	session.client.Transport = tr
	session.client.Timeout = 15 * time.Second
	return nil
}

func (session *Session) Getreq(url string) *Response {
	return session.send_req(url, "GET", nil)
}

func (session *Session) Postreq(url string, data_str string) *Response {
	return session.send_req(url, "POST", get_reader(data_str))
}

func (session *Session) Patchreq(url string, data_str string) *Response {
	return session.send_req(url, "PATCH", get_reader(data_str))
}

func (session *Session) send_req(url string, method string, reader io.Reader) *Response {
	result := &Response{}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		result.err = custm_err("Error on create request", err)
		return result
	}

	session.headers.Del("Cookie")
	req.Header = session.headers

	resp, err := session.client.Do(req)

	if err != nil {
		result.err = custm_err("Error on send requests", err)
		return result
	}

	body, err := read_body(resp)

	if err != nil {
		result.err = custm_err("Error on read result response", err)
		return result
	}
	result.body = body
	result.status = resp.StatusCode
	return result
}

func generate_agent() string {
	agents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/109.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13.2; rv:109.0) Gecko/20100101 Firefox/109.0",
		"Mozilla/5.0 (X11; Linux i686; rv:109.0) Gecko/20100101 Firefox/109.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/109.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux i686; rv:109.0) Gecko/20100101 Firefox/109.0",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/109.0",
		"Mozilla/5.0 (X11; Fedora; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/109.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.3 Safari/605.1.15",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; Trident/4.0)",
		"Mozilla/4.0 (compatible; MSIE 9.0; Windows NT 6.0; Trident/5.0)",
		"Mozilla/4.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; Trident/6.0)",
		"Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (Windows NT 6.2; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (Windows NT 6.3; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (Windows NT 10.0; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Edg/109.0.1518.69",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Edg/109.0.1518.69",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 OPR/94.0.4606.65",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 OPR/94.0.4606.65",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 OPR/94.0.4606.65",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 OPR/94.0.4606.65",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Vivaldi/5.6.2867.62",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Vivaldi/5.6.2867.62",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Vivaldi/5.6.2867.62",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Vivaldi/5.6.2867.62",
		"Mozilla/5.0 (X11; Linux i686) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36 Vivaldi/5.6.2867.62",
		"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 YaBrowser/23.1.0 Yowser/2.5 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 YaBrowser/23.1.0 Yowser/2.5 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 YaBrowser/23.1.0 Yowser/2.5 Safari/537.36",
	}
	randomIndex := rand.Intn(len(agents))
	agent := agents[randomIndex]
	return agent
}

func generate_headers() http.Header {
	headers := http.Header{}
	agent := generate_agent()
	headers.Set("User-Agent", agent)
	headers.Set("Content-Type", "application/json")
	return headers
}
