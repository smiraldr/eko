
package cmd

import (
	"fmt"
	"time"
    "os"
    "log"
    "net"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv6"
	"github.com/spf13/cobra"
)

var dest6 string
var packetsRecv6 int
var packetsSent6 int
var packetLoss6 float64

// pingip6Cmd represents the pingip6 command
const (
	// as iana import was forbidden
    ProtocolIcmp6 = 58
)
// ListenTo to listen on all IPv6 interfaces
var ListenTo = "::"

// Ping6 Function to ping ipv6 address only
func Ping6(addr string) (net.IP,*net.IPAddr, time.Duration,  float64, error) {
	// Get own device IP for origin IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
    // Listen for icmp reply
    c, err := icmp.ListenPacket("ip6:ipv6-icmp", ListenTo)
    if err != nil {
        return localAddr.IP,nil, 0, 0, err
    }
    defer c.Close()

    // Resolution incase Dns is use for ip
    dest6, err := net.ResolveIPAddr("ip6:ipv6-icmp", addr)
    if err != nil {
        panic(err)
        return localAddr.IP,nil, 0, 0, err
    }

    // Forming icmp message and formatting properly using marshal
    mes := icmp.Message{
        Type: ipv6.ICMPTypeEchoRequest, Code: 0,
        Body: &icmp.Echo{
            ID: os.Getpid() & 0xffff, Seq: 1,
            Data: []byte(""),
        },
    }
    bytes, err := mes.Marshal(nil)
    if err != nil {
        return localAddr.IP,dest6, 0, 0, err
    }

    // Transmit packet to destination
    start := time.Now()  //start time is recorded so can be used to get rtt
    n, err := c.WriteTo(bytes, dest6)
    packetsSent6++         //packetsent count for calculating loss of packets
    if err != nil {
        return localAddr.IP,dest6, 0, 0, err
    } else if n != len(bytes) {
        return localAddr.IP,dest6, 0, 0,fmt.Errorf("got %v; want %v", n, len(bytes))
    }

    // Listen for reply from destination
    reply := make([]byte, 1500)
    err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
    if err != nil {
        return localAddr.IP,dest6, 0, 0, err
    }
    n, peer, err := c.ReadFrom(reply)
    packetsRecv6++          //packetrecv count for calculating loss
    if err != nil {
        return localAddr.IP,dest6, 0, 0, err
    }
    rtt := time.Since(start)  //rtt calculated
        
    loss := float64(packetsSent6-packetsRecv6) / float64(packetsSent6) * 100 // loss calculation
	
	//process the message
    rm, err := icmp.ParseMessage(ProtocolIcmp6, reply[:n])
    if err != nil {
        return localAddr.IP,dest6, 0, 0, err
    }
    switch rm.Type {
    case ipv6.ICMPTypeEchoReply:
        return localAddr.IP,dest6, rtt, loss, nil
    default:
        return localAddr.IP,dest6, 0, loss, fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
    }
}

var pingip6Cmd = &cobra.Command{
	Use:   "pingip6",
	Short: "Use this to ping ipv6 address",
    Long: `Use this with flag -6 to ping an ipv6 address
    For Example :
    eko pingip6 -6 google.com or eko pingip6 -6 (ipv6 address)
    `,
	Run: func(cmd *cobra.Command, args []string) {
			p := func(addr string){
			or, dest6, dur,loss ,err := Ping6(addr) //call ping
			if err != nil {
				log.Printf("Ping %s (%s): %s\n", addr, dest6, err)
				return
			}
			log.Printf("Ping from %s to %s resolved to %s : Rtt: %s loss: %f percent \n",or, addr, dest6, dur, loss)
		}
		
		for true{
			p(dest6) //infinite loop until terminated with 2 second delay
			time.Sleep(2 * time.Second)
		}
	},
}

func init() {
    // packetsRecv6= 10
    // packetsSent6= 10
	rootCmd.AddCommand(pingip6Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingip6Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pingip6Cmd.Flags().StringVarP(&dest6, "6", "", "Help message for toggle")
	pingip6Cmd.Flags().StringVarP(&dest6, "ipv6", "6", "", "Enter ipv6 address to ping") //ipv6 flag
}
