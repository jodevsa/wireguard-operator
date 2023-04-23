package agent

import (
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	"log"
	"os"
	"path/filepath"
)

type State struct {
	Server           v1alpha1.Wireguard
	ServerPrivateKey string
	Peers            []v1alpha1.WireguardPeer
}

func OnStateChange(path string, onFileChange func(State)) (func(), error) {

	dir := filepath.Dir(path)
	filename := filepath.Base(path)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	close := func() {
		watcher.Close()
	}

	state, err := GetDesiredState(path)

	if err == nil {
		onFileChange(state)
	}

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				print(event.Name)
				print(event.Op)
				print(event.String())
				if (event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) && event.Name == filename {
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

	err = watcher.Add(dir)
	if err != nil {
		println(".....")
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
