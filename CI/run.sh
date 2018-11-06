#!/bin/sh

# build binary file
mkdir -p /root/go/src/resolver

cp *.go /root/go/src/resolver
cp -r vendor /root/go/src/resolver
cp -r vendor/github.com /root/go/src/

CGO_ENABLED=0 GOOS=linux go build -o ./resolver resolver


# Run Binary
./resolver -v &
sleep 5

# Check
CURL='curl -s -o /dev/null -w '%{http_code}' '
BASEURL='http://127.0.0.1:18888'

checkStatusCode() {
        if [ "$1" -ne  $2 ] 
          then
	    echo -e "Not expected status code.\nIt supposed to be $2 not $1"
	    echo "Exiting with status code 1"
	    kill `pidof ./resolver`
	    exit 1
	  else
	    echo -e "Got expected status code. All good so far"
	fi
}

# always working
echo "GET $BASEURL/stats"
mustBeSuccess=`$CURL $BASEURL/stats`

# 404 not found
echo "GET $BASEURL/notexistingpage"
mustBeNotFound=`$CURL $BASEURL/notexistingpage`

#500 db con err
echo "GET $BASEURL/cloud92/instance/instance-00000000"
mustBeError=`$CURL $BASEURL/cloud92/instance/instance-00000000`

# CHeck stats code
echo
echo  

echo "Checking /stats status code. Must be 200"
checkStatusCode $mustBeSuccess 200

echo "Checking /notexist status code. Must be 404"
checkStatusCode $mustBeNotFound 404

echo "Checking /cloud91/... status code. Must be 500"
checkStatusCode $mustBeError 500


kill `pidof ./resolver`
echo "Done"
