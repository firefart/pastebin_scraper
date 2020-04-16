# Pastebin Scraper

Constantly monitors the Pastebin scrape API and sends E-Mails when a keyword matches. This program needs a paid [Pastebin PRO account](https://pastebin.com/pro).
You need to put the IP you are scraping from into the [Pastebin admin panel](https://pastebin.com/api_scraping_faq).

The sent email contains the Paste metadata, the first matched line per keyword and the zipped paste as an attachment.

Keywords are set to match with a starting [regex boundary](https://www.regular-expressions.info/wordboundaries.html). Matching of CIDRs is also supported (see config.json.sample).

Expected errors during execution are also sent via E-Mail to the E-Mail address configured in `config.json`.

For sending mails you should setup a local SMTP server like postfix to handle resubmission, signing and so on for you. SMTP authentication is currently not implemented.

## Installation on a systemd based system

- Build binary or download it

```bash
make
```

or

```bash
go get
go build
```

or

```bash
make_linux.bat
make_windows.bat
```

- Add a user to run the binary

```bash
adduser --system pastebin
```

- Copy everything to home dir

```bash
cp -R checkout_dir /home/pastebin/
```

- Edit the config

```bash
cp /home/pastebin/config.json.sample /home/pastebin/config.json
vim /home/pastebin/config.json
```

- Install the service

```bash
cd /home/pastebin
./install_service.sh
```

- Watch the logs

```bash
journalctl -u pastebin_scraper.service -f
```

## Example Config

```json
{
  "mailserver": "localhost",
  "mailport": 25,
  "mailfrom": "Pastebin Alert <xxx@xxx.com>",
  "mailto": "Unknown Person <xxx@xxx.com>",
  "mailonerror": true,
  "mailtoerror": "error@xxx.xom",
  "timeout": "10s",
  "keywords": [
    {
      "keyword": "keyword1",
      "exceptions": ["exception1", "exception2", "exception3"]
    },
    {
      "keyword": "keyword2",
      "exceptions": ["exception1", "exception2", "exception3"]
    },
    {
      "keyword": "keyword3",
      "exceptions": ["exception1", "exception2", "exception3"]
    }
  ],
  "cidrs": ["10.0.0.0/8", "192.168.0.0/16"]
}
```
