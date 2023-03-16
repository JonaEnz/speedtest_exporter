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
	"fmt"

	"github.com/showwin/speedtest-go/speedtest"
)

const (
	userAgent = "speedtest_exporter"
)

// Client defines the Speedtest client
type Client struct {
	Server          speedtest.Server
	SpeedtestClient *speedtest.User
	AllServers      speedtest.Servers
	ClosestServers  speedtest.Servers
}

// NewClient defines a new client for Speedtest
func NewClient() (*Client, error) {

	client := speedtest.New()
	user, _ := client.FetchUserInfo()

	fmt.Println("Retrieve configuration")

	fmt.Println("Retrieve all servers")
	var allServers speedtest.Servers
	var err error

	allServers, err = client.FetchServers(user)
	if err != nil {
		return nil, err
	}

	fmt.Println("Retrieve closest servers")

	var findServer, _ = allServers.FindServer([]int{})

	return &Client{
		Server:          *findServer[0],
		SpeedtestClient: user,
		AllServers:      allServers,
		ClosestServers:  allServers,
	}, nil
}

func NewClientWithFixedId(serverId int) (*Client, error) {
	client, err := NewClient()
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
	result := map[string]float64{}

	server := client.Server

	var err error

	fmt.Println("Latency test")
	err = server.PingTest()
	if err == nil {
		fmt.Printf("Latency: %f ms\n", float64(server.Latency.Milliseconds()))
		fmt.Println("Download test")
		server.DownloadTest()
	}
	if err == nil {
		fmt.Printf("Download: %f Mbit/s\n", server.DLSpeed)
		fmt.Println("Upload test")
		server.UploadTest()
	}
	if err != nil {
		fmt.Println(err)
		result["download"] = 0
		result["upload"] = 0
		result["ping"] = 0
		return result
	}
	fmt.Printf("Upload: %f Mbit/s\n", server.ULSpeed)

	fmt.Printf("Speedtest Download: %v Mbps\n", server.DLSpeed)
	fmt.Printf("Speedtest Upload: %v Mbps\n", server.ULSpeed)

	fmt.Printf("Speedtest Latency: %v ms\n", server.Latency)
	result["download"] = server.DLSpeed
	result["upload"] = server.ULSpeed
	result["ping"] = float64(server.Latency.Milliseconds())
	fmt.Println("Speedtest finished")
	speedtest.GlobalDataManager.Reset()
	return result
}
