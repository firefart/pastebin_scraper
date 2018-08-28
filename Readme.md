# Pastebin Scraper

[![Build Status](https://travis-ci.org/FireFart/pastebin_scraper.svg?branch=master)](https://travis-ci.org/FireFart/pastebin_scraper)

Constantly monitors the Pastebin scrape API and sends E-Mails when a keyword matches. This program needs a paid [Pastebin PRO account](https://pastebin.com/pro).
You need to put the IP you are scraping from into the [Pastebin admin panel](https://pastebin.com/api_scraping_faq).

The sent email contains the Paste metadata, the first matched line per keyword and the zipped paste as an attachment.

Keywords are set to match with a starting [regex boundary](https://www.regular-expressions.info/wordboundaries.html).

Expected errors during execution are also sent via E-Mail to the E-Mail address configured in `config.json`.

For sending mails you should setup a local SMTP server like postfix to handle resubmission, signing and so on for you. SMTP authentication is currently not implemented.

## Installation on a systemd based system

* Build binary or download it

```bash
make
```

or

```bash
go get -u gopkg.in/gomail.v2
go build
```

or

```bash
make_linux.bat
make_windows.bat
```

* Add a user to run the binary

```bash
adduser --system pastebin
```

* Copy everything to home dir

```bash
cp -R checkout_dir /home/pastebin/
```

* Edit the config

```bash
cp /home/rss/config.json.sample /home/rss/config.json
vim /home/rss/config.json
```

* Install the service

```bash
./install_service.sh
```

* Watch the logs

```bash
journalctl -u pastebin_scraper.service -f
```
