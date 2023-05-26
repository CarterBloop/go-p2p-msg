# go-p2p-msg

A Peer to Peer messaging application made in golang

# How to test on live

Carter:
./p2pmsg -listen-port=8080 -listen-ip=0.0.0.0 -peer-ip=73.222.193.134 -peer-port=8080 -username=John

Cody:
./p2pmsg -listen-port=8080 -listen-ip=0.0.0.0 -peer-ip=169.233.235.194 -peer-port=8080 -username=Cody

# How to test on localhost

./p2pmsg -listen-port=8080 -peer-port=8081 -username=John
./p2pmsg -listen-port=8081 -peer-port=8080 -username=Bob
