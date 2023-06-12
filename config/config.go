package config

import (
	"errors"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"gopkg.in/ini.v1"
)

const (
	ERR_HOST_NOT_FOUND = "host not found in config"
)

type KeyPaths struct {
	PathHostCAPrivateKey string `ini:"host_ca_privkey"`
	PathHostCAPublicKey  string `ini:"host_ca_pubkey"`
	PathUserCAPrivateKey string `ini:"user_ca_privkey"`
	PathUserCAPublicKey  string `ini:"user_ca_pubkey"`
}

type Keys struct {
	HostCAPrivateKey interface{}
	HostCAPublicKey  ssh.PublicKey
	UserCAPrivateKey interface{}
	UserCAPublicKey  ssh.PublicKey
}

type HostGroup struct {
	KeyPaths
	Keys
	Name  string
	Hosts map[string]string
}

type Config struct {
	HostGroups []HostGroup
}

func LoadConfig(path string) (Config, error) {
	var conf Config
	var defKeyPaths KeyPaths

	cfg, err := ini.Load(path)
	if err != nil {
		return conf, err
	}

	if err := cfg.MapTo(&defKeyPaths); err != nil {
		return conf, err
	}

	// ini doesn't support mapping to map[string]string, do it manually
	for _, hostgroup := range cfg.Sections() {
		if hostgroup.Name() == ini.DefaultSection {
			continue
		}

		// prefill with global values
		keys := &KeyPaths{
			PathHostCAPrivateKey: defKeyPaths.PathHostCAPrivateKey,
			PathHostCAPublicKey:  defKeyPaths.PathHostCAPublicKey,
			PathUserCAPrivateKey: defKeyPaths.PathUserCAPrivateKey,
			PathUserCAPublicKey:  defKeyPaths.PathUserCAPublicKey,
		}

		if err := hostgroup.MapTo(keys); err != nil {
			return conf, err
		}

		hg := &HostGroup{
			KeyPaths: *keys,
			Name:     hostgroup.Name(),
			Hosts:    hostgroup.KeysHash(),
		}

		hosts := make(map[string]string)
		for key, val := range hostgroup.KeysHash() {
			if key == "host_ca_privkey" || key == "host_ca_pubkey" ||
				key == "user_ca_privkey" || key == "user_ca_pubkey" {
				continue
			}

			hosts[key] = val
		}

		hg.Hosts = hosts

		if hg.Name != ini.DefaultSection &&
			(hg.PathHostCAPrivateKey == "" ||
				hg.PathHostCAPublicKey == "" ||
				hg.PathUserCAPrivateKey == "" ||
				hg.PathUserCAPublicKey == "") {
			return conf, errors.New("missing key in hostgroup " + hg.Name)
		}

		conf.HostGroups = append(conf.HostGroups, *hg)
	}

	if loadKeys(&conf) != nil {
		return conf, errors.New("could not open and parse keys")
	}

	return conf, nil
}

func loadKeys(conf *Config) error {
	var uniqPubKeys = make(map[string]ssh.PublicKey)
	var uniqPrivKeys = make(map[string]interface{})

	for i, group := range conf.HostGroups {
		for _, path := range []string{group.PathHostCAPublicKey, group.PathUserCAPublicKey} {
			if _, ok := uniqPubKeys[path]; ok {
				continue
			}

			pk, err := parsePublicKeyFile(path)
			if err != nil {
				return err
			}

			uniqPubKeys[path] = pk
		}

		for _, path := range []string{group.PathHostCAPrivateKey, group.PathUserCAPrivateKey} {
			if _, ok := uniqPrivKeys[path]; ok {
				continue
			}

			pk, err := parsePrivateKeyFile(path)
			if err != nil {
				return err
			}

			uniqPrivKeys[path] = pk
		}

		conf.HostGroups[i].Keys.HostCAPublicKey = uniqPubKeys[group.PathHostCAPublicKey]
		conf.HostGroups[i].Keys.UserCAPublicKey = uniqPubKeys[group.PathUserCAPublicKey]
		conf.HostGroups[i].Keys.HostCAPrivateKey = uniqPrivKeys[group.PathHostCAPrivateKey]
		conf.HostGroups[i].Keys.UserCAPrivateKey = uniqPrivKeys[group.PathUserCAPrivateKey]
	}

	return nil
}

func parsePublicKeyFile(path string) (ssh.PublicKey, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pk, _, _, _, err := ssh.ParseAuthorizedKey(content)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

func parsePrivateKeyFile(path string) (interface{}, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pk, err := ssh.ParseRawPrivateKey(content)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// GetKeys returns the host and user CA keys for the given host.
// Prefix matching using wildcards is supported.
func (c Config) GetKeys(host string) (Keys, error) {
	for _, confGroup := range c.HostGroups {
		for cHost := range confGroup.Hosts {
			if matchesHost(host, cHost) {
				return confGroup.Keys, nil
			}
		}
	}

	return Keys{}, errors.New(ERR_HOST_NOT_FOUND)
}

// matchesHost determines whether the given host matches host2.
// host2 may be a wildcard domain in the form of
//
//	*.example.com
//
// which matches any subdomain of example.com, but not example.com itself.
func matchesHost(host, host2 string) bool {
	if strings.HasPrefix(host2, "*.") {
		root, _ := strings.CutPrefix(host2, "*.")

		return strings.HasSuffix(host, root) && host != root
	} else {
		return host == host2
	}
}

func (c Config) GetMotleyCueURL(host string) (string, error) {
	for _, confGroup := range c.HostGroups {
		for cHost, cCA := range confGroup.Hosts {
			if matchesHost(host, cHost) {
				return cCA, nil
			}
		}
	}

	return "", errors.New(ERR_HOST_NOT_FOUND)
}
