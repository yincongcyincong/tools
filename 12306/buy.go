package main

import (
	"fmt"
	"github.com/tools/12306/module"
	"github.com/tools/12306/utils"
	"log"
	"net/url"
	"time"
)

func SubmitOrder(passengerRes *module.PassengerRes, trainData *module.TrainData) {

}

func CheckOrder(passengerRes *module.PassengerRes, trainData *module.TrainData) *module.CheckOrderRes {
	//passengerTicketStr : 座位编号,0,票类型,乘客名,证件类型,证件号,手机号码,保存常用联系人(Y或N)
	//oldPassengersStr: 乘客名,证件类型,证件号,乘客类型

	data := make(url.Values)
	data.Set("bed_level_order_num", "000000000000000000000000000000")
	data.Set("passengerTicketStr", "O,"+passengerRes.Data.NormalPassengers[0].PassengerTicketStr)
	data.Set("oldPassengerStr", passengerRes.Data.NormalPassengers[0].OldPassengerStr)
	data.Set("tour_flag", "dc")
	data.Set("randCode", "")
	data.Set("cancel_flag", "2")
	data.Set("_json_att", "")
	data.Set("whatsSelected", "1")
	data.Set("scene", "nc_login")
	data.Set("REPEAT_SUBMIT_TOKEN", passengerRes.SubmitToken.Token)
	fmt.Println(data)

	checkOrderRes := new(module.CheckOrderRes)
	err := utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/checkOrderInfo", checkOrderRes)
	if err != nil {
		log.Panicln(err)
	}

	return checkOrderRes
}

func Confirm(passengerRes *module.PassengerRes) {
	passengerTicketStr := fmt.Sprintf("0,%s,%s,%s,%s,%s,N,%s",
		passengerRes.Data.NormalPassengers[0].PassengerType, passengerRes.Data.NormalPassengers[0].PassengerName,
		passengerRes.Data.NormalPassengers[0].PassengerIdTypeCode, passengerRes.Data.NormalPassengers[0].PassengerIdNo,
		passengerRes.Data.NormalPassengers[0].MobileNo, passengerRes.Data.NormalPassengers[0].AllEncStr)
	oldPassengerStr := fmt.Sprintf("%s,%s,%s,%s_",
		passengerRes.Data.NormalPassengers[0].PassengerName,
		passengerRes.Data.NormalPassengers[0].PassengerIdTypeCode, passengerRes.Data.NormalPassengers[0].PassengerIdNo,
		passengerRes.Data.NormalPassengers[0].PassengerType)

	data := make(url.Values)
	data.Set("passengerTicketStr", "O,"+passengerTicketStr)
	data.Set("oldPassengerStr", oldPassengerStr)
	data.Set("purpose_codes", passengerRes.SubmitToken.TicketInfo["purpose_codes"].(string))
	data.Set("key_check_isChange", passengerRes.SubmitToken.TicketInfo["key_check_isChange"].(string))
	data.Set("leftTicketStr", passengerRes.SubmitToken.TicketInfo["leftTicketStr"].(string))
	data.Set("train_location", passengerRes.SubmitToken.TicketInfo["train_location"].(string))
	data.Set("seatDetailType", "")
	data.Set("roomType", "00")
	data.Set("dwAll", "N")
	data.Set("whatsSelect", "1")
	data.Set("_json_at", "")
	data.Set("randCode", "")
	data.Set("choose_seats", "")
	data.Set("REPEAT_SUBMIT_TOKEN", passengerRes.SubmitToken.Token)
	fmt.Println(data)

	qrImage := new(module.QrImage)
	err := utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/confirmSingleForQueue", qrImage)
	if err != nil {
		log.Panicln(err)
	}
}

func AutoBuy(passengerRes *module.PassengerRes, trainData *module.TrainData) {
	//passengerTicketStr : 座位编号,0,票类型,乘客名,证件类型,证件号,手机号码,保存常用联系人(Y或N)
	//oldPassengersStr: 乘客名,证件类型,证件号,乘客类型
	var err error

	data := make(url.Values)
	data.Set("bed_level_order_num", "000000000000000000000000000000")
	data.Set("passengerTicketStr", "O,"+passengerRes.Data.NormalPassengers[0].PassengerTicketStr)
	data.Set("oldPassengerStr", passengerRes.Data.NormalPassengers[0].OldPassengerStr)
	data.Set("tour_flag", "dc")
	data.Set("cancel_flag", "2")
	data.Set("purpose_codes", "ADULT")
	data.Set("REPEAT_SUBMIT_TOKEN", passengerRes.SubmitToken.Token)
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
	err = utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/autoSubmitOrderRequest", qrImage)
	if err != nil {
		log.Panicln(err)
	}
}

func GetQueueCount(passengerRes *module.PassengerRes, trainData *module.TrainData, searchParam *module.SearchParam) *module.QueueCountRes {
	var err error
	startTime, err := time.Parse("2006-01-02", searchParam.TrainDate)
	if err != nil {
		log.Panicln(err)
	}

	data := make(url.Values)
	data.Set("stationTrainCode", trainData.StationTrainCode)
	data.Set("train_date", fmt.Sprintf("%s 00:00:00 GMT+0800 (中国标准时间)", startTime.Format("Mon Jan 02 2006 ")))
	data.Set("train_no", trainData.TrainNo)
	data.Set("seatType", "O")
	data.Set("REPEAT_SUBMIT_TOKEN", passengerRes.SubmitToken.Token)
	data.Set("fromStationTelecode", trainData.FromStation)
	data.Set("toStationTelecode", trainData.ToStation)
	data.Set("leftTicket", trainData.LeftTicket)
	data.Set("_json_att", "")
	data.Set("purpose_codes", "00")
	data.Set("train_location", "P2")
	fmt.Println(data)


	trainData.SecretStr, err = url.QueryUnescape(trainData.SecretStr)
	if err != nil {
		log.Panicln(err)
	}
	data.Set("secretStr", trainData.SecretStr)


	queueRes := new(module.QueueCountRes)
	err = utils.Request(data, utils.GetCookieStr(), "https://kyfw.12306.cn/otn/confirmPassenger/getQueueCountAsync", queueRes)
	if err != nil {
		log.Panicln(err)
	}

	return queueRes
}

