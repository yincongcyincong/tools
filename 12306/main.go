package main

import (
	"fmt"
	"github.com/tools/12306/module"
	"github.com/tools/12306/utils"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {

	http.HandleFunc("/login", UserLogin)
	http.HandleFunc("/loginOut", UserLogout)
	http.HandleFunc("/search", SearchTrain)
	http.HandleFunc("/repeat", GetRepeatToken)
	http.HandleFunc("/passenger", GetPassenger)
	http.HandleFunc("/buy", StartBuy)
	http.HandleFunc("/check-order", CheckOrderReq)
	http.HandleFunc("/test-reg", Test)
	http.HandleFunc("/re-login", ReLogin)
	http.ListenAndServe("127.0.0.1:8000", nil)
}

func UserLogin(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	utils.AddCookieStr([]string{string(body)})
	QrLogin()
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

	GetTrainInfo(searchParam)
}

func GetRepeatToken(w http.ResponseWriter, r *http.Request) {

	GetRepeatSubmitToken()
}

func GetPassenger(w http.ResponseWriter, r *http.Request) {


	if loginUser.SubmitToken == nil || loginUser.SubmitToken.Token == "" {
		GetRepeatSubmitToken()
		if loginUser.SubmitToken.Token == "" {
			log.Panicln("submitToken is empty")
		}
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

	if loginUser.BuyStatus == 1 {
		fmt.Fprint(w, "is buying")
		return
	}
	loginUser.BuyStatus = 1
	defer func() {
		loginUser.BuyStatus = 0
		loginUser.SubmitToken = new(module.SubmitToken)
	}()

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

	if loginUser.SubmitToken == nil || loginUser.SubmitToken.Token == "" {
		GetRepeatSubmitToken()
		if loginUser.SubmitToken.Token == "" {
			log.Panicln("submitToken is empty")
		}
	}

	passenger := GetPassengers()

	loginUser.Passenger = passenger.Data.NormalPassengers[0]
	loginUser.TrainData = trainData

	checkOrderRes := CheckOrder(loginUser.Passenger, loginUser.SubmitToken)
	fmt.Println(fmt.Sprintf("%+v", checkOrderRes))
	if !checkOrderRes.Data.SubmitStatus {
		log.Panicln("error", checkOrderRes)
	}


	queueRes := GetQueueCount(loginUser.SubmitToken, loginUser.TrainData, searchParam)
	fmt.Println(fmt.Sprintf("%+v", queueRes))

	confirmRes := ConfirmQueue(loginUser.Passenger, loginUser.SubmitToken)
	fmt.Println(fmt.Sprintf("%+v", confirmRes))

	orderRes := OrderWait(loginUser.SubmitToken)
	for i := 0; i < 100; i++ {
		fmt.Println(fmt.Sprintf("%+v", orderRes))
		if orderRes.Data.OrderId != "" {
			break
		}

		time.Sleep(3 * time.Second)
	}

	OrderResult(loginUser.SubmitToken, orderRes.Data.OrderId)


}

func Test(w http.ResponseWriter, r *http.Request) {

}

func CheckOrderReq(w http.ResponseWriter, r *http.Request) {
	// 高频率请求会直接失败，也可能是两次登陆导致的，长时间不请求cookie会失效， todo 怎么搞到能用的cookie
	searchParam := &module.SearchParam{
		TrainDate:   "2022-02-17",
		FromStation: "BJP",
		ToStation:   "TJP",
	}
	checkOrderRes := CheckOrder(loginUser.Passenger, loginUser.SubmitToken)
	fmt.Println(fmt.Sprintf("%+v", checkOrderRes))
	if !checkOrderRes.Data.SubmitStatus {
		log.Panicln("error", checkOrderRes)
	}

	queueRes := GetQueueCount(loginUser.SubmitToken, loginUser.TrainData, searchParam)
	fmt.Println(fmt.Sprintf("%+v", queueRes))
}

func ReLogin(w http.ResponseWriter, r *http.Request) {
	loginUser.SubmitToken = nil
	GetLoginData()
}
