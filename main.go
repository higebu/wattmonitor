package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/higebu/wattmonitor/bp35a1"
)

func main() {
	pwd := flag.String("pwd", "", "Bルートサービスのパスワード")
	rbID := flag.String("rbid", "", "Bルートサービスの認証ID")
	fluentPort := flag.Int("fluent_port", 24224, "fluentdのポート")
	fluentHost := flag.String("fluent_host", "127.0.0.1", "fluentdのIPかホスト名")
	tag := flag.String("tag", "wattmonitor.watt", "fluentdで使うタグ")
	interval := flag.Int("interval", 60, "監視間隔 短すぎると固まる")
	flag.Parse()
	logger, err := fluent.New(fluent.Config{
		FluentHost: *fluentHost,
		FluentPort: *fluentPort,
	})
	defer logger.Close()

	log.SetLevel(log.DebugLevel)
	b := bp35a1.NewBP35A1()
	b.Connect()
	log.Info("open")
	defer b.Close()
	time.Sleep(time.Second * 1)
	// Get Version
	v, err := b.SKVER()
	if err != nil {
		log.Error(err)
	}
	log.Info(v)
	// Set PWD
	err = b.SKSETPWD(*pwd)
	if err != nil {
		log.Error(err)
	}
	// Set RBID
	err = b.SKSETRBID(*rbID)
	if err != nil {
		log.Error(err)
	}
	// Search PAN
	log.Info("Searching PAN with SKSCAN...")
	pan, err := b.SKSCAN()
	if err != nil {
		log.Error(err)
	}
	log.Info("Set Channel to S2 register...")
	err = b.SKSREG("S2", pan.Channel)
	if err != nil {
		log.Error(err)
	}
	log.Info("Set PanID to S3 register...")
	err = b.SKSREG("S3", pan.PanID)
	if err != nil {
		log.Error(err)
	}
	log.Info("Get IPv6 Addr with SKLL64...")
	ipv6Addr, err := b.SKLL64(pan.Addr)
	if err != nil {
		log.Error(err)
	}
	log.Info("IPv6 Addr is " + ipv6Addr)
	log.Info("SKJOIN...")
	err = b.SKJOIN(ipv6Addr)
	if err != nil {
		log.Error(err)
	}
	d := []byte{0x10, 0x81, 0x00, 0x01, 0x05, 0xFF, 0x01, 0x02, 0x88, 0x01, 0x62, 0x01, 0xE7, 0x00}
	for {
		log.Info("SKSENDTO...")
		r, err := b.SKSENDTO("1", ipv6Addr, "0E1A", "1", d)
		if err != nil {
			log.Error(err)
		}
		a := strings.Split(r, " ")
		if len(a) != 9 {
			log.Warn(r)
		}
		if a[7] != "0012" {
			log.Warn(fmt.Sprintf("%s is not 0012. ", a[7]))
			continue
		}
		o := a[8]
		w, err := strconv.ParseUint(o[len(o)-8:], 16, 0)
		if err != nil {
			log.Error(err)
		}
		t := time.Now()
		log.Info(t, w)
		data := map[string]string{
			"timestamp": time.Time(t).UTC().Format("2006-01-02T15:04:05.000Z"),
			"watt":      strconv.FormatUint(w, 10),
		}
		err = logger.Post(*tag, data)
		if err != nil {
			log.Error(err)
		}
		time.Sleep(time.Second * time.Duration(*interval))
	}
}
