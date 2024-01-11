package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
)

func check_hash_shit(check_code string) (bool, int) {
	shit0 := "querySelectorAll"
	shit1 := "window.Telegram.WebApp.initDataUnsafe.user.id"
	shit11 := "? 5 : 10"
	shit2 := "window.location.host == 'clicker.joincommunity.xyz' ? 129 : 578"

	if strings.Contains(check_code, shit1) && strings.Contains(check_code, shit11) {
		return true, -1 // 5||10
	}
	if strings.Contains(check_code, shit0) {
		return true, 1
	}
	if strings.Contains(check_code, shit2) {
		return true, 129
	}

	return false, 0
}

func get_needplus(str string) (string, int) {
	splitstr := strings.Split(str, " ")
	last := splitstr[len(splitstr)-1]
	plus, err := strconv.Atoi(last)
	if err != nil {
		ErrorLogger.Println("Error on get_needplus, err =", err)
	}
	first_str := strings.Join(splitstr[:len(splitstr)-1], " ")
	return first_str, plus
}

func hash_resolve(resp []string) int {
	var need_plus int
	encodedString := strings.Join(resp, "")
	if len(encodedString) < 3 {
		return -1
	}

	decodedBytes, _ := base64.StdEncoding.DecodeString(encodedString)
	codejs_str := string(decodedBytes)

	if len(resp) > 1 {
		codejs_str, need_plus = get_needplus(codejs_str)
	}

	findshit, shit := check_hash_shit(codejs_str)
	if findshit {
		return shit + need_plus
	}
	vm := otto.New()
	result, _ := vm.Run(codejs_str)
	resultint, _ := result.ToInteger()
	return int(resultint) + need_plus

}

func get_divide(coef, divide int) int {
	divided := coef / divide
	if divided < 1 {
		divided = 1
	}
	return divided
}

func (Notcoin *Notcoin) get_count_click() int { // in hand 40/sec, turbo = hand*3
	var coinscount int
	var minus int
	if Notcoin.Turbo {
		minus = get_randomint(132, 311, 1)
		coinscount = Notcoin.LimitCoins/4 - minus
		if Notcoin.Timestart_turbo == 0 {
			Notcoin.Timestart_turbo = time.Now().Unix()
		} else if Notcoin.Timestart_turbo+11 <= time.Now().Unix() {
			Notcoin.Turbo = false
			Notcoin.Timestart_turbo = 0
		}
	} else {
		minus = get_divide(Notcoin.Coefficient, 3)
		min := 100*Notcoin.Coefficient/get_divide(Notcoin.Coefficient, 3) + get_randomint(2, 47, 1)
		if minus <= 2 {
			coinscount = min
		} else {
			coinscount = Notcoin.LastAvailableCoins / minus
		}

		if coinscount <= min && Notcoin.LastAvailableCoins > min {
			coinscount = min
		} else if coinscount < min {
			coinscount = Notcoin.LastAvailableCoins
		}
	}

	return coinscount
}

func get_randomint(min, max, coef int) int {
	rand.Seed(time.Now().UnixNano())
	randomInt := rand.Intn(max-min+1) + min
	return randomInt * coef
}

func get_Turbo(resp string) bool {
	re := regexp.MustCompile(`"Turbo":(.*?)}`)
	match := re.FindStringSubmatch(resp)
	if len(match) > 1 {
		if strings.Contains(strings.ToLower(match[1]), "true") {
			return true
		}
	}
	return false
}

func parse_respclick(content []byte) *Click_resp {
	var response Click_resp
	json.Unmarshal(content, &response)
	return &response
}

func (Notcoin *Notcoin) click() {
	var data string
	urlstr := "https://clicker-api.joincommunity.xyz/clicker/core/click"
	count := Notcoin.get_count_click()
	webAppData := Notcoin.TGWebAppData
	if Notcoin.Turbo {
		if Notcoin.Hash != -1 {
			data = fmt.Sprintf(`{"count":%d, "hash":%d, "Turbo": true, "webAppData":"%v"}`, count, Notcoin.Hash, webAppData)
		} else {
			data = fmt.Sprintf(`{"count":%d, "Turbo": true, "webAppData":"%v"}`, count, webAppData)
		}
	} else {
		if Notcoin.Hash != -1 {
			data = fmt.Sprintf(`{"count":%d, "hash":%d, "webAppData":"%v"}`, count, Notcoin.Hash, webAppData)
		} else {
			data = fmt.Sprintf(`{"count":%d,"webAppData":"%v"}`, count, webAppData)
		}
	}

	resp := Notcoin.Ses.Postreq(urlstr, data)
	parsed_resp := parse_respclick(resp.body)

	if parsed_resp.Ok {
		Notcoin.Count_400 = 0
		Notcoin.LimitCoins = parsed_resp.Data[0].LimitCoins
		Notcoin.Hash = hash_resolve(parsed_resp.Data[0].Hash)
		Notcoin.BalanceCoins = parsed_resp.Data[0].BalanceCoins
		Notcoin.Coefficient = parsed_resp.Data[0].MultipleClicks
		Notcoin.Turbo_boost_count = parsed_resp.Data[0].TurboTimes
		Notcoin.LastAvailableCoins = parsed_resp.Data[0].LastAvailableCoins

	} else {
		if resp.status == 400 {
			Notcoin.Count_400++
		}
		Notcoin.Hash = -1
	}

	if Notcoin.LastAvailableCoins < Notcoin.LimitCoins/2 &&
		Notcoin.Turbo_boost_count > 0 &&
		Notcoin.Count_400 == 0 &&
		Notcoin.Hash != -1 &&
		!Notcoin.Turbo {
		Notcoin.Turbo_activate()
	}

	if parsed_resp.Ok {
		SuccessLogger.Printf("[%d] clicked %d times, status = %d, next Turbo = %v\n", Notcoin.UserId, count, resp.status, Notcoin.Turbo)
	} else {
		WarningLogger.Printf("[%d] not success clicked %d times, status = %d\n", Notcoin.UserId, count, resp.status)
	}
}

func (not *Notcoin) Turbo_activate() {
	var url_activate_turbo string = "https://clicker-api.joincommunity.xyz/clicker/core/active-turbo"
	var parsed_resp Active_turbo_resp
	var ok bool

	resp := not.Ses.Postreq(url_activate_turbo, "{}")
	json.Unmarshal(resp.body, &parsed_resp)
	ok = parsed_resp.Ok
	if !ok {
		return
	}

	not.Turbo = true
	not.Turbo_boost_count--
	not.Timestart_turbo = time.Now().Unix()
	SuccessLogger.Printf("[%d] Activated turbo", not.UserId)
}
