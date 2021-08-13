#!/bin/sh

#
# This script curl to mapi and iterate until it successfully add a bitcoin node to mAPI

set -e
ret=0
while [[ $ret -ne 200 ]]
do
    echo -e "\n\nAttempt to add bitcoin node to mapi"
    ret=$(curl -k --write-out %{http_code} -H "Api-Key: apikey" -H "Content-Type: application/json" -X POST http://mapi:80/api/v1/Node -d "{ \"id\" : \"node1:18332\", \"username\": \"bitcoin\", \"password\": \"bitcoin\",  \"ZMQNotificationsEndpoint\": \"tcp://node1:28332\"}")
    sleep 1
done
