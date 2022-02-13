package utils

import (
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type cookieInfo struct {
	cookie map[string]string
	lock   sync.Mutex
}

var cookie *cookieInfo

func init() {
	cookie = &cookieInfo{
		cookie: make(map[string]string),
		lock:   sync.Mutex{},
	}

	data := url.Values{}
	data.Set("adblock", "0")
	data.Set("cookieEnabled", "1")
	data.Set("custID", "133")
	data.Set("doNotTrack", "unknown")
	data.Set("flashVersion", "0")
	data.Set("javaEnabled", "0")
	data.Set("jsFonts", "c227b88b01f5c513710d4b9f16a5ce52")
	data.Set("localCode", "3232236206")
	data.Set("mimeTypes", "52d67b2a5aa5e031084733d5006cc664")
	data.Set("os", "MacIntel")
	data.Set("platform", "WEB")
	data.Set("plugins", "d22ca0b81584fbea62237b14bd04c866")
	data.Set("scrAvailSize", strconv.Itoa(rand.Intn(1000)) + "x1920")
	data.Set("srcScreenSize", "24xx1080x1920")
	data.Set("storeDb", "i1l1o1s1")
	data.Set("timeZone", "-8")
	data.Set("touchSupport", "99115dfb07133750ba677d055874de87")
	data.Set("userAgent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36")
	data.Set("webSmartID", "f4e3b7b14cc647e30a6267028ad54c56")
	data.Set("timestamp", strconv.Itoa(int(time.Now().Unix() * 1000)))


	//response = session.httpClint.send(urls.get("GetJS"))
	//result = re.search(r'algID\\x3d(.*?)\\x26', response)

	cookie.cookie["RAIL_DEVICEID"] = "XpiVsNgwTdlWaiJ4o8fyNowlDYx4yAHUvuZYGjWsZ76OeObGN9fv9TX4ZpnTnyy2OkKd755kk2mGCc6mbqDoDzFNjRyPRYfahfklcdWDnsBpHo24jSvpUjy2To00xN8LCYwBVbHzoZagYdfmQU67FclOTJlrgriE"
	cookie.cookie["RAIL_EXPIRATION"] = "1645054427799"

}

func AddCookie(kv map[string]string) {
	cookie.lock.Lock()
	defer cookie.lock.Unlock()
	for k, v := range kv {
		cookie.cookie[k] = v
	}
}

func GetCookieVal(key string) string {
	cookie.lock.Lock()
	defer cookie.lock.Unlock()
	return cookie.cookie[key]
}

func AddCookieStr(setCookies []string) {

	cookie.lock.Lock()
	defer cookie.lock.Unlock()
	for _, setCookie := range setCookies {
		cookieKVs := strings.Split(setCookie, ";")
		for _, cookieKV := range cookieKVs {
			cookieKV = strings.TrimSpace(cookieKV)
			cookieSlice := strings.SplitN(cookieKV, "=", 2)
			if len(cookieSlice) >= 2 {
				cookie.cookie[cookieSlice[0]] = cookieSlice[1]
			}
		}
	}
}

func GetCookieStr() string {
	res := ""
	cookie.lock.Lock()
	defer cookie.lock.Unlock()
	for k, v := range cookie.cookie {
		res = fmt.Sprintf("%s%s=%s; ", res, k, v)
	}
	return res
}
