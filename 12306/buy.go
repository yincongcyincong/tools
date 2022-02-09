package main

import (
	"fmt"
	"github.com/tools/12306/module"
	"github.com/tools/12306/utils"
	"log"
	"net/url"
	"strings"
	"time"
)

func SubmitOrder(trainData *module.TrainData, searchParam *module.SearchParam) *module.SubmitOrderRes {
	var err error
	data := make(url.Values)
	data.Set("train_date", searchParam.TrainDate)
	data.Set("back_train_date", time.Now().Format("2006-01-02"))
	data.Set("tour_flag", "dc")
	data.Set("purpose_codes", "ADULT")
	data.Set("query_from_station_name", "北京")
	data.Set("query_to_station_name", "天津")
	secretStr, err := url.QueryUnescape(trainData.SecretStr)
	if err != nil {
		log.Panicln(err)
	}
	data.Set("secretStr", secretStr)
	fmt.Println(data)

	checkOrderRes := new(module.SubmitOrderRes)
	err = utils.Request(data.Encode(), utils.GetCookieStr(), "https://kyfw.12306.cn/otn/leftTicket/submitOrderRequest", checkOrderRes, map[string]string{"Referer": "https://kyfw.12306.cn/otn/leftTicket/init?linktypeid=dc"})
	if err != nil {
		log.Panicln(err)
	}

	return checkOrderRes
}

func CheckOrder(passenger *module.Passenger, submitToken *module.SubmitToken) *module.CheckOrderRes {
	//passengerTicketStr : 座位编号,0,票类型,乘客名,证件类型,证件号,手机号码,保存常用联系人(Y或N)
	//oldPassengersStr: 乘客名,证件类型,证件号,乘客类型

	data := fmt.Sprintf("bed_level_order_num=000000000000000000000000000000&passengerTicketStr=%s&oldPassengerStr=%s&tour_flag=dc&randCode=&sessionId=&sig=&cancel_flag=2&_json_att&whatsSelecte=1&scene=nc_login&REPEAT_SUBMIT_TOKEN=%s",
		strings.Replace(url.QueryEscape("O,"+passenger.PassengerTicketStr), "%2A", "*", -1), strings.Replace(url.QueryEscape(passenger.OldPassengerStr), "%2A", "*", -1), submitToken.Token)

	checkOrderRes := new(module.CheckOrderRes)
	fmt.Println(utils.GetCookieStr())
	fmt.Println(data)
	err := utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/checkOrderInfo", checkOrderRes, map[string]string{"Referer": "https://kyfw.12306.cn/otn/confirmPassenger/initDc"})
	if err != nil {
		log.Panicln(err)
	}

	return checkOrderRes
}

func ConfirmQueue(passenger *module.Passenger, submitToken *module.SubmitToken) {

	data := make(url.Values)
	data.Set("passengerTicketStr", "O,"+passenger.PassengerTicketStr)
	data.Set("oldPassengerStr", passenger.OldPassengerStr)
	data.Set("purpose_codes", submitToken.TicketInfo["purpose_codes"].(string))
	data.Set("key_check_isChange", submitToken.TicketInfo["key_check_isChange"].(string))
	data.Set("leftTicketStr", submitToken.TicketInfo["leftTicketStr"].(string))
	data.Set("train_location", submitToken.TicketInfo["train_location"].(string))
	data.Set("seatDetailType", "000")
	data.Set("roomType", "00")
	data.Set("dwAll", "N")
	data.Set("whatsSelect", "1")
	data.Set("_json_at", "")
	data.Set("randCode", "")
	data.Set("choose_seats", "1D")
	data.Set("REPEAT_SUBMIT_TOKEN", submitToken.Token)
	fmt.Println(data)

	qrImage := new(module.QrImage)
	err := utils.Request(data.Encode(), utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/confirmSingleForQueue", qrImage, map[string]string{"Referer": "https://kyfw.12306.cn/otn/confirmPassenger/initDc"})
	if err != nil {
		log.Panicln(err)
	}
}

func AutoBuy(passenger *module.Passenger, trainData *module.TrainData, submitToken *module.SubmitToken) {
	//passengerTicketStr : 座位编号,0,票类型,乘客名,证件类型,证件号,手机号码,保存常用联系人(Y或N)
	//oldPassengersStr: 乘客名,证件类型,证件号,乘客类型
	var err error

	data := make(url.Values)
	data.Set("bed_level_order_num", "000000000000000000000000000000")
	data.Set("passengerTicketStr", "O,"+passenger.PassengerTicketStr)
	data.Set("oldPassengerStr", passenger.OldPassengerStr)
	data.Set("tour_flag", "dc")
	data.Set("cancel_flag", "2")
	data.Set("purpose_codes", "ADULT")
	data.Set("REPEAT_SUBMIT_TOKEN", submitToken.Token)
	data.Set("query_from_station_name", trainData.FromStation)
	data.Set("query_to_station_name", trainData.ToStation)
	data.Set("train_date", trainData.StartTime)

	trainData.SecretStr, err = url.QueryUnescape(trainData.SecretStr)
	if err != nil {
		log.Panicln(err)
	}
	data.Set("secretStr", trainData.SecretStr)

	fmt.Println(data)

	qrImage := new(module.AutoBuyRes)
	err = utils.Request(data.Encode(), utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/autoSubmitOrderRequest", qrImage, map[string]string{"Referer": "https://kyfw.12306.cn/otn/confirmPassenger/initDc"})
	if err != nil {
		log.Panicln(err)
	}
}

func GetQueueCount(submitToken *module.SubmitToken, trainData *module.TrainData, searchParam *module.SearchParam) *module.QueueCountRes {
	var err error
	startTime, err := time.Parse("2006-01-02", searchParam.TrainDate)
	if err != nil {
		log.Panicln(err)
	}

	data := fmt.Sprintf("train_location=%s&purpose_codes=%s&_json_att=&leftTicket=%s&toStationTelecode=%s&fromStationTelecode=%s&REPEAT_SUBMIT_TOKEN=%s&seatType=O&train_no=%s&stationTrainCode=%s&train_date=%s",
		submitToken.TicketInfo["train_location"].(string), submitToken.TicketInfo["purpose_codes"].(string), submitToken.TicketInfo["leftTicketStr"].(string),
		submitToken.TicketInfo["queryLeftTicketRequestDTO"].(map[string]interface{})["to_station"].(string), submitToken.TicketInfo["queryLeftTicketRequestDTO"].(map[string]interface{})["from_station"].(string),
		submitToken.TicketInfo["queryLeftTicketRequestDTO"].(map[string]interface{})["from_station"].(string), submitToken.TicketInfo["queryLeftTicketRequestDTO"].(map[string]interface{})["train_no"].(string),
		submitToken.TicketInfo["queryLeftTicketRequestDTO"].(map[string]interface{})["station_train_code"].(string),
		strings.Replace(
			strings.Replace(
				url.QueryEscape(fmt.Sprintf("%s  00:00:00 GMT+0800 (中国标准时间)", startTime.Format("Mon Jan 02 2006"))), "%28", "(", -1),
			"%29", ")", -1),
	)

	data = data + "&secretStr=" + trainData.SecretStr
	fmt.Println(data)

	queueRes := new(module.QueueCountRes)
	err = utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/getQueueCountAsync", queueRes, map[string]string{"Referer": "https://kyfw.12306.cn/otn/confirmPassenger/initDc"})
	if err != nil {
		log.Panicln(err)
	}

	return queueRes
}
