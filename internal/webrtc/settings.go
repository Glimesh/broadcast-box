package webrtc

import (
	"net"
	"strings"

	"github.com/pion/dtls/v3/pkg/crypto/elliptic"
	"github.com/pion/ice/v4"
	"github.com/pion/webrtc/v4"

	"log"
	"os"
	"strconv"

	"github.com/glimesh/broadcast-box/internal/ip"
)

func GetSettingEngine(isWhip bool, tcpMuxCache map[string]ice.TCPMux, udpMuxCache map[int]*ice.MultiUDPMuxDefault) (settingEngine webrtc.SettingEngine) {
	var (
		udpMuxOpts   []ice.UDPMuxFromPortOption
		networkTypes []webrtc.NetworkType
	)

	setupNetworkTypes(networkTypes)
	setupNAT(settingEngine)
	setupInterfaceFilter(settingEngine, udpMuxOpts)
	setupUDPMux(settingEngine, isWhip, udpMuxCache, udpMuxOpts)
	setupTCPMux(settingEngine, tcpMuxCache)

	settingEngine.SetDTLSEllipticCurves(elliptic.X25519, elliptic.P384, elliptic.P256)
	settingEngine.SetNetworkTypes(networkTypes)
	settingEngine.DisableSRTCPReplayProtection(true)
	settingEngine.DisableSRTPReplayProtection(true)
	settingEngine.SetIncludeLoopbackCandidate(os.Getenv("INCLUDE_LOOPBACK_CANDIDATE") != "")

	return
}

func setupNetworkTypes(networkTypes []webrtc.NetworkType) {
	networkTypesEnv := os.Getenv("NETWORK_TYPES")
	tcpMuxForce := os.Getenv("TCP_MUX_FORCE")

	// TCP Mux Force will enforce TCP4/6 instead of requested types
	if tcpMuxForce != "" {
		networkTypes = []webrtc.NetworkType{
			webrtc.NetworkTypeTCP4,
			webrtc.NetworkTypeTCP6,
		}
	}

	if networkTypesEnv != "" {
		for networkTypeStr := range strings.SplitSeq(networkTypesEnv, "|") {
			networkType, err := webrtc.NewNetworkType(networkTypeStr)
			if err != nil {
				networkTypes = append(networkTypes, networkType)
			}
		}
	} else {
		// No network types found, use default values
		networkTypes = append(networkTypes, webrtc.NetworkTypeUDP4, webrtc.NetworkTypeUDP6)
	}

}

func setupTCPMux(settingEngine webrtc.SettingEngine, tcpMuxCache map[string]ice.TCPMux) {
	// Use TCP Mux port if set
	if tcpAddr := getTCPMuxAddress(); tcpAddr != nil {
		address := os.Getenv("TCP_MUX_ADDRESS")
		tcpMux, ok := tcpMuxCache[address]

		if !ok {
			tcpListener, err := net.ListenTCP("tcp", tcpAddr)
			if err != nil {
				log.Fatal(err)
			}

			tcpMux = webrtc.NewICETCPMux(nil, tcpListener, 8)
			tcpMuxCache[address] = tcpMux
		}

		settingEngine.SetICETCPMux(tcpMux)
	} else {
		// log.Println("No Mux TCP ports configured")
	}
}

func setupUDPMux(settingEngine webrtc.SettingEngine, isWhip bool, udpMuxCache map[int]*ice.MultiUDPMuxDefault, udpMuxOpts []ice.UDPMuxFromPortOption) {
	// Use UDP Mux port if set
	if udpMuxPort := getUDPMuxPort(isWhip); udpMuxPort != 0 {
		setUDPMuxPort(isWhip, udpMuxPort, udpMuxCache, udpMuxOpts, settingEngine)
	} else {
		// log.Println("No Mux UDP ports configured")
	}
}

func setupInterfaceFilter(settingEngine webrtc.SettingEngine, muxOpts []ice.UDPMuxFromPortOption) {
	interfaceFilter := func(i string) bool {
		return i == os.Getenv("INTERFACE_FILTER")
	}

	settingEngine.SetInterfaceFilter(interfaceFilter)
	muxOpts = append(muxOpts, ice.UDPMuxFromPortWithInterfaceFilter(interfaceFilter))
}

func getTCPMuxAddress() *net.TCPAddr {
	sharedAddress := os.Getenv("TCP_MUX_ADDRESS")

	if sharedAddress != "" {
		tcpAddr, err := net.ResolveTCPAddr("tcp", sharedAddress)

		if err != nil {
			log.Fatal(err)
		}

		return tcpAddr
	}

	return nil
}

func getUDPMuxPort(isWhip bool) int {
	sharedPort := os.Getenv("UDP_MUX_PORT")
	whipPort := os.Getenv("UDP_MUX_PORT_WHIP")
	whepPort := os.Getenv("UDP_MUX_PORT_WHEP")

	// Set for WHIP
	if isWhip && whipPort != "" {
		port, err := strconv.Atoi(whipPort)
		if err != nil {
			log.Fatal(err)
		}

		return port
	}

	// Set for WHEP
	if !isWhip && whepPort != "" {
		port, err := strconv.Atoi(whepPort)
		if err != nil {
			log.Fatal(err)
		}

		return port
	}

	// Set generalized
	if sharedPort != "" {
		port, err := strconv.Atoi(sharedPort)
		if err != nil {
			log.Fatal(err)
		}

		return port
	}

	// Do not use mux
	return 0
}

func setUDPMuxPort(isWhip bool, udpMuxPort int, udpMuxCache map[int]*ice.MultiUDPMuxDefault, udpMuxOpts []ice.UDPMuxFromPortOption, settingEngine webrtc.SettingEngine) {
	if isWhip {
		log.Println("Setting up WHIP UDP Mux to", udpMuxPort)
	} else {
		log.Println("Setting up WHEP UDP Mux to", udpMuxPort)
	}

	udpMux, ok := udpMuxCache[udpMuxPort]
	if !ok {
		// No Mux for current port, create new
		newUdpMux, err := ice.NewMultiUDPMuxFromPort(udpMuxPort, udpMuxOpts...)

		if err != nil {
			log.Fatal(err)
		}

		udpMuxCache[udpMuxPort] = newUdpMux
	}

	// Set to Mux on existing port
	settingEngine.SetICEUDPMux(udpMux)
}

func setupNAT(settingEngine webrtc.SettingEngine) {
	var (
		natIps []string
	)

	natICECandidateType := webrtc.ICECandidateTypeHost

	if os.Getenv("INCLUDE_PUBLIC_IP_IN_NAT_1_TO_1_IP") != "" {
		natIps = append(natIps, ip.GetPublicIp())
	}

	if os.Getenv("NAT_1_TO_1_IP") != "" {
		natIps = append(natIps, strings.Split(os.Getenv("NAT_1_TO_1_IP"), "|")...)
	}

	if os.Getenv("NAT_ICE_CANDIDATE_TYPE") != "srflx" {
		natICECandidateType = webrtc.ICECandidateTypeSrflx
	}

	if len(natIps) != 0 {
		settingEngine.SetNAT1To1IPs(natIps, natICECandidateType)
	}
}
