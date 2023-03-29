// Copyright (C) 2016, 2017 Nicolas Lamirault <nicolas.lamirault@gmail.com>

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package speedtest_client

import (
	"log"
	"sync"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

const (
	userAgent = "speedtest_exporter"
)

// Client defines the Speedtest client
type Client struct {
	Server           speedtest.Server
	SpeedtestClient  *speedtest.User
	AllServers       speedtest.Servers
	ClosestServers   speedtest.Servers
	mu               sync.Mutex
	lastResult       *map[string]float64
	resultValidUntil time.Time
	resultValidFor   time.Duration
}

func (client *Client) RefreshFastestServer() error {
	allServers, err := client.Server.Context.FetchServers(client.SpeedtestClient)
	if err != nil {
		log.Println("Could not get speedtest servers.")
		return err
	}

	findServer, err := allServers.FindServer([]int{})
	if err != nil {
		return err
	}

	if findServer[0].ID != client.Server.ID {
		log.Printf("Selected new speedtest server %s with ID %s\n", findServer[0].Name, findServer[0].ID)
		client.Server = *findServer[0]
	}
	return nil
}

// NewClient defines a new client for Speedtest
func NewClient(interval time.Duration) (*Client, error) {

	client := speedtest.New()
	user, _ := client.FetchUserInfo()

	log.Println("Retrieve configuration")

	log.Println("Retrieve all servers")
	var allServers speedtest.Servers
	var err error

	allServers, err = client.FetchServers(user)
	if err != nil {
		return nil, err
	}

	log.Println("Retrieve closest servers")

	var findServer, _ = allServers.FindServer([]int{})
	selectedServer := findServer[0]
	log.Printf("Selected server %s with ID %s\n", selectedServer.Name, selectedServer.ID)

	return &Client{
		Server:           *findServer[0],
		SpeedtestClient:  user,
		AllServers:       allServers,
		ClosestServers:   allServers,
		mu:               sync.Mutex{},
		lastResult:       nil,
		resultValidUntil: time.UnixMicro(0),
		resultValidFor:   interval,
	}, nil
}

func NewClientWithFixedId(interval time.Duration, serverId int) (*Client, error) {
	client, err := NewClient(interval)
	if err != nil {
		return nil, err
	}
	servers, err := (*client).AllServers.FindServer([]int{serverId})
	if err != nil || len(servers) == 0 {
		return nil, err
	}
	client.Server = *servers[0]
	return client, nil
}

func (client *Client) NetworkMetrics() map[string]float64 {
	client.mu.Lock()
	defer client.mu.Unlock()

	// Check if result is still valid
	if client.resultValidUntil.Unix() > time.Now().Unix() {
		validForSeconds := time.Until(client.resultValidUntil)
		log.Printf("Using cached result, still valid for %s.\n", validForSeconds)
		return *client.lastResult
	}

	// Refresh selected client
	client.RefreshFastestServer()

	// Start actual test
	result := map[string]float64{}

	server := client.Server

	var err error

	log.Println("Latency test")
	err = server.PingTest()
	if err == nil {
		log.Printf("Latency: %f ms\n", float64(server.Latency.Milliseconds()))
		log.Println("Download test")
		err = server.DownloadTest()
	}
	if err == nil {
		log.Printf("Download: %f Mbit/s\n", server.DLSpeed)
		log.Println("Upload test")
		err = server.UploadTest()
	}
	if err != nil {
		log.Println(err)
		result["download"] = 0
		result["upload"] = 0
		result["ping"] = 0
		return result
	}
	log.Printf("Upload: %f Mbit/s\n", server.ULSpeed)

	log.Printf("Speedtest Download: %v Mbps\n", server.DLSpeed)
	log.Printf("Speedtest Upload: %v Mbps\n", server.ULSpeed)

	log.Printf("Speedtest Latency: %v ms\n", server.Latency)
	result["download"] = server.DLSpeed
	result["upload"] = server.ULSpeed
	result["ping"] = float64(server.Latency.Milliseconds())

	speedtest.GlobalDataManager.Reset()
	log.Println("Speedtest finished")

	// Update cached result
	client.lastResult = &result
	client.resultValidUntil = time.Now().Add(client.resultValidFor)

	return result
}
