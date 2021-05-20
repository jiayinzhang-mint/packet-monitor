package xdp

import (
	"context"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/mintxtinm/packet-monitor/pkg/graph"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// Instance xdp tracer instance
type Instance struct {
	IfIndexName string

	DepGraph     *graph.Graph
	PacketSource *gopacket.PacketSource
}

// Config xdp tracer setup config
type Config struct {
	IfIndexName string
}

// New start new xdp tracer instance
func New(ctx context.Context, c *Config) (i *Instance) {

	return &Instance{
		IfIndexName: c.IfIndexName,
		DepGraph:    &graph.Graph{},
	}
}

// Load load xdp tracer
func (i *Instance) Load(c *cli.Context) (err error) {

	// load instruction
	handle, err := pcap.OpenLive(i.IfIndexName, 1600, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}

	bpfInstructions := []pcap.BPFInstruction{
		{Code: 0x6, Jt: 0, Jf: 0, K: 0x00040000},
		{Code: 0x6, Jt: 0, Jf: 0, K: 0x00000000},
	}

	defer handle.Close()
	if err := handle.SetBPFInstructionFilter(bpfInstructions); err != nil {
		panic(err)
	}

	i.PacketSource = gopacket.NewPacketSource(handle, handle.LinkType())

	// parse events
	for packet := range i.PacketSource.Packets() {
		var (
			ipv4 *layers.IPv4
			ipv6 *layers.IPv6

			tcp *layers.TCP

			node = &graph.Node{}
			edge = &graph.Edge{}
		)

		edge.Length = packet.Metadata().Length

		// L1 Data link layer (OSI L2)
		if packet.LinkLayer() == nil {
			continue
		}
		switch packet.LinkLayer().LayerType() {
		case layers.LayerTypeEthernet:
		default:
			continue
		}

		// L2 network layer (OSI L3)
		if packet.NetworkLayer() == nil {
			continue
		}
		switch packet.NetworkLayer().LayerType() {
		case layers.LayerTypeIPv4:
			ipv4 = packet.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
			node.PodIP = ipv4.SrcIP.String()

			edge.SourceIP = ipv4.SrcIP.String()
			edge.DestIP = ipv4.DstIP.String()

		case layers.LayerTypeIPv6:
			ipv6 = packet.Layer(layers.LayerTypeIPv6).(*layers.IPv6)

			node.PodIP = ipv6.SrcIP.String()

			edge.SourceIP = ipv6.SrcIP.String()
			edge.DestIP = ipv6.DstIP.String()

		case layers.LayerTypeICMPv4:
		case layers.LayerTypeICMPv6:
		default:
			continue
		}

		// L3 transport layer (OSI L4)
		if packet.TransportLayer() == nil {
			continue
		}
		switch packet.TransportLayer().LayerType() {
		case layers.LayerTypeTCP:
			tcp = packet.Layer(layers.LayerTypeTCP).(*layers.TCP)
			edge.Protocol = "TCP"
			edge.SourcePort = uint16(tcp.SrcPort)
			edge.DestPort = uint16(tcp.DstPort)

		default:
			continue
		}

		// L4 application layer (OSI L7)
		if packet.ApplicationLayer() == nil {
			continue
		}

		if edge.SourceIP != "" {
			if c.Bool("verbose") {
				logrus.Info(edge)
			}
			i.DepGraph.AddEdge(edge)
		}

	}

	return
}
