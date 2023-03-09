#! /bin/bash

set -e

echo "build..."
GOOS=linux GOARCH=amd64 go build -o botastic
echo "scp..."
scp botastic ptest-dolphin:/home/ubuntu/botastic/botastic-new
read -p "Do you need to synchronize the database migration? Enter Y/N to continue." -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]
then
  echo "migrate database..."
  ssh ptest-dolphin "cd /home/ubuntu/botastic && ./botastic-new migrate up"
fi
echo "restart..."
ssh ptest-dolphin "cd /home/ubuntu/botastic && mv botastic-new botastic && sudo systemctl restart botastic-api.service"
echo "🙆 deployed!"
