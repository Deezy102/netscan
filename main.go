package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
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
	defer wg.Done()

	out, err := exec.Command("ping", "-n", "2", ipv4.String()).Output()
	if err != nil {
		log.Fatalf("Ping host %s failed. Exit status 1", ipv4.String())
	}
	if !strings.Contains(string(out), "Destination host unreachable") {
		atomic.AddInt32(hostCounter, 1)
		fmt.Printf("Host %s is up\n", ipv4.String())
	}
}

func main() {
	_, ipv4Net, err := net.ParseCIDR(os.Args[1])
	ipv4add := IncIP(ipv4Net.IP)
	if err != nil {
		log.Fatal(err)
	}

	var hostCounter int32 = 0
	start := time.Now()
	wg := &sync.WaitGroup{}

	for i := 0; i < int(NetSize(ipv4Net))-1; i++ {
		wg.Add(1)
		go Ping(ipv4add, wg, &hostCounter)
		ipv4add = IncIP(ipv4add)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("\n%d hosts up", hostCounter)
	fmt.Printf("\nTime spent: %.2f seconds\n", elapsed.Seconds())
}
