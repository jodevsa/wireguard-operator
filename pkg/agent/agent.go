package agent

import (
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	"log"
	"os"
)

type State struct {
	Server           v1alpha1.Wireguard
	ServerPrivateKey string
	Peers            []v1alpha1.WireguardPeer
}

func OnStateChange(path string, onFileChange func(State)) (func(), error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	close := func() {
		watcher.Close()
	}

	state, err := GetDesiredState(path)
	if err != nil {
		log.Println(err)
	}
	onFileChange(state)

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Has(fsnotify.Write) {
					if err != nil {
						log.Println(err)
					} else {
						state, err := GetDesiredState(path)
						if err != nil {
							log.Println(err)
						}

						onFileChange(state)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		return close, err
	}
	return close, nil
}

func GetDesiredState(path string) (State, error) {
	var state State
	jsonFile, err := os.ReadFile(path)
	if err != nil {
		return State{}, err
	}
	err = json.Unmarshal(jsonFile, &state)
	if err != nil {
		return State{}, err
	}
	return state, nil
}
