GoReplay is the simplest and safest way to test your app using real traffic before you put it into production.

简单来说就是goreplay抓取线上真实的流量，并将捕捉到的流量转发到测试服务器上

# 安装
github release
```sh
wget https://github.com/buger/goreplay/releases/download/v1.2.0/gor_v1.2.0_x64.tar.gz
```
# 录制
```sh
sudo ./gor --input-raw :8080 --output-file requests.log --http-allow-url /perfads
```
# 回放
```sh
./gor --input-file ./req.log --output-http="http://appclick-test.rqmob.com"
```
