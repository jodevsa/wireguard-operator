package agent

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
)

type State struct {
	Server           v1alpha1.Wireguard
	ServerPrivateKey string
	Peers            []v1alpha1.WireguardPeer
}

func isStateValid(state State) error {

	if state.ServerPrivateKey == "" {
		return fmt.Errorf("server private key is not defined")
	}

	if len(state.ServerPrivateKey) != 44 {
		return fmt.Errorf("server private key should be of length 44")
	}

	if state.Server.Status.Address == "" {
		return fmt.Errorf("server address is not defined")
	}

	if state.Server.Status.Dns == "" {
		return fmt.Errorf("dns is not defined")
	}

	for i, peer := range state.Peers {
		if peer.Spec.Address == "" {
			return fmt.Errorf("peer with index %d does not have the address defined", i)
		}

		if peer.Spec.PublicKey == "" {
			return fmt.Errorf("peer with index %d does not have a public key defined", i)
		}
	}

	return nil
}

func OnStateChange(path string, onFileChange func(State)) (func(), error) {

	dir := filepath.Dir(path)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	close := func() {
		watcher.Close()
	}

	state, hash, err := GetDesiredState(path)

	if err == nil {
		err := isStateValid(state)

		if err != nil {
			log.Println("State is not valid: " + err.Error())
		} else {
			onFileChange(state)
		}
	}

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					if err != nil {
						log.Println(err)
					} else {
						state, newHash, err := GetDesiredState(path)

						if err != nil {
							log.Println(err)
						}

						if newHash == hash {
							continue
						}
						hash = newHash

						err = isStateValid(state)

						if err != nil {
							log.Println("State is not valid: " + err.Error())
						} else {
							onFileChange(state)
						}
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

func GetDesiredState(path string) (State, string, error) {
	var state State
	jsonFile, err := os.ReadFile(path)
	if err != nil {
		return State{}, "", err
	}
	err = json.Unmarshal(jsonFile, &state)
	if err != nil {
		return State{}, "", err
	}
	hash := md5.Sum(jsonFile)

	return state, hex.EncodeToString(hash[:]), nil
}
