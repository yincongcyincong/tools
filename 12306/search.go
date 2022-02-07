package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type SearchRes struct {
	HTTPStatus int    `json:"httpstatus"`
	Message    string `json:"message"`
	Status     bool   `json:"status"`
	Data       struct {
		Result []string          `json:"result"`
		Flag   string            `json:"flag"`
		Map    map[string]string `json:"map"`
	} `json:"data"`
}

type SearchData struct {
	SecretStr        string
	TrainNo          string
	FromStationName  string
	FromStation      string
	ToStationName    string
	ToStation        string
	TrainLocation    string
	StationTrainCode string
	LeftTicket       string
	StatTime         string
	ArrivalTime      string
	DistanceTime     string
	Status           string
	SeatInfo         map[string]string
}

type SearchParam struct {
	TrainData   string
	FromStation string
	ToStation   string
}

type PassengerRes struct {
	ValidateMessagesShowId string `json:"validateMessagesShowId"`
	Status                 bool   `json:"status"`
	HTTPStatus             int    `json:"httpstatus"`
	Data                   struct {
		NotifyForGat     string       `json:"notify_for_gat"`
		IsExist          bool         `json:"is_exist"`
		ExMsg            string       `json:"exMsg"`
		TwoIsOpenCLick   []string     `json:"two_isOpenClick"`
		OtherIsOpenClick []string     `json:"other_isOpenClick"`
		NormalPassengers []*Passenger `json:"normal_passengers"`
	} `json:"data"`
	Messages    []string `json:"messages"`
	SubmitToken string
}

type Passenger struct {
	PassengerName       string `json:"passenger_name"`
	SexCode             string `json:"sex_code"`
	SexName             string `json:"sex_name"`
	BornDate            string `json:"born_date"`
	CountryCode         string `json:"country_code"`
	PassengerIdTypeCode string `json:"passenger_id_type_code"`
	PassengerIdTypeName string `json:"passenger_id_type_name"`
	PassengerIdNo       string `json:"passenger_id_no"`
	PassengerType       string `json:"passenger_type"`
	PassengerTypeName   string `json:"passenger_type_name"`
	MobileNo            string `json:"mobile_no"`
	PhoneNo             string `json:"phone_no"`
	Email               string `json:"email"`
	Address             string `json:"address"`
	Postalcode          string `json:"postalcode"`
	FirstLetter         string `json:"first_letter"`
	RecordCount         string `json:"record_count"`
	TotalTimes          string `json:"total_times"`
	IndexId             string `json:"index_id"`
	AllEncStr           string `json:"AllEncStr"`
	IsAdult             string `json:"IsAdult"`
	IsYongThan10        string `json:"IsYongThan10"`
	IsYongThan14        string `json:"IsYongThan14"`
	IsOldThan60         string `json:"IsOldThan60"`
	IfReceive           string `json:"if_receive"`
	IsActive            string `json:"is_active"`
	IsBuyTicket         string `json:"is_buy_ticket"`
	LastTime            string `json:"last_time"`
	PassengerUuid       string `json:"passenger_uuid"`
}

var (
	TokenRe = regexp.MustCompile("var globalRepeatSubmitToken = '(.+)';")
)

func Search(searchParam *SearchParam) []*SearchData {

	var err error
	req, err := http.NewRequest("GET", fmt.Sprintf("https://kyfw.12306.cn/otn/leftTicket/queryA?leftTicketDTO.train_date=%s&leftTicketDTO.from_station=%s&leftTicketDTO.to_station=%s&purpose_codes=ADULT", searchParam.TrainData, searchParam.FromStation, searchParam.ToStation), strings.NewReader(""))
	if err != nil {
		log.Panicln(err)
	}

	req.Header.Set("Cookie", cookieStr())

	resp, err := client.Do(req)
	if err != nil {
		log.Panicln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err)
	}

	searchRes := new(SearchRes)
	err = json.Unmarshal(body, searchRes)
	if err != nil {
		log.Panicln(err)
	}

	if searchRes.HTTPStatus != 200 && searchRes.Status {
		log.Panicln(searchRes.Message)
	}

	searchDatas := make([]*SearchData, len(searchRes.Data.Result))
	for _, res := range searchRes.Data.Result {
		resSlice := strings.Split(res, "|")
		sd := new(SearchData)
		sd.Status = resSlice[1]
		sd.TrainNo = resSlice[3]
		sd.FromStationName = searchRes.Data.Map[resSlice[6]]
		sd.ToStationName = searchRes.Data.Map[resSlice[7]]
		sd.FromStation = resSlice[6]
		sd.ToStation = resSlice[7]

		if resSlice[1] == "预订" {
			sd.SecretStr = resSlice[0]
			sd.LeftTicket = resSlice[29]
			sd.StatTime = resSlice[8]
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

		searchDatas = append(searchDatas, sd)

		fmt.Println(fmt.Sprintf("车次: %s, 状态: %s, 始发车站: %s, 终点站:%s,  %s: %s, 历时：%s",
			sd.TrainNo, sd.Status, sd.FromStationName, sd.ToStationName, sd.StatTime, sd.ArrivalTime, sd.DistanceTime))

	}

	return searchDatas
}

func GetRepeatSubmitToken() string {
	req, err := http.NewRequest("GET", "https://kyfw.12306.cn/otn/confirmPassenger/initDc", strings.NewReader(""))
	if err != nil {
		log.Panicln(err)
	}
	fmt.Println(cookieStr())
	req.Header.Set("Cookie", cookieStr())

	resp, err := client.Do(req)
	if err != nil {
		log.Panicln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err)
	}

	matchRes := TokenRe.FindStringSubmatch(string(body))
	if len(matchRes) > 1 {
		return matchRes[1]
	}

	return ""
}

func GetPassengers(loginRes *LoginRes) *PassengerRes {
	submitToken := GetRepeatSubmitToken()
	if submitToken == "" {
		log.Panicln("submitToken is empty")
	}
	// submit token 需要一样
	if tmpCookie["submit_token"] != "" {
		submitToken = tmpCookie["submit_token"]
	}

	data := make(url.Values)
	data.Set("_json_att", "")
	data.Set("REPEAT_SUBMIT_TOKEN", submitToken)
	res := new(PassengerRes)
	err := request(data, cookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/getPassengerDTOs", res)
	if err != nil {
		log.Panicln(err)
	}

	if res.Status && res.HTTPStatus != 200 {
		log.Panicln(res.Data.ExMsg)
	}

	fmt.Println(submitToken)
	fmt.Println("passengers", res.Data.NormalPassengers)
	fmt.Println(fmt.Sprintf("%+v", res))

	res.SubmitToken = submitToken

	return res

}
