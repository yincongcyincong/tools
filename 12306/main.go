package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var loginRes *LoginRes

func main() {

	http.HandleFunc("/login", UserLogin)
	http.HandleFunc("/search", SearchTrain)
	http.HandleFunc("/repeat", GetRepeatToken)
	http.HandleFunc("/passenger", GetPassenger)
	http.HandleFunc("/buy", StartBuy)
	http.HandleFunc("/test-reg", TestReg)
	http.ListenAndServe("127.0.0.1:8000", nil)
}

func UserLogin(w http.ResponseWriter, r *http.Request) {
	loginRes = QrLogin()
	res, _ := json.Marshal(loginRes)
	fmt.Fprint(w, string(res))
}

func SearchTrain(w http.ResponseWriter, r *http.Request) {
	searchParam := &SearchParam{
		TrainData:   "2022-02-17",
		FromStation: "BJP",
		ToStation:   "TJP",
	}
	if len(tmpCookie) == 2 {
		body, _ := ioutil.ReadAll(r.Body)
		a := &http.Response{
			Header: make(map[string][]string),
		}
		a.Header.Set("Set-Cookie", string(body))
		addCookie(a)
	}

	Search(searchParam)
}

func GetRepeatToken(w http.ResponseWriter, r *http.Request) {

	if len(tmpCookie) == 2 {
		body, _ := ioutil.ReadAll(r.Body)
		a := &http.Response{}
		a.Header.Set("Set-Cookie", string(body))
		addCookie(a)
	}
	GetRepeatSubmitToken()
}

func GetPassenger(w http.ResponseWriter, r *http.Request) {

	if len(tmpCookie) == 2 {
		body, _ := ioutil.ReadAll(r.Body)
		a := &http.Response{
			Header: make(map[string][]string),
		}
		a.Header.Set("Set-Cookie", string(body))
		addCookie(a)
	}

	GetPassengers(loginRes)

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

	if len(tmpCookie) == 2 {
		body, _ := ioutil.ReadAll(r.Body)
		a := &http.Response{
			Header: make(map[string][]string),
		}
		a.Header.Set("Set-Cookie", string(body))
		addCookie(a)
	}

	passenger := GetPassengers(loginRes)
	Buy(passenger)

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

func TestReg(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	matchRes := TokenRe.FindStringSubmatch(string(body))
	fmt.Println(matchRes)

	ticketRes := TicketInfoRe.FindStringSubmatch(string(body))
	fmt.Println(ticketRes)

	orderRes := OrderRequestParam.FindStringSubmatch(string(body))
	fmt.Println(orderRes)
}