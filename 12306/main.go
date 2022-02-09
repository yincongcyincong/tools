package main

import (
	"fmt"
	"github.com/tools/12306/module"
	"github.com/tools/12306/utils"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/login", UserLogin)
	http.HandleFunc("/loginOut", UserLogout)
	http.HandleFunc("/search", SearchTrain)
	http.HandleFunc("/repeat", GetRepeatToken)
	http.HandleFunc("/passenger", GetPassenger)
	http.HandleFunc("/buy", StartBuy)
	http.HandleFunc("/test-reg", Test)
	http.ListenAndServe("127.0.0.1:8000", nil)
}

func UserLogin(w http.ResponseWriter, r *http.Request) {
	QrLogin()
}

func UserLogout(w http.ResponseWriter, r *http.Request) {
	if len(utils.GetCookie().Cookie) == 2 {
		body, _ := ioutil.ReadAll(r.Body)
		utils.AddCookieStr([]string{string(body)})
	}
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
	searchParam := &module.SearchParam{
		TrainDate:   "2022-02-17",
		FromStation: "BJP",
		ToStation:   "TJP",
	}
	trainDatas := GetTrainInfo(searchParam)

	var trainData *module.TrainData
	for _, td := range trainDatas {
		if td.TrainNo == "G171" {
			trainData = td
		}
	}
	fmt.Println(fmt.Sprintf("%+v", trainData))

	CheckUser()
	submitOrderRes := SubmitOrder(trainData, searchParam)
	fmt.Println(fmt.Sprintf("%+v", submitOrderRes))

	passenger := GetPassengers()


	checkOrderRes := CheckOrder(passenger.Data.NormalPassengers[0], passenger.SubmitToken)
	fmt.Println(fmt.Sprintf("%+v", checkOrderRes))
	if !checkOrderRes.Data.SubmitStatus {
		log.Panicln("error", checkOrderRes)
	}

	queueRes := GetQueueCount(passenger.SubmitToken, trainData, searchParam)
	fmt.Println(fmt.Sprintf("%+v", queueRes))
	//AutoBuy(passenger, trainDatas[10])

}

func Test(w http.ResponseWriter, r *http.Request) {
	loginUser.SubmitToken = nil
	GetLoginData()
}
