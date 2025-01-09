package main

import (
	"net"
	"strings"
	"time"

	"github.com/coreos/go-iptables/iptables"
	"github.com/vishvananda/netlink"
	"gorm.io/gorm"
)

var (
	tables *iptables.IPTables
	db     *gorm.DB

	// 禁闭24小时
	timeLength  = 24 * time.Hour
	timeLengthS = uint32(timeLength.Seconds())
	ipsetName   = "blacklist"

	addr = "0.0.0.0:3306"
)

func main() {
	var err error
	tables, err = iptables.New()
	if err != nil {
		panic(err)
	}

	db, err = OpenDB()
	if err != nil {
		panic(err)
	}

	if err := netlink.IpsetCreate(ipsetName, "hash:ip", netlink.IpsetCreateOptions{
		Replace: true,
		Timeout: &timeLengthS,
	}); err != nil {
		panic(err)
	}
	if err := tables.AppendUnique("filter", "INPUT", "-m", "set", "--match-set", ipsetName, "src", "-j", "DROP"); err != nil {
		panic(err)
	}
	var c []*ConfinementCell
	db.Find(&c)
	for _, v := range c {
		if s := time.Since(v.Time); s >= timeLength {
			db.Where("ip = ?", v.Ip).Delete(&v)
		} else {
			ss := uint32(s.Seconds())
			netlink.IpsetAdd(ipsetName, &netlink.IPSetEntry{
				IP:      net.ParseIP(v.Ip),
				Timeout: &ss,
				Replace: true,
			})
		}
	}

	go func() {
		t := time.NewTicker(time.Hour)
		for {
			<-t.C
			if err := Patrol(); err != nil {
				panic(err)
			}
		}
	}()

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}
		ip := net.ParseIP(strings.Split(conn.RemoteAddr().String(), ":")[0])
		if len(ip) != 0 {
			db.Create(&ConfinementCell{
				Ip:   ip.String(),
				Time: time.Now(),
			})
			err := netlink.IpsetAdd(ipsetName, &netlink.IPSetEntry{
				IP:      ip,
				Replace: true,
			})
			if err != nil {
				panic(err)
			}
		}
		conn.Close()
	}
}

func Patrol() error {
	var c []*ConfinementCell
	if err := db.Find(&c).Error; err != nil {
		return err
	}
	for _, v := range c {
		if s := time.Since(v.Time); s >= timeLength {
			db.Where("ip = ?", v.Ip).Delete(&v)
		}
	}
	return nil
}
