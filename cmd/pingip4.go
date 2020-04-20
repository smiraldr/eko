
package cmd

import (
	"fmt"
	"time"
    "os"
    "log"
    "net"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"github.com/spf13/cobra"
)

var dest string
var packetsRecv int
var packetsSent int
var packetLoss float64

// pingip4Cmd represents the pingip4 command
const (
	// as iana import was forbidden
    ProtocolIcmp = 1
)
// ListenOn to listen on all IPv4 interfaces
var ListenOn = "0.0.0.0"

// Ping Function to ping ipv4 address only
func Ping(addr string) (net.IP,*net.IPAddr, time.Duration,  float64, error) {
	// Get own device IP for origin IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
    // Listen for icmp reply
    c, err := icmp.ListenPacket("ip4:icmp", ListenOn)
    if err != nil {
        return localAddr.IP,nil, 0, 0, err
    }
    defer c.Close()

    // Resolution incase Dns is use for ip
    dest, err := net.ResolveIPAddr("ip4", addr)
    if err != nil {
        panic(err)
        return localAddr.IP,nil, 0, 0, err
    }

    // Forming icmp message and formatting properly using marshal
    mes := icmp.Message{
        Type: ipv4.ICMPTypeEcho, Code: 0,
        Body: &icmp.Echo{
            ID: os.Getpid() & 0xffff, Seq: 1,
            Data: []byte(""),
        },
    }
    bytes, err := mes.Marshal(nil)
    if err != nil {
        return localAddr.IP,dest, 0, 0, err
    }

    // Transmit packet to destination
    start := time.Now()  //start time is recorded so can be used to get rtt
    n, err := c.WriteTo(bytes, dest)
    packetsSent++         //packetsent count for calculating loss of packets
    if err != nil {
        return localAddr.IP,dest, 0, 0, err
    } else if n != len(bytes) {
        return localAddr.IP,dest, 0, 0,fmt.Errorf("got %v; want %v", n, len(bytes))
    }

    // Listen for reply from destination
    reply := make([]byte, 1500)
    err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
    if err != nil {
        return localAddr.IP,dest, 0, 0, err
    }
    n, peer, err := c.ReadFrom(reply)
    packetsRecv++          //packetrecv count for calculating loss
    if err != nil {
        return localAddr.IP,dest, 0, 0, err
    }
    rtt := time.Since(start)  //rtt calculated
        
    loss := float64(packetsSent-packetsRecv) / float64(packetsSent) * 100 // loss calculation
	
	//process the message
    rm, err := icmp.ParseMessage(ProtocolIcmp, reply[:n])
    if err != nil {
        return localAddr.IP,dest, 0, 0, err
    }
    switch rm.Type {
    case ipv4.ICMPTypeEchoReply:
        return localAddr.IP,dest, rtt, loss, nil
    default:
        return localAddr.IP,dest, 0, loss, fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
    }
}

var pingip4Cmd = &cobra.Command{
	Use:   "pingip4",
	Short: "Use this to ping ipv4 address",
    Long: `Use this with flag -4 to ping an ipv4 address
    For Example :
    eko pingip4 -4 google.com or pingip4 -4 (ipv4 address)
    `,
	Run: func(cmd *cobra.Command, args []string) {
			p := func(addr string){
			or, dest, dur,loss ,err := Ping(addr) //call ping
			if err != nil {
				log.Printf("Ping %s (%s): %s\n", addr, dest, err)
				return
			}
			log.Printf("Ping from %s to %s resolved to %s : Rtt: %s loss: %f percent \n",or, addr, dest, dur, loss)
		}
		
		for true{
			p(dest) //infinite loop until terminated with 2 second delay
			time.Sleep(2 * time.Second)
		}
	},
}

func init() {
    // packetsRecv= 10
    // packetsSent= 10
	rootCmd.AddCommand(pingip4Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingip4Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pingip4Cmd.Flags().StringVarP(&dest, "4", "", "Help message for toggle")
	pingip4Cmd.Flags().StringVarP(&dest, "ipv4", "4", "", "Enter ipv4 address to ping") //ipv4 flag
}
