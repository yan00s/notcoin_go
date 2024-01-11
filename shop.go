package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

var (
	url_store         string = "https://clicker-api.joincommunity.xyz/clicker/store/merged"
	url_buy           string = "https://clicker-api.joincommunity.xyz/clicker/store/buy/"
	url_task          string = "https://clicker-api.joincommunity.xyz/clicker/task/"
	url_combine_compl string = "https://clicker-api.joincommunity.xyz/clicker/task/combine-completed"
)

func (not *Notcoin) Active_task(id_task int) bool {
	var task_name string
	switch id_task {
	case 2:
		task_name = "FullEnergy"
	case 3:
		task_name = "Turbo"
	}

	resp := not.Ses.Postreq(fmt.Sprintf("%v%d", url_task, id_task), "{}")
	parsed_resp := parse_respclick(resp.body)

	if parsed_resp.Ok {
		SuccessLogger.Printf("[%v] '%v' task activated %v\n", not.UserId, task_name, task_name)
		switch task_name {
		case "FullEnergy":
			not.Fullenergy_boost--
			not.LastAvailableCoins = not.LimitCoins
		case "Turbo":
			not.Turbo_boost_count++
		}
		return true
	} else {
		WarningLogger.Printf("[%v] '%v' task activated is not ok, response: %v\n", not.UserId, task_name, resp.String())
		return false
	}
}

func (not *Notcoin) Buy_item(item int) (bool, string) {
	var ok bool
	var str string
	var resp_parsed Buy_item_resp
	resp := not.Ses.Postreq(fmt.Sprintf("%v%d", url_buy, item), "{}")
	json.Unmarshal(resp.body, &resp_parsed)

	ok = resp_parsed.OK
	if ok {
		not.BalanceCoins = resp_parsed.Data.BalanceCoins
	} else {
		str = resp.String()
	}
	return ok, str
}

func (not *Notcoin) Update_shop() {
	var parsed_resp Store_resp
	var select_item bool
	var item_count int
	var respstr string
	var isok bool
	var max_tapbot, _ = strconv.Atoi(os.Getenv("max_tapbot"))             //# id= 18
	var max_multitab, _ = strconv.Atoi(os.Getenv("max_multitab"))         //# id= 3
	var max_recharging, _ = strconv.Atoi(os.Getenv("max_recharging"))     //# id= 2
	var max_energy_limit, _ = strconv.Atoi(os.Getenv("max_energy_limit")) //# id= 1

	shop_resp := not.Ses.Getreq(url_store)
	json.Unmarshal(shop_resp.body, &parsed_resp)
	if !parsed_resp.OK {
		WarningLogger.Printf("[%v] update shop items is not ok, response: %v\n", not.UserId, shop_resp.String())
		return
	}
	for _, item := range parsed_resp.Data {
		switch item.ID {
		case 1:
			select_item = true
			item_count = max_energy_limit

		case 2:
			select_item = true
			item_count = max_recharging

		case 3:
			select_item = true
			item_count = max_multitab
		case 18:
			select_item = true
			item_count = max_tapbot
		}
		if !select_item {
			continue
		}

		if item_count > item.Count &&
			not.BalanceCoins >= item.Price &&
			item.Status == "active" {
			isok, respstr = not.Buy_item(item.ID)
			if isok {
				SuccessLogger.Printf("[%v] buy item '%v'\n", not.UserId, item.Name)
			} else {
				WarningLogger.Printf("[%v] item '%v' buying is not success: %v\n", not.UserId, item.Name, respstr)
			}
			time.Sleep(2 * time.Second)
		}
		select_item = false
	}

	SuccessLogger.Printf("[%v] updated shop items\n", not.UserId)
}

func (not *Notcoin) Update_boosters() {
	var resp_parsed Task_completed_resp
	var energy_count int = 3
	var turbo_count int = 3

	resp := not.Ses.Getreq(url_combine_compl)
	json.Unmarshal(resp.body, &resp_parsed)
	if !resp_parsed.Ok == true {
		WarningLogger.Printf("[%v] update boosters is not ok, response: %v\n", not.UserId, resp.String())
		return
	}

	for _, boost := range resp_parsed.Data {
		switch boost.TaskId {
		case 2: // fullenergy
			if boost.Task.Status == "active" {
				energy_count--
			}
		case 3: // turboboost
			if boost.Task.Status == "active" {
				turbo_count--
			}
		}
	}

	not.Fullenergy_boost = energy_count
	for i := 0; i < turbo_count; i++ {
		not.Active_task(3)
		time.Sleep(3 * time.Second)
	}
}
