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

		//cookie.Cookie["guidesStatus"] = "off"
		//cookie.Cookie["highContrastMode"] = "defaltMode"
		//cookie.Cookie["cursorStatus"] = "off"
		//cookie.Cookie["current_captcha_type"] = "Z"
		//
		//// 第二
		//cookie.Cookie["_jc_save_fromStation"] = "%u5317%u4EAC%2CBJP"
		//cookie.Cookie["_jc_save_toStation"] = "%u5929%u6D25%2CTJP"
		//cookie.Cookie["_jc_save_wfdc_flag"] = "dc"
		//cookie.Cookie["_jc_save_showIns"] = "true"
		//cookie.Cookie["_jc_save_fromDate"] = "2022-02-17"
		//cookie.Cookie["_jc_save_toDate"] = "2022-02-09"
		//
		//// 第三
		//cookie.Cookie["BIGipServerportal"] = "3151233290.17695.0000"
		//cookie.Cookie["BIGipServerpool_passport"] = "165937674.50215.0000"
		//cookie.Cookie["Expires"] = "Thu, 01-Jan-1970 00:00:10 GMT"
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
