package main

import (
	"encoding/base64"
	"fmt"
	"github.com/tools/12306/module"
	"github.com/tools/12306/utils"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var loginUser *module.LoginUser

func init()  {
	loginUser = new(module.LoginUser)
}

func QrLogin() {
	initReq, err := http.NewRequest("GET", "https://kyfw.12306.cn/otn/login/init", strings.NewReader(""))
	if err != nil {
		log.Panicln(err)
	}
	initReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	resp, err := utils.GetClient().Do(initReq)
	if err != nil {
		log.Panicln(err)
	}
	utils.AddCookieStr(resp.Header.Values("Set-Cookie"))

	data := make(url.Values)
	data.Set("appid", "otn")
	qrImage := new(module.QrImage)
	err = utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/passport/web/create-qr64", qrImage, nil)
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
	data.Set("RAIL_DEVICEID", utils.GetCookie().Cookie["RAIL_DEVICEID"])
	data.Set("RAIL_EXPIRATION", utils.GetCookie().Cookie["RAIL_EXPIRATION"])
	qrRes := new(module.QrRes)
	for i := 0; i < 100; i++ {
		err = utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/passport/web/checkqr", qrRes, nil)
		if err == nil && qrRes.ResultCode == "2" {
			break
		} else {
			log.Println(err, qrRes.ResultMessage, "继续循环")
		}
		time.Sleep(1 * time.Second)
	}
	loginUser.QrRes = qrRes

	// 验证信息，获取tk
	tk := new(module.TkRes)
	utils.GetCookie().Cookie["uamtk"] = qrRes.Uamtk
	err = utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/passport/web/auth/uamtk", tk, nil)
	if err != nil {
		log.Panicln(err)
	}
	if tk.ResultCode != 0 {
		log.Panicln(tk.ResultMessage)
	}
	loginUser.TkRes = tk

	GetLoginData()
}


func GetLoginData() {
	data := make(url.Values)
	data.Set("appid", "otn")
	// 获取用户信息
	userRes := new(module.UserRes)
	data.Set("tk", loginUser.TkRes.Newapptk)
	utils.GetCookie().Cookie["tk"] = loginUser.TkRes.Newapptk
	err := utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/otn/uamauthclient", userRes, nil)
	if err != nil {
		log.Panicln(err)
	}
	if userRes.ResultCode != 0 {
		log.Panicln(userRes.ResultMessage)
	}

	apiRes := new(module.ApiRes)
	err = utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/otn/index/initMy12306Api", apiRes, nil)
	if err != nil {
		log.Panicln(err)
	}
	loginUser.ApiRes = apiRes
	fmt.Println(fmt.Sprintf("%+v", apiRes))

}

func LoginOut() {
	req, err := http.NewRequest("GET", "https://kyfw.12306.cn/otn/login/loginOut", strings.NewReader(""))
	if err != nil {
		log.Panicln(err)
	}
	req.Header.Set("Cookie", utils.GetCookieStr())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := utils.GetClient().Do(req)
	if err != nil {
		log.Panicln(err)
	}
	fmt.Println(resp.StatusCode)
}

func createQrCode(captchBody []byte) {
	imgPath := "./image/qrcode.png"
	err := ioutil.WriteFile(imgPath, captchBody, fs.ModePerm)
	if err != nil {
		log.Panicln(err)
		return
	}
}
