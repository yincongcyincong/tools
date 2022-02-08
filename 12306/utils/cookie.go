package utils

import (
	"fmt"
	"strings"
	"sync"
)

type CookieInfo struct {
	Cookie map[string]string
	lock   sync.Mutex
}

var cookie *CookieInfo

func GetCookie() *CookieInfo {
	if cookie == nil {
		cookie = &CookieInfo{
			Cookie: make(map[string]string),
			lock:   sync.Mutex{},
		}

		cookie.Cookie["RAIL_DEVICEID"] = "HxVKRYFybjjce3j_YFUSj3YCSikCtGnQMTRB_ivkGogYJI_Zub0z5XAjSE6mis4hAeHrm0b9WIr8rpCwIpTP3wfFa2PUE67-RmNB25iPrKTA1XFxiQk4PywZh0czQHuGNifLJpeXTUzMDC7fRpMy5qH0kWuIktLB"
		cookie.Cookie["RAIL_EXPIRATION"] = "1644399124143"

	}

	return cookie
}

func AddCookieStr(setCookies []string) {

	GetCookie().lock.Lock()
	defer GetCookie().lock.Unlock()
	for _, setCookie := range setCookies {
		cookieKVs := strings.Split(setCookie, ";")
		for _, cookieKV := range cookieKVs {
			cookieKV = strings.TrimSpace(cookieKV)
			cookieSlice := strings.SplitN(cookieKV, "=", 2)
			if len(cookieSlice) >= 2 {
				GetCookie().Cookie[cookieSlice[0]] = cookieSlice[1]
			}
		}
	}
}

func GetCookieStr() string {
	res := ""
	GetCookie().lock.Lock()
	defer GetCookie().lock.Unlock()
	for k, v := range GetCookie().Cookie {
		res = fmt.Sprintf("%s%s=%s; ", res, k, v)
	}
	return res
}
