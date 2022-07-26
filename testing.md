# Development

See Building from source docs: [https://github.com/BishopFox/sliver/wiki/Compile-From-Source](https://github.com/BishopFox/sliver/wiki/Compile-From-Source)

## Environment

- windows host
- wsl installed
- golang, build-essential, net-tools, protobuf-compiler, zip
- clone fork of sliver repo [`git@github.com](mailto:git@github.com):rawk77/sliver.git`
- setup `netsh portproxy` to `8888` for WSL
- Install protobuf go plugins

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

# Build

```bash
cd sliver
./go-assets.sh  # on first setup only
make
```

# Portscan

Code locations for `Portscan` command

client/command/network/portscan.go

```go
package network

/*
	Sliver Implant Framework
	Copyright (C) 2021  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bishopfox/sliver/client/command/settings"
	"github.com/bishopfox/sliver/client/console"
	"github.com/bishopfox/sliver/protobuf/clientpb"
	"github.com/bishopfox/sliver/protobuf/sliverpb"
	"github.com/desertbit/grumble"
	"github.com/jedib0t/go-pretty/v6/table"
	"google.golang.org/protobuf/proto"
)

// PortscanCmd - Display network interfaces on the remote system
func PortscanCmd(ctx *grumble.Context, con *console.SliverConsoleClient) {
	session, beacon := con.ActiveTarget.GetInteractive()
	if session == nil && beacon == nil {
		return
	}
	portscan, err := con.Rpc.Portscan(context.Background(), &sliverpb.PortscanReq{
		Request: con.ActiveTarget.Request(ctx),
	})
	if err != nil {
		con.PrintErrorf("%s\n", err)
		return
	}
	all := ctx.Flags.Bool("all")
	if portscan.Response != nil && portscan.Response.Async {
		con.AddBeaconCallback(portscan.Response.TaskID, func(task *clientpb.BeaconTask) {
			err = proto.Unmarshal(task.Response, portscan)
			if err != nil {
				con.PrintErrorf("Failed to decode response %s\n", err)
				return
			}
			PrintPortscan(portscan, all, con)
		})
		con.PrintAsyncResponse(portscan.Response)
	} else {
		PrintPortscan(portscan, all, con)
	}
}

// PrintPortscan - Print the portscan response
func PrintPortscan(portscan *sliverpb.Portscan, all bool, con *console.SliverConsoleClient) {
	var err error
	interfaces := portscan.NetInterfaces
	sort.Slice(interfaces, func(i, j int) bool {
		return interfaces[i].Index < interfaces[j].Index
	})
	hidden := 0
	for index, iface := range interfaces {
		tw := table.NewWriter()
		tw.SetStyle(settings.GetTableWithBordersStyle(con))
		tw.SetTitle(fmt.Sprintf(console.Bold+"%s"+console.Normal, iface.Name))
		tw.SetColumnConfigs([]table.ColumnConfig{
			{Name: "#", AutoMerge: true},
			{Name: "IP Address", AutoMerge: true},
			{Name: "MAC Address", AutoMerge: true},
		})
		rowConfig := table.RowConfig{AutoMerge: true}
		tw.AppendHeader(table.Row{"#", "IP Addresses", "MAC Address"}, rowConfig)
		macAddress := ""
		if 0 < len(iface.MAC) {
			macAddress = iface.MAC
		}
		ips := []string{}
		for _, ip := range iface.IPAddresses {
			// Try to find local IPs and colorize them
			subnet := -1
			if strings.Contains(ip, "/") {
				parts := strings.Split(ip, "/")
				subnetStr := parts[len(parts)-1]
				subnet, err = strconv.Atoi(subnetStr)
				if err != nil {
					subnet = -1
				}
			}
			if 0 < subnet && subnet <= 32 && !isLoopback(ip) {
				ips = append(ips, fmt.Sprintf(console.Bold+console.Green+"%s"+console.Normal, ip))
			} else if all {
				ips = append(ips, fmt.Sprintf("%s", ip))
			}
		}
		if !all && len(ips) < 1 {
			hidden++
			continue
		}
		if 0 < len(ips) {
			for _, ip := range ips {
				tw.AppendRow(table.Row{iface.Index, ip, macAddress}, rowConfig)
			}
		} else {
			tw.AppendRow(table.Row{iface.Index, " ", macAddress}, rowConfig)
		}
		con.Printf("%s\n", tw.Render())
		if index+1 < len(interfaces) {
			con.Println()
		}
	}
	if !all {
		con.Printf("%d adapters not shown.\n", hidden)
	}
}
```

client/command/network/README.md

```markdown
Network related command implementations such as `netstat` and `ifconfig` and `portscan`
```

client/command/tasks/fetch.go - line 419

```go
case sliverpb.MsgPortscanReq:
		portscan := &sliverpb.Portscan{}
		err := proto.Unmarshal(task.Response, portscan)
		if err != nil {
			con.PrintErrorf("Failed to decode task response: %s\n", err)
			return
		}
		network.PrintPortscan(portscan, true, con)
```

client/constants/constants.go - line 180

```go
PortscanStr = "portscan"
```

client/command/commands.go - line 1881

```go
con.App.AddCommand(&grumble.Command{
		Name:     consts.PortscanStr,
		Help:     "Scan the network for open ports",
		LongHelp: help.GetHelpFor([]string{consts.PortscanStr}),
		Flags: func(f *grumble.Flags) {
			f.Bool("A", "all", false, "show all network adapters (default only shows IPv4)")

			f.Int("t", "timeout", defaultTimeout, "command timeout in seconds")
		},
		Run: func(ctx *grumble.Context) error {
			con.Println()
			network.PortscanCmd(ctx, con)
			con.Println()
			return nil
		},
		HelpGroup: consts.SliverHelpGroup,
	})
```

implant/sliver/handlers/handlers_darwin.go - line 46

```go
pb.MsgPortscanReq:  portscanHandler,
```

implant/sliver/handlers/handlers_linux.go - line 44

```go
sliverpb.MsgPortscanReq:  portscanHandler,
```

implant/sliver/handlers/handlers_windows.go - line 84

```go
sliverpb.MsgPortscanReq:            portscanHandler,
```

implant/sliver/handlers/rpc-handlers.go - line 176

```go
func portscanHandler(_ []byte, resp RPCResponse) {
	interfaces := portscan()
	// {{if .Config.Debug}}
	log.Printf("network interfaces: %#v", interfaces)
	// {{end}}
	data, err := proto.Marshal(interfaces)
	resp(data, err)
}

func portscan() *sliverpb.Portscan {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	interfaces := &sliverpb.Portscan{
		NetInterfaces: []*sliverpb.NetInterface{},
	}
	for _, iface := range netInterfaces {
		netIface := &sliverpb.NetInterface{
			Index: int32(iface.Index),
			Name:  iface.Name,
		}
		if iface.HardwareAddr != nil {
			netIface.MAC = iface.HardwareAddr.String()
		}
		addresses, err := iface.Addrs()
		if err == nil {
			for _, address := range addresses {
				netIface.IPAddresses = append(netIface.IPAddresses, address.String())
			}
		}
		interfaces.NetInterfaces = append(interfaces.NetInterfaces, netIface)
	}
	return interfaces
}
```

protobuf/sliverpb/constants.go - line 288

```go
	// MsgPortscanReq - Portscan (network interface config) request
	MsgPortscanReq
	// MsgPortscan - Portscan response
	MsgPortscan
```

protobuf/sliverpb/constants.go - line 506

```go
case *PortscanReq:
		return MsgPortscanReq
	case *Portscan:
		return MsgPortscan
```

protobuf/sliverpb/sliver.proto - line 144

```go
// PortscanReq - Request the implant to list network interfaces
message PortscanReq {
  commonpb.Request Request = 9;
}

message Portscan {
  repeated NetInterface NetInterfaces = 1;

  commonpb.Response Response = 9;
}
```

protobuf/rpcpb/services.proto - line 94

```go
rpc Portscan(sliverpb.PortscanReq) returns (sliverpb.Portscan);
```

server/rpc/rpc-net.go - line 48

```go
// Portscan - Get remote interface configurations
func (rpc *Server) Portscan(ctx context.Context, req *sliverpb.PortscanReq) (*sliverpb.Portscan, error) {
	resp := &sliverpb.Portscan{Response: &commonpb.Response{}}
	err := rpc.GenericHandler(req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
```
