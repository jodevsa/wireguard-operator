package agent

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"

	"github.com/fsnotify/fsnotify"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
)

type State struct {
	Server           v1alpha1.Wireguard
	ServerPrivateKey string
	Peers            []v1alpha1.WireguardPeer
}

func IsStateValid(state State) error {

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

func OnStateChange(path string, logger logr.Logger, onFileChange func(State)) (func(), error) {

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
		err := IsStateValid(state)

		if err != nil {
			logger.Error(err, "State is not valid")
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
				logger.V(9).Info("Received a new event", "filename", event.Name, "operation", event.Op.String())
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {

					state, newHash, err := GetDesiredState(path)
					if err != nil {
						logger.Error(err, "unable to read or parse state")
						continue
					}

					if newHash == hash {
						logger.V(9).Info("Received a new event but state content did not change")
						continue
					}

					logger.V(9).Info("State content changed", "oldHash", hash, "newHash", newHash)
					hash = newHash

					err = IsStateValid(state)

					if err != nil {
						logger.Error(err, "State is not valid")
						continue
					}

					onFileChange(state)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Error(err, "watcher error")
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
