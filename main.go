package main

import (
	"io"
	"log"
	"net"
	"os"

	"gopkg.in/yaml.v2"
)

type Mapping struct {
	Type string `yaml:"type"`
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

type Config struct {
	Mappings []Mapping `yaml:"mappings"`
}

func main() {
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	for _, mapping := range config.Mappings {
		go startForwarder(mapping)
	}

	// Keep the main goroutine running
	select {}
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func startForwarder(mapping Mapping) {
	switch mapping.Type {
	case "tcp":
		startTCPForwarder(mapping)
	case "udp":
		startUDPForwarder(mapping)
	default:
		log.Printf("Unknown protocol type: %s", mapping.Type)
	}
}

func startTCPForwarder(mapping Mapping) {
	listener, err := net.Listen("tcp", mapping.From)
	if err != nil {
		log.Printf("Error starting TCP listener on %s: %v", mapping.From, err)
		return
	}
	defer listener.Close()

	log.Printf("TCP forwarder listening on %s, forwarding to %s", mapping.From, mapping.To)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go handleTCPConnection(clientConn, mapping.To)
	}
}

func handleTCPConnection(clientConn net.Conn, target string) {
	defer clientConn.Close()

	targetConn, err := net.Dial("tcp", target)
	if err != nil {
		log.Printf("Error connecting to target %s: %v", target, err)
		return
	}
	defer targetConn.Close()

	go io.Copy(targetConn, clientConn)
	io.Copy(clientConn, targetConn)
}

func startUDPForwarder(mapping Mapping) {
	addr, err := net.ResolveUDPAddr("udp", mapping.From)
	if err != nil {
		log.Printf("Error resolving UDP address %s: %v", mapping.From, err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("Error starting UDP listener on %s: %v", mapping.From, err)
		return
	}
	defer conn.Close()

	log.Printf("UDP forwarder listening on %s, forwarding to %s", mapping.From, mapping.To)

	buffer := make([]byte, 65507) // Max UDP packet size

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading UDP packet: %v", err)
			continue
		}

		go handleUDPPacket(conn, remoteAddr, buffer[:n], mapping.To)
	}
}

func handleUDPPacket(conn *net.UDPConn, remoteAddr *net.UDPAddr, data []byte, target string) {
	targetAddr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		log.Printf("Error resolving target UDP address %s: %v", target, err)
		return
	}

	targetConn, err := net.DialUDP("udp", nil, targetAddr)
	if err != nil {
		log.Printf("Error connecting to target %s: %v", target, err)
		return
	}
	defer targetConn.Close()

	_, err = targetConn.Write(data)
	if err != nil {
		log.Printf("Error forwarding UDP packet to %s: %v", target, err)
		return
	}

	buffer := make([]byte, 65507)
	n, _, err := targetConn.ReadFromUDP(buffer)
	if err != nil {
		log.Printf("Error reading UDP response from %s: %v", target, err)
		return
	}

	_, err = conn.WriteToUDP(buffer[:n], remoteAddr)
	if err != nil {
		log.Printf("Error sending UDP response back to %s: %v", remoteAddr, err)
	}
}
