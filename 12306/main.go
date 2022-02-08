package main

import (
	"encoding/json"
	"fmt"
	"github.com/tools/12306/module"
	"github.com/tools/12306/utils"
	"io/ioutil"
	"net/http"
)

func main() {

	http.HandleFunc("/login", UserLogin)
	http.HandleFunc("/loginOut", UserLogout)
	http.HandleFunc("/search", SearchTrain)
	http.HandleFunc("/repeat", GetRepeatToken)
	http.HandleFunc("/passenger", GetPassenger)
	http.HandleFunc("/buy", StartBuy)
	http.HandleFunc("/test-reg", TestReg)
	http.ListenAndServe("127.0.0.1:8000", nil)
}

func UserLogin(w http.ResponseWriter, r *http.Request) {
	loginRes := QrLogin()
	res, _ := json.Marshal(loginRes)
	fmt.Fprint(w, string(res))
}


func UserLogout(w http.ResponseWriter, r *http.Request) {
	LoginOut()
}

func SearchTrain(w http.ResponseWriter, r *http.Request) {
	searchParam := &module.SearchParam{
		TrainDate:   "2022-02-17",
		FromStation: "BJP",
		ToStation:   "TJP",
	}
	if len(utils.GetCookie().Cookie) == 2 {
		body, _ := ioutil.ReadAll(r.Body)
		utils.AddCookieStr([]string{string(body)})
	}

	GetTrainInfo(searchParam)
}

func GetRepeatToken(w http.ResponseWriter, r *http.Request) {

	if len(utils.GetCookie().Cookie) == 2 {
		body, _ := ioutil.ReadAll(r.Body)
		utils.AddCookieStr([]string{string(body)})
	}
	GetRepeatSubmitToken()
}

func GetPassenger(w http.ResponseWriter, r *http.Request) {

	if len(utils.GetCookie().Cookie) == 2 {
		body, _ := ioutil.ReadAll(r.Body)
		utils.AddCookieStr([]string{string(body)})
	}

	GetPassengers()

	//body, _ := ioutil.ReadAll(r.Body)
	//
	//res := new(PassengerRes)
	//res.Data.NormalPassengers = make([]*Passenger, 0)
	//
	//json.Unmarshal(body, res)
	//
	//fmt.Println(res)
	//fmt.Println(res.Data)
}

func StartBuy(w http.ResponseWriter, r *http.Request) {

	if len(utils.GetCookie().Cookie) == 2 {
		body, _ := ioutil.ReadAll(r.Body)
		utils.AddCookieStr([]string{string(body)})
	}

	CheckUser()
	fmt.Println(utils.GetCookieStr())

	searchParam := &module.SearchParam{
		TrainDate:   "2022-02-17",
		FromStation: "BJP",
		ToStation:   "TJP",
	}
	trainDatas := GetTrainInfo(searchParam)
	passenger := GetPassengers()

	checkOrderRes := CheckOrder(passenger, trainDatas[10])
	fmt.Println(fmt.Sprintf("%+v", checkOrderRes))

	queueRes := GetQueueCount(passenger, trainDatas[10], searchParam)
	fmt.Println(fmt.Sprintf("%+v", queueRes))
	//AutoBuy(passenger, trainDatas[10])


}

func TestReg(w http.ResponseWriter, r *http.Request) {
	CheckUser()

	//body, _ := ioutil.ReadAll(r.Body)
	//matchRes := TokenRe.FindStringSubmatch(string(body))
	//fmt.Println(matchRes)
	//
	//ticketRes := TicketInfoRe.FindSubmatch(body)
	//fmt.Println(ticketRes)
	//
	//orderRes := OrderRequestParam.FindSubmatch(body)
	//fmt.Println(string(orderRes[1]))
	//
	//orderRes[1] = bytes.Replace(orderRes[1], []byte("'"), []byte(`"`), -1)
	//err := json.Unmarshal(orderRes[1], &submitToken.OrderRequestParam)
	//fmt.Println(submitToken.OrderRequestParam, err)
}