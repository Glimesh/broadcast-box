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

	"github.com/glimesh/broadcast-box/internal/environment"
	"github.com/glimesh/broadcast-box/internal/ip"
)

func getSettingEngine(isWHIP bool, tcpMuxCache map[string]ice.TCPMux, udpMuxCache map[int]*ice.MultiUDPMuxDefault) (settingEngine webrtc.SettingEngine) {
	var (
		udpMuxOpts []ice.UDPMuxFromPortOption
	)

	setupNetworkTypes()
	setupNAT(&settingEngine)
	setupInterfaceFilter(&settingEngine, &udpMuxOpts)
	setupUDPMux(&settingEngine, isWHIP, udpMuxCache, udpMuxOpts)
	setupTCPMux(&settingEngine, tcpMuxCache)

	settingEngine.SetDTLSEllipticCurves(elliptic.X25519, elliptic.P384, elliptic.P256)
	settingEngine.SetNetworkTypes(setupNetworkTypes())
	settingEngine.DisableSRTCPReplayProtection(true)
	settingEngine.DisableSRTPReplayProtection(true)
	settingEngine.SetIncludeLoopbackCandidate(os.Getenv(environment.IncludeLoopbackCandidate) != "")

	return
}

func setupNetworkTypes() []webrtc.NetworkType {
	networkTypesEnv := os.Getenv(environment.NetworkTypes)
	tcpMuxForce := os.Getenv(environment.TCPMuxForce)

	networkTypes := []webrtc.NetworkType{}
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
		networkTypes = append(networkTypes, []webrtc.NetworkType{webrtc.NetworkTypeUDP4, webrtc.NetworkTypeUDP6}...)
	}

	return networkTypes
}

func setupTCPMux(settingEngine *webrtc.SettingEngine, tcpMuxCache map[string]ice.TCPMux) {
	// Use TCP Mux port if set
	if tcpAddr := getTCPMuxAddress(); tcpAddr != nil {
		address := os.Getenv(environment.TCPMuxAddress)
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
	}
}

func setupUDPMux(settingEngine *webrtc.SettingEngine, isWHIP bool, udpMuxCache map[int]*ice.MultiUDPMuxDefault, udpMuxOpts []ice.UDPMuxFromPortOption) {
	// Use UDP Mux port if set
	if udpMuxPort := getUDPMuxPort(isWHIP); udpMuxPort != 0 {
		setUDPMuxPort(isWHIP, udpMuxPort, udpMuxCache, udpMuxOpts, settingEngine)
	}
}

func setupInterfaceFilter(settingEngine *webrtc.SettingEngine, muxOpts *[]ice.UDPMuxFromPortOption) {
	filter := os.Getenv(environment.InterfaceFilter)

	if filter != "" {
		interfaceFilter := func(i string) bool {
			return i == filter
		}

		settingEngine.SetInterfaceFilter(interfaceFilter)
		*muxOpts = append(*muxOpts, ice.UDPMuxFromPortWithInterfaceFilter(interfaceFilter))
	}
}

func getTCPMuxAddress() *net.TCPAddr {
	sharedAddress := os.Getenv(environment.TCPMuxAddress)

	if sharedAddress != "" {
		tcpAddr, err := net.ResolveTCPAddr("tcp", sharedAddress)

		if err != nil {
			log.Fatal(err)
		}

		return tcpAddr
	}

	return nil
}

func getUDPMuxPort(isWHIP bool) int {
	sharedPort := os.Getenv(environment.UDPMuxPort)
	whipPort := os.Getenv(environment.UDPMuxPortWHIP)
	whepPort := os.Getenv(environment.UDPMuxPortWHEP)

	// Set for WHIP
	if isWHIP && whipPort != "" {
		port, err := strconv.Atoi(whipPort)
		if err != nil {
			log.Fatal(err)
		}

		return port
	}

	// Set for WHEP
	if !isWHIP && whepPort != "" {
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

func setUDPMuxPort(isWHIP bool, udpMuxPort int, udpMuxCache map[int]*ice.MultiUDPMuxDefault, udpMuxOpts []ice.UDPMuxFromPortOption, settingEngine *webrtc.SettingEngine) {
	if isWHIP {
		log.Println("Setting up WHIP UDP Mux to", udpMuxPort)
	} else {
		log.Println("Setting up WHEP UDP Mux to", udpMuxPort)
	}

	udpMux, ok := udpMuxCache[udpMuxPort]

	if !ok {
		// No Mux for current port, create new
		newUDPMux, err := ice.NewMultiUDPMuxFromPort(udpMuxPort, udpMuxOpts...)

		if err != nil {
			log.Fatal(err)
		}

		udpMuxCache[udpMuxPort] = newUDPMux
		udpMux = newUDPMux
	}

	// Set to Mux on existing port
	settingEngine.SetICEUDPMux(udpMux)
}

func setupNAT(settingEngine *webrtc.SettingEngine) {
	var (
		natIps []string
	)

	natICECandidateType := webrtc.ICECandidateTypeHost

	if os.Getenv(environment.IncludePublicIPInNAT1To1IP) != "" {
		natIps = append(natIps, ip.GetPublicIP())
	}

	if os.Getenv(environment.NAT1To1IP) != "" {
		natIps = append(natIps, strings.Split(os.Getenv(environment.NAT1To1IP), "|")...)
	}

	if os.Getenv(environment.NATICECandidateType) == "srflx" {
		natICECandidateType = webrtc.ICECandidateTypeSrflx
	}

	if len(natIps) != 0 {
		if err := settingEngine.SetICEAddressRewriteRules(webrtc.ICEAddressRewriteRule{
			External:        natIps,
			AsCandidateType: natICECandidateType,
			Mode:            webrtc.ICEAddressRewriteAppend,
		}); err != nil {
			log.Fatal("Configuration error: INCLUDE_PUBLIC_IP_IN_NAT_1_TO_1_IP:", err)
		}

	}
}
