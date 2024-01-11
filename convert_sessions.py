from telethon.sync import TelegramClient
from telethon.sessions import StringSession
from os.path import exists
from os import listdir, replace, mkdir
from json import loads


def_env ="""enable_daily_boosters = 1 # 1 = enable
# update to max lvl in the shop:
max_tapbot = 0 # id= 18
max_multitab = 8 # id= 3
max_recharging = 3 # id= 2
max_energy_limit = 9 # id= 1"""

if __name__ == "__main__":
    if not exists(".env"):
        with open(".env", "w") as f:
            f.write(def_env)
    if not exists("./telethon_sessions"):
        mkdir("telethon_sessions")
        
    try:
        with open("./accounts.json", "r") as f:
            accounts = loads(f.read())
    except Exception as e:
        print(f"accounts.json is not found in folder or it broken, {e}")
        exit(1)

    for file in listdir("./telethon_sessions//"):
        clear_name = file.split(".")[0]
        path_file = f"./telethon_sessions/{file}"
        
        account = accounts.get(clear_name, False)
        if account is False:
            print(f"ERROR {clear_name} NOT FOUND in accounts.json")
            continue
        else:
            api_id:int = account.get("api_id", 0)
            api_hash:str = account.get("api_hash", "")
            proxy:str = account.get("proxy", "")
            if api_id == 0 or api_hash == "":
                print(f"ERROR {file} api_id or api_hash NOT FOUND in accounts.json")
                continue
            if len(proxy) > 1:
                if "http://" in proxy:
                    proxy = {"socks4":proxy, "socks4":proxy}
                elif "socks5" in proxy:
                    proxy = {"socks5":proxy, "socks5":proxy}
                elif "socks4" in proxy:
                    proxy = {"http":proxy, "https":proxy}
                else:
                    proxy = {"http":f"http://{proxy}", "https":f"http://{proxy}",}
            else:
                proxy = None
        try:
            print(f"Current = {clear_name}")
            if not ("str_sessions"):
                mkdir("str_sessions")
            with TelegramClient(session=path_file, api_id=api_id, api_hash=api_hash, proxy=proxy) as client:
                string = StringSession.save(client.session)
                new_path_file = f"./str_sessions/{clear_name}.strses"
                with open(new_path_file, "w") as f:
                    f.write(string)
            print(f"SUCCESSFULY convert {clear_name}")
        except Exception as e:
            print(f"Error convert {clear_name}, moved to errors folder")
            if not exists("error_sessions"):
                mkdir("error_sessions")
            replace(path_file, f"./error_sessions//{file}")