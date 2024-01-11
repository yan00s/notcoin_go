package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
)

func get_userid_acesshash(api *tg.Client, ctx context.Context, username string) (int64, int64, error) {
	ress, err := api.ContactsResolveUsername(ctx, username)
	if err != nil {
		return 0, 0, err
	}
	user := ress.Users[0]
	user1, isok := user.AsNotEmpty()
	if !isok {
		return 0, 0, fmt.Errorf("Anything err in user.AsNotEmpty")
	}
	userid := user1.GetID()
	accesshash, isok := user1.GetAccessHash()
	if !isok {
		return 0, 0, fmt.Errorf("Anything err in user.AsNotEmpty")
	}
	return userid, accesshash, nil
}

func get_peer(id, hash int64) *tg.InputPeerUser {
	return &tg.InputPeerUser{
		UserID:     id,
		AccessHash: hash,
	}
}

func get_bot(id, hash int64) *tg.InputUser {
	return &tg.InputUser{
		UserID:     id,
		AccessHash: hash,
	}
}

func read_str_ses(path string) string {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read file: %s", err)
	}
	return string(content)
}

func (Notcoin *Notcoin) get_appdata() (string, error) {
	var result_url string
	var resolver dcs.Resolver
	var storage = new(session.StorageMemory)
	var loader = session.Loader{Storage: storage}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	str_ses := read_str_ses(Notcoin.Path_file)
	data, err := session.TelethonSession(str_ses)
	if err != nil {
		ErrorLogger.Printf("Err on decoded telethon session err = %v\n", err)
		return "", err
	}

	if err := loader.Save(ctx, data); err != nil {
		ErrorLogger.Printf("Err on save decoded telethon session err = %v\n", err)
		return "", err
	}

	if len(Notcoin.Proxy) > 1 {
		dial, err := proxyDialer(Notcoin.Proxy)
		if err != nil {
			return "", err
		}
		resolver = dcs.Plain(dcs.PlainOptions{Dial: dial.DialContext})
	} else {
	}

	options := telegram.Options{SessionStorage: storage, Resolver: resolver}
	client := telegram.NewClient(Notcoin.TG_appID, Notcoin.TG_appHash, options)
	if err := client.Run(ctx, func(ctx context.Context) error {
		api := client.API()

		userid, accesshash, err := get_userid_acesshash(api, ctx, "notcoin_bot")
		if err != nil {
			return fmt.Errorf("in get_userid_acesshash err: %v", err.Error())
		}
		request := &tg.MessagesRequestWebViewRequest{
			Peer:        get_peer(userid, accesshash),
			Bot:         get_bot(userid, accesshash),
			Platform:    "android",
			FromBotMenu: false,
			URL:         "https://clicker.joincommunity.xyz/clicker",
		}
		res, err := api.MessagesRequestWebView(ctx, request)
		result_url = res.GetURL()
		return nil
	}); err != nil {
		return "", fmt.Errorf("in clientRun global err: %v", err.Error())
	}
	return result_url, nil
}
