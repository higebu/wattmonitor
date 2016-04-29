package bp35a1

import (
	"bufio"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/tarm/serial"
)

type BP35A1 struct {
	Baudrate     int
	SerialDevice string
	Port         *serial.Port
}

func NewBP35A1() *BP35A1 {
	return &BP35A1{
		Baudrate:     115200,
		SerialDevice: "/dev/ttyUSB0",
	}
}

func (b *BP35A1) Connect() error {
	c := &serial.Config{
		Name: b.SerialDevice,
		Baud: b.Baudrate,
	}
	s, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	b.Port = s
	return nil
}

func (b *BP35A1) Close() {
	b.Port.Close()
}

func (b *BP35A1) write(s string) error {
	_, err := b.Port.Write([]byte(s))
	if err != nil {
		return err
	}
	return nil
}

func (b *BP35A1) readUntilOK() ([]string, error) {
	reader := bufio.NewReader(b.Port)
	scanner := bufio.NewScanner(reader)
	var reply []string
	for scanner.Scan() {
		l := scanner.Text()
		log.Debug(l)
		reply = append(reply, l)
		if l == "OK" {
			break
		}
	}
	return reply, nil
}

func (b *BP35A1) SKVER() (string, error) {
	err := b.write("SKVER\r\n")
	if err != nil {
		return "", err
	}
	lines, err := b.readUntilOK()
	if err != nil {
		return "", err
	}
	return strings.Split(lines[1], " ")[1], nil
}

func (b *BP35A1) SKSETPWD(pwd string) error {
	err := b.write("SKSETPWD C " + pwd + "\r\n")
	if err != nil {
		return err
	}
	return nil

}

func (b *BP35A1) SKSETRBID(rbid string) error {
	err := b.write("SKSETRBID " + rbid + "\r\n")
	if err != nil {
		return err
	}
	return nil
}

func (b *BP35A1) SKSCAN() (*PAN, error) {
	err := b.write("SKSCAN 2 FFFFFFFF 6\r\n")
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(b.Port)
	scanner := bufio.NewScanner(reader)
	pan := &PAN{}
	for scanner.Scan() {
		l := scanner.Text()
		log.Debug(l)
		switch {
		case strings.Contains(l, "Channel:"):
			pan.Channel = strings.Split(l, ":")[1]
		case strings.Contains(l, "Channel Page:"):
			pan.ChannelPage = strings.Split(l, ":")[1]
		case strings.Contains(l, "Pan ID:"):
			pan.PanID = strings.Split(l, ":")[1]
		case strings.Contains(l, "Addr:"):
			pan.Addr = strings.Split(l, ":")[1]
		case strings.Contains(l, "LQI:"):
			pan.LQI = strings.Split(l, ":")[1]
		case strings.Contains(l, "PairID:"):
			pan.PairID = strings.Split(l, ":")[1]
		}
		if strings.Contains(l, "EVENT 22 ") {
			break
		}
	}
	return pan, nil
}

func (b *BP35A1) SKSREG(k, v string) error {
	err := b.write("SKSREG " + k + " " + v + "\r\n")
	if err != nil {
		return err
	}
	_, err = b.readUntilOK()
	if err != nil {
		return err
	}
	return nil
}

func (b *BP35A1) SKLL64(addr string) (string, error) {
	err := b.write("SKLL64 " + addr + "\r\n")
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(b.Port)
	r, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}
	log.Debug(r)
	r, _, err = reader.ReadLine()
	if err != nil {
		return "", err
	}
	log.Debug(r)
	return string(r), nil
}

func (b *BP35A1) SKJOIN(ipv6Addr string) error {
	err := b.write("SKJOIN " + ipv6Addr + "\r\n")
	if err != nil {
		return err
	}
	reader := bufio.NewReader(b.Port)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		l := scanner.Text()
		log.Debug(l)
		if strings.Contains(l, "FAIL ") {
			return fmt.Errorf("Failed to SKJOIN. %s", l)
		}
		if strings.Contains(l, "EVENT 25 ") {
			break
		}
	}
	if scanner.Scan() {
		log.Debug(scanner.Text())
	}
	return nil
}

func (b *BP35A1) SKSENDTO(handle, ipAddr, port, sec string, data []byte) (string, error) {
	s := fmt.Sprintf("SKSENDTO %s %s %s %s %.4X ", handle, ipAddr, port, sec, len(data))
	d := append([]byte(s), data[:]...)
	d = append(d, []byte("\r\n")[:]...)
	_, err := b.Port.Write(d)
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(b.Port)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		l := scanner.Text()
		log.Debug(l)
		if strings.Contains(l, "FAIL ") {
			return "", fmt.Errorf("Failed to SKSENDTO. %s", l)
		}
		if strings.Contains(l, "ERXUDP ") {
			return l, nil
		}
	}
	return "", nil
}
