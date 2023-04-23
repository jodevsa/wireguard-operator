package main

import (
	"flag"
	"fmt"
	"github.com/jodevsa/wireguard-operator/internal/iptables"
	"github.com/jodevsa/wireguard-operator/pkg/agent"
	"github.com/jodevsa/wireguard-operator/pkg/wireguard"
	"log"
)

func main() {
	var configFilePath string
	var iface string
	flag.StringVar(&configFilePath, "state", "./state.json", "The location of the file that states the desired state")
	flag.StringVar(&iface, "iface", "wg0", "the wg device name. Default is wg0")
	flag.Parse()

	println(
		`
 __      __  ________        _____     ____                __    
/  \    /  \/  _____/       /  _  \   / ___\  ____   _____/  |_  
\   \/\/   /   \  ___      /  /_\  \ / /_/  _/ __ \ /    \   __\ 
 \        /\    \_\  \    /    |    \\___  /\  ___/|   |  |  |   
  \__/\  /  \______  /    \____|__  /_____/  \___  |___|  |__|   
       \/          \/             \/             \/     \/
`)

	flag.VisitAll(func(f *flag.Flag) {
		log.Println(fmt.Sprintf("parameters - %s: %s\n", f.Name, f.Value))
	})

	close, err := agent.OnStateChange(configFilePath, func(state agent.State) {

		log.Println("Syncing wireguard")
		err := wireguard.Sync(state, iface)
		if err != nil {
			log.Println(err)
		}

		log.Println("Syncing iptables rules")
		err = iptables.Sync(state)
		if err != nil {
			log.Println(err)
		}
	})

	if err != nil {
		log.Fatal(err)
	}

	defer close()

	// Block main goroutine forever.
	<-make(chan struct{})
}
