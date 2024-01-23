# notcoin_go

## Requirements:
- Windows/Linux 64bit
- python v3.9+
- telethon/string sessions

## Features:
- Multithreading
- Random farming
- Automatic purchase of upgrades
- Automatic activation of boosts
- Random delay between clicks
- Support for http(s) and socks5 proxies (1 account = 1 proxy)

## Instructions:
1. Install the necessary libraries with the command `pip install -r requirements.txt`
2. Open `create_sessions.py`. After performing all necessary actions, the script will save the sessions to the `telethon_sessions` folder (it will be created after running the script). If you created new sessions, you can skip step 3, as you already have a string session (needed for work), which is placed in the `str_sessions` folder.
3. Open `convert_sessions.py` (or immediately move the sessions to `str_sessions` if you initially have a session string, BUT you need to rename it to `.strsess`).
4. Run the file.

The `.env` file stores variables that indicate to what levels to download upgrades and whether to use daily boosts. Initial values:

```env
# Enable boost activation
enable_daily_boosters = 1

# Bot farming level, 1 - maximum
max_tapbot = 0

# More coins per click, this is important for the bot, otherwise it may give errors that it’s farming too fast
max_multitab = 8

# Energy recovery speed, maximum
max_recharging = 3

# Energy reserve, maximum
max_energy_limit = 9
