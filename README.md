# wattmonitor

Wi-SUNモジュール[BP35A1](http://www.rohm.co.jp/web/japan/products/-/product/BP35A1)をUSB接続してスマートメーターから瞬時電力値を取得し、fluentdに投げるツールです。

適当に作ったツールなので動作の保証などはありません。

## 準備

PCでは下記の準備が必要かもしれませんが、Raspberry pi 3(raspbian)では必要ありませんでした。

* `/dev/ttyUSB0` の読み書きができるようにします。

```
cat <<EOF | sudo tee -a /etc/udev/rules.d/50-udev.rules
KERNEL=="tty[A-Z]*[0-9]|pppox[0-9]*|ircomm[0-9]*|noz[0-9]*|rfcomm[0-9]*", GROUP="dialout", MODE="0666"
EOF
sudo adduser <your_username> dialout
```

## インストール

* goのインストールをしておいてください

```
go get -u github.com/higebu/wattmonitor
```

* Raspberry pi 3用のバイナリのビルド

```
cd $GOPATH/src/github.com/higebu/wattmonitor
GOARCH=arm GOARM=7 go build
```

## 使い方

* help

```
./wattmonitor -h
Usage of ./wattmonitor:
  -fluent_host string
        fluentdのIPかホスト名 (default "127.0.0.1")
  -fluent_port int
        fluentdのポート (default 24224)
  -interval int
        監視間隔 短すぎると固まる (default 60)
  -pwd string
        Bルートサービスのパスワード
  -rbid string
        Bルートサービスの認証ID
  -tag string
        fluentdで使うタグ (default "wattmonitor.watt")
```
