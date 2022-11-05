package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
)

func ToUInt32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32([]byte(ip.To4()))
}

func FromUInt32(u uint32) net.IP {
	tmp := make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, u)
	return net.IP(tmp)
}

func MaskToUint32(net net.IPMask) uint32 {
	return binary.BigEndian.Uint32([]byte(net))
}

func IncIP(ipv4 net.IP) net.IP {
	i := ToUInt32(ipv4)
	if i == math.MaxUint32 {
		return ipv4
	}
	return FromUInt32(i + 1)

}

func BroadcastAdd(net *net.IPNet) net.IP {
	i := ToUInt32(net.IP)
	return FromUInt32(i + NetSize(net))
}

func NetSize(net *net.IPNet) uint32 {
	mask := MaskToUint32(net.Mask)
	return ^mask
}

func Ping(ipv4 net.IP, wg *sync.WaitGroup, hostCounter *int32) {
	out, err := exec.Command("ping", "-n", "2", ipv4.String()).Output()
	if err != nil {
		log.Fatalf("Ping host %s failed. Exit status 1", ipv4.String())
		// fmt.Print(err, ipv4.String())
	}
	if !strings.Contains(string(out), "Destination host unreachable") && !strings.Contains(string(out), "Заданный узел недоступен") {
		atomic.AddInt32(hostCounter, 1)
		fmt.Printf("Host %s is up\n", ipv4.String())
	}
	wg.Done()
}

type Host struct {
	Ipv4        net.IP
	ActivePorts []int
}

func tcpScan(ipv4 net.IP, wg *sync.WaitGroup, ports []int) {
	defer wg.Done()
	var host Host
	host.Ipv4 = ipv4
	for _, p := range ports {
		address := fmt.Sprintf("%s:%d", ipv4, p)
		conn, err := net.Dial("tcp", address)
		if err != nil {
			continue
		}
		conn.Close()
		host.ActivePorts = append(host.ActivePorts, p)

	}

	fmt.Println(host)

}

func main() {

	var scanType = flag.String("t", "ping", "")
	var netAddr = flag.String("a", "192.168.1.0/24", "CIDR notation")
	var startPort = flag.Int("sPort", 1, "start port to tcp scan")
	var endPort = flag.Int("ePort", 1, "end port to tcp scan")
	flag.Parse()
	fmt.Print(*scanType)

	var hostCounter int32 = 0

	_, ipv4Net, err := net.ParseCIDR(*netAddr)
	ipv4add := IncIP(ipv4Net.IP)
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}

	switch *scanType {
	case "ping":
		for i := 0; i < int(NetSize(ipv4Net))-1; i++ {
			wg.Add(1)
			go Ping(ipv4add, wg, &hostCounter)
			ipv4add = IncIP(ipv4add)
		}
	case "tcp":

		ports := make([]int, 0)
		for i := *startPort; i < *endPort+1; i++ {
			ports = append(ports, i)
		}
		fmt.Println(ports)
		for i := 0; i < int(NetSize(ipv4Net))-1; i++ {
			wg.Add(1)
			go tcpScan(ipv4add, wg, ports)
			ipv4add = IncIP(ipv4add)
		}

	}
	wg.Wait()
	fmt.Printf("\n%d hosts up", hostCounter)

}
