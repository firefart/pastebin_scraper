#!/bin/sh

echo "Copying unit file"
cp /home/pastebin/pastebin_scraper.service /etc/systemd/system/pastebin_scraper.service
echo "reloading systemctl"
systemctl daemon-reload
echo "enabling service"
systemctl enable pastebin_scraper.service
systemctl start pastebin_scraper.service
# sleep some time to check if binary crashed
sleep 5
systemctl status pastebin_scraper.service
