package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cihub/seelog"
	"github.com/tools/12306/module"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type cookieInfo struct {
	cookie map[string]string
	lock   sync.Mutex
}

var (
	cookie  *cookieInfo
	AlgIDRe = regexp.MustCompile("algID(.*?)x26")
)

func init() {
	cookie = &cookieInfo{
		cookie: make(map[string]string),
		lock:   sync.Mutex{},
	}

	// 动态获取设备信息
	body, err := RequestGetWithoutJson("", "https://kyfw.12306.cn/otn/HttpZF/GetJS", nil)
	if err != nil {
		seelog.Error(err)
		return
	}

	matchData := AlgIDRe.FindSubmatch(body)
	if len(matchData) < 2 {
		seelog.Error("get algID fail")
		return
	}
	algId := strings.TrimLeft(string(matchData[1]), `\x3d`)
	algId = strings.TrimRight(algId, `\`)

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
	data.Set("scrAvailSize", strconv.Itoa(rand.Intn(1000))+"x1920")
	data.Set("srcScreenSize", "24xx1080x1920")
	data.Set("storeDb", "i1l1o1s1")
	data.Set("timeZone", "-8")
	data.Set("touchSupport", "99115dfb07133750ba677d055874de87")
	data.Set("userAgent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36")
	data.Set("webSmartID", "f4e3b7b14cc647e30a6267028ad54c56")
	data.Set("timestamp", strconv.Itoa(int(time.Now().Unix()*1000)))
	data.Set("algID", algId)
	body, err = RequestGetWithoutJson("", "https://kyfw.12306.cn/otn/HttpZF/logdevice?"+data.Encode(), nil)
	if err != nil {
		seelog.Error(err)
		return
	}
	if bytes.Contains(body, []byte("callbackFunction")) {
		body = bytes.TrimLeft(body, "callbackFunction('")
		body = bytes.TrimRight(body, "')")
		deviceInfo := new(module.DeviceInfo)
		err = json.Unmarshal(body, deviceInfo)
		if err != nil {
			seelog.Error(err)
			return
		}
		cookie.cookie["RAIL_DEVICEID"] = deviceInfo.Dfp
		cookie.cookie["RAIL_EXPIRATION"] = deviceInfo.Exp
	}

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
