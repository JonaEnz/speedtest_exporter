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

package speedtest

import (
	"github.com/prometheus/common/log"
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
func NewClient(configURL string, serversURL string) (*Client, error) {
	log.Debugf("New Speedtest client %s %s", configURL, serversURL)

	user, _ := speedtest.FetchUserInfo()

	log.Debug("Retrieve configuration")

	log.Debugf("Retrieve all servers")
	var allServers speedtest.Servers
	var err error

	allServers, err = speedtest.FetchServers(user)
	if err != nil {
		return nil, err
	}

	log.Debugf("Retrieve closest servers")

	//closestServers := speedtest.ByDistance{Servers: allServers}

	// log.Infof("Closest Servers: %s", closestServers)
	// testServer := stClient.GetFastestServer(closestServers)
	//log.Infof("Test server: %s", testServer)

	var findServer, _ = allServers.FindServer([]int{})

	return &Client{
		Server:          *findServer[0],
		SpeedtestClient: user,
		AllServers:      allServers,
		ClosestServers:  allServers,
	}, nil
}

func (client *Client) NetworkMetrics() map[string]float64 {
	result := map[string]float64{}

	server := client.Server

	var err error

	log.Info("Latency test")
	err = server.PingTest()
	if err == nil {
		log.Infof("Latency: %f ms", float64(server.Latency.Milliseconds()))
		log.Info("Download test")
		server.DownloadTest(false)
	}
	if err == nil {
		log.Infof("Download: %f Mbit/s", server.DLSpeed)
		log.Info("Upload test")
		server.UploadTest(false)
	}
	if err != nil {
		log.Error(err)
		result["download"] = 0
		result["upload"] = 0
		result["ping"] = 0
		return result
	}
	log.Infof("Upload: %f Mbit/s", server.ULSpeed)

	// tester := tests.NewTester(client.SpeedtestClient, tests.DefaultDLSizes, tests.DefaultULSizes, false, false)
	// downloadMbps := tester.Download(client.Server)
	log.Infof("Speedtest Download: %v Mbps", server.DLSpeed)
	//uploadMbps := tester.Upload(client.Server)
	log.Infof("Speedtest Upload: %v Mbps", server.ULSpeed)

	//ping, err := client.SpeedtestClient.GetLatency(client.Server, client.SpeedtestClient.GetLatencyURL//(client.Server))
	//if err != nil {
	//	log.Fatal(err)
	//}

	log.Infof("Speedtest Latency: %v ms", server.Latency)
	result["download"] = server.DLSpeed
	result["upload"] = server.ULSpeed
	result["ping"] = float64(server.Latency.Milliseconds())
	log.Infof("Speedtest finished")
	return result
}
