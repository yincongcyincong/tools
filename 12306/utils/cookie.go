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

		cookie.Cookie["RAIL_DEVICEID"] = "NQl1KbtiC9ytGXWvYmSKevxQLhMDHuidN3AhAIoyeatifKs9WHMlOa3zIkpJmTQsj39fUlrwE5ai9tUlTCYu7wZUjHbbPy1KhQbN9QhNgSkeIbUWa8ij_sXKoh2RtFUKogKSq6k3y0Vk2oZJxd0N-UiJzpKgN3sf"
		cookie.Cookie["RAIL_EXPIRATION"] = "1644684709232"

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
