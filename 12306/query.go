package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tools/12306/module"
	"github.com/tools/12306/utils"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	TokenRe           = regexp.MustCompile("var globalRepeatSubmitToken = '(.+)';")
	TicketInfoRe      = regexp.MustCompile("var ticketInfoForPassengerForm=(.+);")
	OrderRequestParam = regexp.MustCompile("var orderRequestDTO=(.+);")
	submitToken       = new(module.SubmitToken)
)

func GetTrainInfo(searchParam *module.SearchParam) []*module.TrainData {

	var err error
	req, err := http.NewRequest("GET", fmt.Sprintf("https://kyfw.12306.cn/otn/leftTicket/queryA?leftTicketDTO.train_date=%s&leftTicketDTO.from_station=%s&leftTicketDTO.to_station=%s&purpose_codes=ADULT", searchParam.TrainDate, searchParam.FromStation, searchParam.ToStation), strings.NewReader(""))
	if err != nil {
		log.Panicln(err)
	}

	req.Header.Set("Cookie", utils.GetCookieStr())

	resp, err := utils.GetClient().Do(req)
	if err != nil {
		log.Panicln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err)
	}

	searchRes := new(module.TrainRes)
	err = json.Unmarshal(body, searchRes)
	if err != nil {
		log.Panicln(err)
	}

	if searchRes.HTTPStatus != 200 && searchRes.Status {
		log.Panicln(searchRes.Message)
	}

	searchDatas := make([]*module.TrainData, len(searchRes.Data.Result))
	for i, res := range searchRes.Data.Result {
		resSlice := strings.Split(res, "|")
		sd := new(module.TrainData)
		sd.Status = resSlice[1]
		sd.TrainNo = resSlice[3]
		sd.FromStationName = searchRes.Data.Map[resSlice[6]]
		sd.ToStationName = searchRes.Data.Map[resSlice[7]]
		sd.FromStation = resSlice[6]
		sd.ToStation = resSlice[7]

		if resSlice[1] == "预订" {
			sd.SecretStr = resSlice[0]
			sd.LeftTicket = resSlice[29]
			sd.StartTime = resSlice[8]
			sd.ArrivalTime = resSlice[9]
			sd.DistanceTime = resSlice[10]

			sd.SeatInfo = make(map[string]string)
			sd.SeatInfo["商务座"] = resSlice[32]
			sd.SeatInfo["一等座"] = resSlice[31]
			sd.SeatInfo["二等座"] = resSlice[30]
			sd.SeatInfo["软卧"] = resSlice[23]
			sd.SeatInfo["硬卧"] = resSlice[28]
			sd.SeatInfo["硬座"] = resSlice[29]
			sd.SeatInfo["无座"] = resSlice[26]
			sd.SeatInfo["动卧"] = resSlice[33]
		}

		searchDatas[i] = sd

		//fmt.Println(fmt.Sprintf("车次: %s, 状态: %s, 始发车站: %s, 终点站:%s,  %s: %s, 历时：%s",
		//	sd.TrainNo, sd.Status, sd.FromStationName, sd.ToStationName, sd.StatTime, sd.ArrivalTime, sd.DistanceTime))

	}
	return searchDatas
}

func GetRepeatSubmitToken() {
	req, err := http.NewRequest("GET", "https://kyfw.12306.cn/otn/confirmPassenger/initDc", strings.NewReader(""))
	if err != nil {
		log.Panicln(err)
	}
	req.Header.Set("Cookie", utils.GetCookieStr())

	resp, err := utils.GetClient().Do(req)
	if err != nil {
		log.Panicln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err)
	}

	matchRes := TokenRe.FindStringSubmatch(string(body))
	if len(matchRes) > 1 {
		submitToken.Token = matchRes[1]
	}

	ticketRes := TicketInfoRe.FindSubmatch(body)
	if len(ticketRes) > 1 {
		ticketRes[1] = bytes.Replace(ticketRes[1], []byte("'"), []byte(`"`), -1)
		err = json.Unmarshal(ticketRes[1], &submitToken.TicketInfo)
		if err != nil {
			log.Panicln(err)
		}
	}

	orderRes := OrderRequestParam.FindSubmatch(body)
	if len(orderRes) > 1 {
		orderRes[1] = bytes.Replace(orderRes[1], []byte("'"), []byte(`"`), -1)
		err = json.Unmarshal(orderRes[1], &submitToken.OrderRequestParam)
		if err != nil {
			log.Panicln(err)
		}
	}

	loginUser.SubmitToken = submitToken
}

func GetPassengers() *module.PassengerRes {

	if loginUser.SubmitToken == nil {
		GetRepeatSubmitToken()
		if submitToken.Token == "" {
			log.Panicln("submitToken is empty")
		}
	}

	data := make(url.Values)
	data.Set("_json_att", "")
	data.Set("REPEAT_SUBMIT_TOKEN", submitToken.Token)
	res := new(module.PassengerRes)
	err := utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/getPassengerDTOs", res, nil)
	if err != nil {
		log.Panicln(err)
	}

	if res.Status && res.HTTPStatus != 200 {
		log.Panicln(res.Data.ExMsg)
	}

	for _, p := range res.Data.NormalPassengers {

		passengerTicketStr := fmt.Sprintf("0,%s,%s,%s,%s,%s,N,%s",
			p.PassengerType, p.PassengerName, p.PassengerIdTypeCode, p.PassengerIdNo, p.MobileNo, p.AllEncStr)
		oldPassengerStr := fmt.Sprintf("%s,%s,%s,%s_",
			p.PassengerName, p.PassengerIdTypeCode, p.PassengerIdNo, p.PassengerType)
		p.PassengerTicketStr = passengerTicketStr
		p.OldPassengerStr = oldPassengerStr
	}

	res.SubmitToken = submitToken
	fmt.Println(submitToken.Token, utils.GetCookieStr())
	fmt.Println(fmt.Sprintf("%+v", res))

	return res

}

func CheckUser() {
	data := make(url.Values)
	data.Set("_json_att", "")
	res := new(module.CheckUserRes)
	fmt.Println(utils.GetCookieStr())
	err := utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/otn/login/checkUser", res, nil)
	if err != nil {
		log.Panicln(err)
	}

	if res.Status && res.HTTPStatus != 200 {
		log.Panicln(res.Messages)
	}

	if !res.Data.Flag {
		log.Panicln("check user:", res.Data.Flag)
	} else {
		fmt.Println("check user success")
	}
	return

}
