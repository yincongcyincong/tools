package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var client *http.Client

func GetClient() *http.Client {
	if client == nil {
		client = &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 5 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   5 * time.Second,
				ResponseHeaderTimeout: 5 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				MaxIdleConnsPerHost:   -1,
			},
		}
	}

	return client

}

func Request(data url.Values, cookieStr, url string, res interface{}, headers map[string]string) error {

	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", cookieStr)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	req.Header.Set("Host", "kyfw.12306.cn")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://kyfw.12306.cn")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := GetClient().Do(req)
	if err != nil {
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, res)
	if err != nil {
		log.Panicln(err, string(respBody), url)
		return err
	}

	// 添加cookie
	setCookies := resp.Header.Values("Set-Cookie")
	AddCookieStr(setCookies)

	if url == "https://kyfw.12306.cn/otn/confirmPassenger/checkOrderInfo" || url == "https://kyfw.12306.cn/otn/confirmPassenger/getQueueCountAsync" || url == "https://kyfw.12306.cn/otn/leftTicket/submitOrderRequest" {
		fmt.Println(url, string(respBody))
	}

	return nil
}
