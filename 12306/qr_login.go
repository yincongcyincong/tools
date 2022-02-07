package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type QrImage struct {
	Image         string `json:"image"`
	ResultCode    string `json:"result_code"`
	ResultMessage string `json:"result_message"`
	Uuid          string `json:"uuid"`
}

type QrRes struct {
	ResultMessage string `json:"result_message"`
	ResultCode    string `json:"result_code"`
	Uamtk         string `json:"uamtk"`
}

type TkRes struct {
	ResultMessage string `json:"result_message"`
	ResultCode    int    `json:"result_code"`
	Newapptk      string `json:"newapptk"`
}

type UserRes struct {
	ResultMessage string `json:"result_message"`
	ResultCode    int    `json:"result_code"`
	Apptk         string `json:"apptk"`
	Username      string `json:"username"`
}

type LoginRes struct {
	QrRes   *QrRes
	TkRes   *TkRes
	UserRes *UserRes
}

type ApiRes struct {
	ValidateMessagesShowId string                 `json:"validateMessagesShowId"`
	Status                 bool                   `json:"status"`
	HTTPStatus             int                    `json:"httpstatus"`
	Data                   map[string]interface{} `json:"data"`
	Messages               []string               `json:"messages"`
}

var client http.Client
var tmpCookie map[string]string

func init() {
	client = http.Client{
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

	tmpCookie = make(map[string]string)
	tmpCookie["RAIL_DEVICEID"] = "HxVKRYFybjjce3j_YFUSj3YCSikCtGnQMTRB_ivkGogYJI_Zub0z5XAjSE6mis4hAeHrm0b9WIr8rpCwIpTP3wfFa2PUE67-RmNB25iPrKTA1XFxiQk4PywZh0czQHuGNifLJpeXTUzMDC7fRpMy5qH0kWuIktLB"
	tmpCookie["RAIL_EXPIRATION"] = "1644399124143"

}

func QrLogin() *LoginRes {
	initReq, err := http.NewRequest("GET", "https://kyfw.12306.cn/otn/login/init", strings.NewReader(""))
	if err != nil {
		log.Panicln(err)
	}
	initReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	resp, err := client.Do(initReq)
	if err != nil {
		log.Panicln(err)
	}
	addCookie(resp)

	data := make(url.Values)
	data.Set("appid", "otn")

	qrImage := new(QrImage)
	err = request(data, cookieStr(), "https://kyfw.12306.cn/passport/web/create-qr64", qrImage)
	if err != nil {
		log.Panicln(err)
	}
	if qrImage.ResultCode != "0" {
		log.Panicln(qrImage)
	}

	image, err := base64.StdEncoding.DecodeString(qrImage.Image)
	if err != nil {
		log.Panicln(err)
	}
	createQrCode(image)

	// 扫描二维码
	data.Set("uuid", qrImage.Uuid)
	data.Set("RAIL_DEVICEID", tmpCookie["RAIL_DEVICEID"])
	data.Set("RAIL_EXPIRATION", tmpCookie["RAIL_EXPIRATION"])
	qrRes := new(QrRes)
	for i := 0; i < 100; i++ {
		err = request(data, cookieStr(), "https://kyfw.12306.cn/passport/web/checkqr", qrRes)
		if err == nil && qrRes.ResultCode == "2" {
			break
		} else {
			log.Println(err, qrRes.ResultMessage, "继续循环")
		}
		time.Sleep(1 * time.Second)
	}

	// 验证信息，获取tk
	tk := new(TkRes)
	tmpCookie["uamtk"] = qrRes.Uamtk
	fmt.Println(cookieStr())
	err = request(data, cookieStr(), "https://kyfw.12306.cn/passport/web/auth/uamtk", tk)
	if err != nil {
		log.Panicln(err)
	}
	if tk.ResultCode != 0 {
		log.Panicln(tk.ResultMessage)
	}

	// 获取用户信息
	userRes := new(UserRes)
	data.Set("tk", tk.Newapptk)
	tmpCookie["tk"] = tk.Newapptk
	err = request(data, cookieStr(), "https://kyfw.12306.cn/otn/uamauthclient", userRes)
	if err != nil {
		log.Panicln(err)
	}
	if userRes.ResultCode != 0 {
		log.Panicln(userRes.ResultMessage)
	}

	apiRes := new(ApiRes)
	err = request(data, cookieStr(), "https://kyfw.12306.cn/otn/index/initMy12306Api", apiRes)
	if err != nil {
		log.Panicln(err)
	}
	fmt.Println(fmt.Sprintf("%+v", apiRes))

	return &LoginRes{
		TkRes:   tk,
		QrRes:   qrRes,
		UserRes: userRes,
	}
}

func createQrCode(captchBody []byte) {
	imgPath := "./image/qrcode.png"
	err := ioutil.WriteFile(imgPath, captchBody, fs.ModePerm)
	if err != nil {
		log.Panicln(err)
		return
	}
}

func request(data url.Values, tmpCookie, url string, res interface{}) error {

	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", tmpCookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, res)
	if err != nil {
		log.Panicln(err, string(respBody))
		return err
	}

	if url == "https://kyfw.12306.cn/otn/confirmPassenger/checkOrderInfo" {
		fmt.Println(string(respBody))
	}

	return nil
}

func addCookie(resp *http.Response) {
	setCookies := resp.Header.Values("Set-Cookie")

	for _, setCookie := range setCookies {
		cookieKVs := strings.Split(setCookie, ";")
		for _, cookieKV := range cookieKVs {
			cookieKV = strings.TrimSpace(cookieKV)
			cookieSlice := strings.SplitN(cookieKV, "=", 2)
			if len(cookieSlice) >= 2 {
				tmpCookie[cookieSlice[0]] = cookieSlice[1]
			}
		}
	}
}

func cookieStr() string {
	res := ""
	for k, v := range tmpCookie {
		res = fmt.Sprintf("%s%s=%s; ", res, k, v)
	}
	return res
}
