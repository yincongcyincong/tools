package utils

import (
	"fmt"
	"strings"
	"sync"
)

type cookieInfo struct {
	Cookie map[string]string
	lock   sync.Mutex
}

var cookie *cookieInfo

func init() {
	cookie = &cookieInfo{
		Cookie: make(map[string]string),
		lock:   sync.Mutex{},
	}

	cookie.Cookie["RAIL_DEVICEID"] = "NQl1KbtiC9ytGXWvYmSKevxQLhMDHuidN3AhAIoyeatifKs9WHMlOa3zIkpJmTQsj39fUlrwE5ai9tUlTCYu7wZUjHbbPy1KhQbN9QhNgSkeIbUWa8ij_sXKoh2RtFUKogKSq6k3y0Vk2oZJxd0N-UiJzpKgN3sf"
	cookie.Cookie["RAIL_EXPIRATION"] = "1644684709232"

}

func AddCookie(kv map[string]string) {
	cookie.lock.Lock()
	defer cookie.lock.Unlock()
	for k, v := range kv {
		cookie.Cookie[k] = v
	}
}

func GetCookieVal(key string) string {
	cookie.lock.Lock()
	defer cookie.lock.Unlock()
	return cookie.Cookie[key]
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
				cookie.Cookie[cookieSlice[0]] = cookieSlice[1]
			}
		}
	}
}

func GetCookieStr() string {
	res := ""
	cookie.lock.Lock()
	defer cookie.lock.Unlock()
	for k, v := range cookie.Cookie {
		res = fmt.Sprintf("%s%s=%s; ", res, k, v)
	}
	return res
}
