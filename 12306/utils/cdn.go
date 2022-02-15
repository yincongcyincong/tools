package utils

import (
	"github.com/cihub/seelog"
	"github.com/tools/12306/conf"
	"sync"
)

type AvailableCDN struct {
	cdns     []string
	currency int
	lock     sync.Mutex
	wg       sync.WaitGroup
}

var availableCDN = &AvailableCDN{
	cdns:     make([]string, 0),
	currency: 10,
	lock:     sync.Mutex{},
	wg:       sync.WaitGroup{},
}

func InitAvailableCDN() {
	num := (len(conf.CDNs) / availableCDN.currency) + 1
	for i := 0; i < num; i++ {
		availableCDN.wg.Add(1)
		var tmpCDNs []string
		if i != num-1 {
			tmpCDNs = conf.CDNs[i*availableCDN.currency : (i+1)*availableCDN.currency]
		} else {
			tmpCDNs = conf.CDNs[i*availableCDN.currency:]
		}

		go func(cdns []string) {
			defer availableCDN.wg.Done()
			for _, cdn := range cdns {
				_, err := RequestGetWithCDN(GetCookieStr(), "https://kyfw.12306.cn/otn/dynamicJs/omseuuq", nil, cdn)
				if err != nil {
					seelog.Tracef("%s query fail", cdn)
					continue
				}

				availableCDN.lock.Lock()
				availableCDN.cdns = append(availableCDN.cdns, cdn)
				availableCDN.lock.Unlock()

			}

		}(tmpCDNs)
	}

	availableCDN.wg.Wait()
	seelog.Infof("available cdn num: %d", len(availableCDN.cdns))

}
