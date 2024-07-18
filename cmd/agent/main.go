package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-logr/stdr"
	"github.com/jodevsa/wireguard-operator/internal/iptables"
	"github.com/jodevsa/wireguard-operator/pkg/agent"
	"github.com/jodevsa/wireguard-operator/pkg/wireguard"
)

func main() {
	var configFilePath string
	var iface string
	var verbosity int
	var wgUserspaceImplementationFallback string
	var wireguardListenPort int
	var wgUseUserspaceImpl bool
	flag.StringVar(&configFilePath, "state", "./state.json", "The location of the file that states the desired state")
	flag.StringVar(&iface, "wg-iface", "wg0", "the wg device name. Default is wg0")
	flag.StringVar(&wgUserspaceImplementationFallback, "wg-userspace-implementation-fallback", "wireguard-go", "The userspace implementation of wireguard to fallback to")
	flag.IntVar(&wireguardListenPort, "wg-listen-port", 51820, "the UDP port wireguard is listening on")
	flag.IntVar(&verbosity, "v", 1, "the verbosity level")
	flag.BoolVar(&wgUseUserspaceImpl, "wg-use-userspace-implementation", false, "Use userspace implementation")
	flag.Parse()

	println(fmt.Sprintf(
		`	
               .:::::::::::::::::::::::::...::::::::::::::::::::.               
             .::::::::::::::::::::.:^7J5PBGY!^::::::::::::::::::::.             
            :::::::::::::::::::::~?J??5&@@@@@&G!~~~::::::::::::::::.      WG Agent Configuration      
           ::::::::::::::::::::::^7&@@@@@@@@@@@@&&&G^:::::::::::::::.     ------------------------------------------       
          .::::::::::::::::::::::!J#@@@@@@@BBBGPPG7:::::::::::::::::.     wg-iface: %s      
          .:::::::::::::::::::::^?Y5#@@@@@@5^...:::::::::::::::::::::     state: %s      
          .::::::::::::::::::::::..:!7Y#@@@@@#Y~:.::::::::::::::::::.     wg-listen-port: %d      
          .:::::::::::::::::::.:^!?JYYJ?JG&@@@@@#7::::::::::::::::::.     wg-use-userspace-implementation: %v      
          .:::::::::::::::::.^J#@@@@@@@@@&#B&@@@@@G:::::::::::::::::.     wg-userspace-implementation-fallback: %s           
          .:::::::::::::::::J@@@@@@@@@@@@@@@&G@@@@@J.:::::::::::::::.           
          .::::::::::::::::5@@@@@#?~~~7P@@@@@&B@@@@P.:::::::::::::::.           
          .:::::::::::::::^@@@@@P..::::.~@@@@B&@@@@!::::::::::::::::.           
          .:::::::::::::::~@@@@@J.::::::^@@@#&@@@@P:::::::::::::::::.           
          .::::::::::::::::B@@@@@P!^:.:~G&&&@@@@@5::::::::::::::::::.           
          .:::::::::::::::::G@@@@@&#BB&@@@@@@@@B~.::::::::::::::::::.           
          .::::::::::::::::..~G&&&@@@@@@@@@&&&&&P^.:::::::::::::::::.           
          .::::::::::::::.:~YGGY&@@@@@&GY7JB@@@@@@7:::::::::::::::::.           
          .::::::::::::::?&@@@B&@@@@#!:..:::~B@@@@@~::::::::::::::::.           
          .:::::::::::::J&#P5?5@@@@@:.::::::::&@@@@5.:::::::::::::::.           
          .:::::::::::::^:....J@@@@@~.::::::.^@@@@@5.::::::::::::::::           
          .::::::::::::::::::::&@@@@@Y~::::^J&@@@@&^::::::::::::::::.           
           ::::::::::::::::::::^B@@@@@@&##&@@@@@@#~:::::::::::::::::.           
            :::::::::::::::::::::7B@@@@@@@@@@@@#?::::::::::::::::::.            
             .::::::::::::::::::::.^7YGB##BGY7^:.:::::::::::::::::.             
               .:::::::::::::::::::::..::::..::::::::::::::::::..               
                  .....:...............................:.....                   

	/  \    /  \/  _____/       /  _  \   / ___\  ____   _____/  |_  
	\   \/\/   /   \  ___      /  /_\  \ / /_/  _/ __ \ /    \   __\ 
	 \        /\    \_\  \    /    |    \\___  /\  ___/|   |  |  |   
	  \__/\  /  \______  /    \____|__  /_____/  \___  |___|  |__|   
		   \/          \/             \/             \/     \/
`, iface, configFilePath, wireguardListenPort, wgUseUserspaceImpl, wgUserspaceImplementationFallback))

	stdr.SetVerbosity(verbosity)
	log := stdr.NewWithOptions(log.New(os.Stderr, "", log.LstdFlags), stdr.Options{LogCaller: stdr.All})
	log = log.WithName("agent")

	wg := wireguard.Wireguard{
		Logger:                            log.WithName("wireguard"),
		Iface:                             iface,
		ListenPort:                        wireguardListenPort,
		WgUserspaceImplementationFallback: wgUserspaceImplementationFallback,
		WgUseUserspaceImpl:                wgUseUserspaceImpl,
	}
	it := iptables.Iptables{
		Logger: log.WithName("iptables"),
	}

	close, err := agent.OnStateChange(configFilePath, log.WithName("onStateChange"), func(state agent.State) {
		log.Info("Received a new state")
		err := wg.Sync(state)
		if err != nil {
			log.Error(err, "Error while sycncing wireguard")
		}

		err = it.Sync(state)
		if err != nil {
			log.Error(err, "Error while syncing network policies")
		}

	})

	if err != nil {
		log.Error(err, "Error while watching changes")
		os.Exit(1)
	}

	defer close()

	httpLog := log.WithName("http")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		state, _, err := agent.GetDesiredState(configFilePath)

		if err != nil {
			httpLog.Error(err, "agent is not ready as it cannot get server state")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		err = agent.IsStateValid(state)

		if err != nil {
			httpLog.Error(err, "agent is not ready as server state not valid")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		err = wg.Sync(state)

		if err != nil {
			httpLog.Error(err, "agent is not ready as it cannot sync wireguard")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		httpLog.Info("agent is ready")

		w.WriteHeader(http.StatusOK)
	})
	http.ListenAndServe(":8080", nil)
}
