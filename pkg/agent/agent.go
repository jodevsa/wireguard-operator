package agent

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
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
