from telethon.sync import TelegramClient
from telethon.sessions import StringSession
from os.path import exists
from os import listdir, replace, mkdir
from json import loads, load, dump, dumps



def by_one():
    file_name = ""
    try:
        data_for_add = {}
        path_file = input("Input full path to .session file (or new file name): ").strip()
        file_name = path_file.split("/")[-1].split("\\")[-1].strip()
        clear_name = file_name.split(".")[0]
        file_name = f"{clear_name}.session"
        print("Input need data from https://my.telegram.org/auth")
        api_id = int(input("Input api_id: ").strip())
        api_hash = input("Input api_hash: ").strip()
        proxy = input("Input proxy: ").strip()
        
        if not ("str_sessions"):
            mkdir("str_sessions")
        with TelegramClient(session=path_file, api_id=api_id, api_hash=api_hash, proxy=proxy) as client:
            string = StringSession.save(client.session)
            new_path_file = f"./str_sessions/{clear_name}.strses"
            with open(new_path_file, "w") as f:
                f.write(string)
        print(f"SUCCESSFULY convert {clear_name}")
        if len(proxy) > 3:
            data_for_add["proxy"] = proxy
        data_for_add["api_hash"] = api_hash
        data_for_add["api_id"] = api_id
        add_accountsjson({clear_name:data_for_add})
        print(f"SUCCESSFULY added {clear_name} to base ")
    except Exception as e:
        print(f"Error convert {clear_name}, moved to errors folder")
        if not exists("error_sessions"):
            mkdir("error_sessions")
        if len(file_name) > 3:
            replace(path_file, f"./error_sessions//{file_name}")
        input()


def add_accountsjson(data: dict):
    try:
        if exists("accounts.json"):
            with open("accounts.json", "r") as f:
                old_data = loads(f.read())
            data = {**old_data, **data}
        with open("accounts.json", "w") as f:
            dump(data, f, indent=4)
    except Exception as e:
        print("err on add_accountsjson: ", e)
        input()


def import_file_variable(variantone: bool):
    try:
        path = str(input("input full path to file: ")) # name_session:api_hash:api_id:proxy
        clear_info = {}
        with open(path, "r") as f:
            file = f.read().split("\n")
        for info in file:
            proxy = ""
            info_for_add = {}
            if len(info) < 3:
                continue
            info_split = info.split(":")
            if len(info_split) < 3:
                print("err lines: ", info)
                continue
            if not variantone:
                if len(info_split) > 3:
                    name, aid, ahash, proxy = info_split
                else:
                    name, aid, ahash = info_split
            else:
                if len(info_split) > 3:
                    name, ahash, aid, proxy = info_split
                else:
                    name, ahash, aid = info_split
            if len(proxy) > 3:
                info_for_add["proxy"] = str(proxy).strip()
            clear_name = (name.split(".")[0]).strip()
            info_for_add["api_hash"] = str(ahash).strip()
            info_for_add["api_id"] = int(str(aid).strip())
            clear_info[clear_name] = info_for_add
        add_accountsjson(clear_info)
    except Exception as e:
        print("Anything Error:", e)
        input()



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
    while True:
        try:
            result = input( "Hello, select variant:\n"\
                            "1 = import file name_session:api_hash:api_id:proxy // proxy can be empty\n"\
                            "2 = import file name_session:api_id:api_hash:proxy // proxy can be empty\n"
                            "3 = import new telegram accounts (by 1)\n"\
                            "0 = exit\n"\
                            "support proxies: socks5 and http/s\n"\
                            "format proxy: http://login:password@ip:port or http://ip:port \n"\
                            "support sessions: string sessions(.strses) and telethon sessions(.session)\n"\
                            "Select: "
                        )
            try:
                assert int(result) <= 3 and int(result) > 0
            except:
                exit(0)
            match int(result):
                case 0:
                    exit(0)
                case 1:
                    import_file_variable(True)
                    input("Successfully imported")
                    break
                case 2:
                    import_file_variable(False)
                    input("Successfully imported")
                    break
                case 3:
                    by_one()
        except Exception as e:
            print(e)
            input()