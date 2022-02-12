package main

import (
	"fmt"
	"github.com/cihub/seelog"
	"github.com/tools/12306/conf"
	"github.com/tools/12306/module"
	"github.com/tools/12306/utils"
	"math"
	"strings"
	"time"
)

func CommandStart() {
	qrImage, err := CreateImage()
	if err != nil {
		seelog.Errorf("创建二维码失败:%v", err)
		return
	}
	qrImage.Image = ""

	err = QrLogin(qrImage)
	if err != nil {
		seelog.Errorf("登陆失败:%v", err)
		return
	}
	searchParam := new(module.SearchParam)
	var trainStr, seatStr, passengerStr string
	for i := 1; i < math.MaxInt64; i++ {
		getUserInfo(searchParam, &trainStr, &seatStr, &passengerStr)
		if trainStr != "" && seatStr != "" && passengerStr != "" {
			break
		}

		err = GetLoginData()
		if err != nil {
			seelog.Errorf("自动登陆失败：%v", err)
		}

		time.Sleep(5 * time.Second)
	}

	// 开始轮训买票
	trainMap := utils.GetBoolMap(strings.Split(trainStr, ","))
	passengerMap := utils.GetBoolMap(strings.Split(passengerStr, ","))
	seatSlice := strings.Split(seatStr, ",")

Search:
	var trainData *module.TrainData
	for i := 0; i < math.MaxInt64; i++ {
		time.Sleep(2 * time.Second)
		searchParam.SeatType = ""
		trainData = new(module.TrainData)

		// 一分钟进行一次自动登陆
		if i%30 == 0 {
			err = GetLoginData()
			if err != nil {
				seelog.Errorf("自动登陆失败：%v", err)
			}
		}

		trains, err := GetTrainInfo(searchParam)
		if err != nil {
			seelog.Errorf("查询车站失败:%v", err)
			continue
		}

		for _, t := range trains {
			// 在选中的，但是不在小黑屋里面
			if utils.InBlackList(t.TrainNo) {
				fmt.Println(t.TrainNo, "在小黑屋，需等待60s")
				continue
			}


			if trainMap[t.TrainNo] {
				fmt.Println(trainMap, t.TrainNo, seatSlice, t.SeatInfo)
				for _, s := range seatSlice {
					if t.SeatInfo[s] != "" && t.SeatInfo[s] != "无" {
						trainData = t
						searchParam.SeatType = conf.OrderSeatType[s]
						break
					}
					seelog.Infof("%s %s 数量: %s", t.TrainNo, s, t.SeatInfo[s])
				}

				if trainData != nil && searchParam.SeatType != "" {
					break
				}
			}
		}
		if trainData == nil || searchParam.SeatType == "" {
			fmt.Println("暂无车票可以购买")
			continue
		} else {
			break
		}

	}

	fmt.Println("开始购买", trainData.TrainNo)
	err = startOrder(searchParam, trainData, passengerMap)
	if err != nil {
		utils.AddBlackList(trainData.TrainNo, time.Now().Unix())
		goto Search
	}

}

func getUserInfo(searchParam *module.SearchParam, trainStr, seatStr, passengerStr *string) {
	fmt.Println("请输入日期 起始站 到达站: ")
	fmt.Scanf("%s %s %s", &searchParam.TrainDate, &searchParam.FromStationName, &searchParam.ToStationName)
	searchParam.FromStation = conf.Station[searchParam.FromStationName]
	searchParam.ToStation = conf.Station[searchParam.ToStationName]

	trains, err := GetTrainInfo(searchParam)
	if err != nil {
		seelog.Errorf("查询车站失败:%v", err)
		return
	}
	for _, t := range trains {
		fmt.Println(fmt.Sprintf("车次: %s, 状态: %s, 始发车站: %s, 终点站:%s,  %s: %s, 历时：%s",
			t.TrainNo, t.Status, t.FromStationName, t.ToStationName, t.StartTime, t.ArrivalTime, t.DistanceTime))

	}

	fmt.Println("请输入车次(多个,分割):")
	fmt.Scanf("%s", trainStr)

	fmt.Println("请输入座位类型(多个,分割，二等座，硬座，卧铺等):")
	fmt.Scanf("%s", seatStr)

	submitToken, err := GetRepeatSubmitToken()
	if err != nil {
		seelog.Errorf("获取提交数据失败:%v", err)
		return
	}
	passengers, err := GetPassengers(submitToken)
	if err != nil {
		seelog.Errorf("获取用户失败:%v", err)
		return
	}
	for _, p := range passengers.Data.NormalPassengers {
		fmt.Println(fmt.Sprintf("乘客姓名：%s", p.PassengerName))
	}

	if *passengerStr == "" {
		fmt.Println("请输入乘客姓名(多个,分割): ")
		fmt.Scanf("%s", passengerStr)
	}

	return
}

func startOrder(searchParam *module.SearchParam, trainData *module.TrainData, passengerMap map[string]bool) error {
	err := CheckUser()
	if err != nil {
		seelog.Errorf("检查用户状态失败：%v", err)
		return err
	}

	err = SubmitOrder(trainData, searchParam)
	if err != nil {
		seelog.Errorf("提交订单失败：%v", err)
		return err
	}

	submitToken, err := GetRepeatSubmitToken()
	if err != nil {
		seelog.Errorf("获取提交数据失败：%v", err)
		return err
	}

	passengers, err := GetPassengers(submitToken)
	if err != nil {
		seelog.Errorf("获取乘客失败：%v", err)
		return err
	}
	buyPassengers := make([]*module.Passenger, 0)
	for _, p := range passengers.Data.NormalPassengers {
		if passengerMap[p.PassengerName] {
			buyPassengers = append(buyPassengers, p)
		}
	}

	err = CheckOrder(buyPassengers, submitToken, searchParam)
	if err != nil {
		seelog.Errorf("检查订单失败：%v", err)
		return err
	}

	err = GetQueueCount(submitToken, searchParam)
	if err != nil {
		seelog.Errorf("获取排队数失败：%v", err)
		return err
	}

	err = ConfirmQueue(buyPassengers, submitToken, searchParam)
	if err != nil {
		seelog.Errorf("提交订单失败：%v", err)
		return err
	}

	var orderWaitRes *module.OrderWaitRes
	for i := 0; i < 10; i++ {
		orderWaitRes, err = OrderWait(submitToken)
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}
		if orderWaitRes.Data.OrderId != "" {
			break
		}
	}

	err = OrderResult(submitToken, orderWaitRes.Data.OrderId)
	if err != nil {
		seelog.Errorf("获取订单状态失败：%v", err)
		return err
	}

	seelog.Info("购买成功，订单号：%s", orderWaitRes.Data.OrderId)
	return nil
}
