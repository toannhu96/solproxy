package main

import (
	"gosol/passthrough"
	"gosol/solana/handle_solana_01"
	"gosol/solana/handle_solana_admin"
	"gosol/solana/handle_solana_info"
	"gosol/solana/handle_solana_raw"
	"runtime"
	"strings"

	"github.com/slawomir-pryczek/handler_socket2"
	"github.com/slawomir-pryczek/handler_socket2/handle_echo"
	"github.com/slawomir-pryczek/handler_socket2/handle_profiler"

	"fmt"
	"os"
)

func _read_node_config() {

	fmt.Println("\nReading node config...")
	nodes := (handler_socket2.Config().GetRawData("SOL_NODES", "")).([]interface{})
	if len(nodes) <= 0 {
		fmt.Println("ERROR: No nodes defined, please define at least one solana node to connect to")
		os.Exit(10)
		return
	}

	for _, v := range nodes {
		handle_solana_admin.NodeRegisterFromConfig(v.(map[string]interface{}))
	}
	fmt.Println("")
}

func main() {
	_read_node_config()

	num_cpu := runtime.NumCPU() * 2
	runtime.GOMAXPROCS(num_cpu) // register handlers
	handlers := []handler_socket2.ActionHandler{}
	handlers = append(handlers, &handle_echo.HandleEcho{})
	handlers = append(handlers, &handle_profiler.HandleProfiler{})
	handlers = append(handlers, &handle_solana_raw.Handle_solana_raw{})
	handlers = append(handlers, &handle_solana_01.Handle_solana_01{})
	handlers = append(handlers, &handle_solana_info.Handle_solana_info{})
	handlers = append(handlers, &handle_passthrough.Handle_passthrough{})
	handlers = append(handlers, &handle_solana_admin.Handle_solana_admin{})

	if len(handler_socket2.Config().Get("RUN_SERVICES", "")) > 0 && handler_socket2.Config().Get("RUN_SERVICES", "") != "*" {
		_h_modified := []handler_socket2.ActionHandler{}
		_tmp := strings.Split(handler_socket2.Config().Get("RUN_SERVICES", ""), ",")
		supported := make(map[string]bool)
		for _, v := range _tmp {
			supported[strings.Trim(v, "\r\n \t")] = true
		}

		for _, v := range handlers {
			should_enable := false
			for _, action := range v.GetActions() {
				if supported[action] {
					should_enable = true
					break
				}
			}

			if should_enable {
				_h_modified = append(_h_modified, v)
			}
		}

		handlers = _h_modified
	}

	// start the server
	handler_socket2.RegisterHandler(handlers...)
	handler_socket2.StartServer(strings.Split(handler_socket2.Config().Get("BIND_TO", ""), ","))
}
