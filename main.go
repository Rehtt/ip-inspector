package main

import (
	"flag"
	"log"
	"net"
	"strconv"
	"sync"
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
)

func main() {
	portsFlag := flag.String("ports", "", "comma-separated TCP ports to listen on (required)")
	flag.Parse()

	ports, err := parsePorts(*portsFlag)
	if err != nil {
		log.Fatal(err)
	}

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

	listeners := make([]net.Listener, 0, len(ports))
	for _, port := range ports {
		addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(port))
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Printf("failed to listen on port %d: %v", port, err)
			continue
		}
		listeners = append(listeners, listener)
		log.Printf("listening on %s", addr)
	}
	if len(listeners) == 0 {
		log.Fatal("failed to listen on any configured port")
	}

	var wg sync.WaitGroup
	for _, listener := range listeners {
		wg.Add(1)
		go func(listener net.Listener) {
			defer wg.Done()
			acceptConnections(listener)
		}(listener)
	}
	wg.Wait()
}

func acceptConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		return
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return
	}

	db.Create(&ConfinementCell{
		Ip:   ip.String(),
		Time: time.Now(),
	})
	if err := netlink.IpsetAdd(ipsetName, &netlink.IPSetEntry{
		IP:      ip,
		Replace: true,
	}); err != nil {
		panic(err)
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
