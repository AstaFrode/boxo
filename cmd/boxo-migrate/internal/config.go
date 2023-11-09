package migrate

import (
	"encoding/json"
	"fmt"
	"io"
)

type Config struct {
	ImportPaths map[string]string
	Modules     []string
}

var DefaultConfig = Config{
	ImportPaths: map[string]string{
		"github.com/ipfs/go-bitswap":                     "github.com/AstaFrode/boxo/bitswap",
		"github.com/ipfs/go-ipfs-files":                  "github.com/AstaFrode/boxo/files",
		"github.com/ipfs/tar-utils":                      "github.com/AstaFrode/boxo/tar",
		"github.com/ipfs/interface-go-ipfs-core":         "github.com/AstaFrode/boxo/coreiface",
		"github.com/ipfs/go-unixfs":                      "github.com/AstaFrode/boxo/ipld/unixfs",
		"github.com/ipfs/go-pinning-service-http-client": "github.com/AstaFrode/boxo/pinning/remote/client",
		"github.com/ipfs/go-path":                        "github.com/AstaFrode/boxo/path",
		"github.com/ipfs/go-namesys":                     "github.com/AstaFrode/boxo/namesys",
		"github.com/ipfs/go-mfs":                         "github.com/AstaFrode/boxo/mfs",
		"github.com/ipfs/go-ipfs-provider":               "github.com/AstaFrode/boxo/provider",
		"github.com/ipfs/go-ipfs-pinner":                 "github.com/AstaFrode/boxo/pinning/pinner",
		"github.com/ipfs/go-ipfs-keystore":               "github.com/AstaFrode/boxo/keystore",
		"github.com/ipfs/go-filestore":                   "github.com/AstaFrode/boxo/filestore",
		"github.com/ipfs/go-ipns":                        "github.com/AstaFrode/boxo/ipns",
		"github.com/ipfs/go-blockservice":                "github.com/AstaFrode/boxo/blockservice",
		"github.com/ipfs/go-ipfs-chunker":                "github.com/AstaFrode/boxo/chunker",
		"github.com/ipfs/go-fetcher":                     "github.com/AstaFrode/boxo/fetcher",
		"github.com/ipfs/go-ipfs-blockstore":             "github.com/AstaFrode/boxo/blockstore",
		"github.com/ipfs/go-ipfs-posinfo":                "github.com/AstaFrode/boxo/filestore/posinfo",
		"github.com/ipfs/go-ipfs-util":                   "github.com/AstaFrode/boxo/util",
		"github.com/ipfs/go-ipfs-ds-help":                "github.com/AstaFrode/boxo/datastore/dshelp",
		"github.com/ipfs/go-verifcid":                    "github.com/AstaFrode/boxo/verifcid",
		"github.com/ipfs/go-ipfs-exchange-offline":       "github.com/AstaFrode/boxo/exchange/offline",
		"github.com/ipfs/go-ipfs-routing":                "github.com/AstaFrode/boxo/routing",
		"github.com/ipfs/go-ipfs-exchange-interface":     "github.com/AstaFrode/boxo/exchange",
		"github.com/ipfs/go-merkledag":                   "github.com/AstaFrode/boxo/ipld/merkledag",
		"github.com/boxo/ipld/car":                       "github.com/ipld/go-car",

		// Pre Boxo rename
		"github.com/ipfs/go-libipfs/gateway":               "github.com/AstaFrode/boxo/gateway",
		"github.com/ipfs/go-libipfs/bitswap":               "github.com/AstaFrode/boxo/bitswap",
		"github.com/ipfs/go-libipfs/files":                 "github.com/AstaFrode/boxo/files",
		"github.com/ipfs/go-libipfs/tar":                   "github.com/AstaFrode/boxo/tar",
		"github.com/ipfs/go-libipfs/coreiface":             "github.com/AstaFrode/boxo/coreiface",
		"github.com/ipfs/go-libipfs/unixfs":                "github.com/AstaFrode/boxo/ipld/unixfs",
		"github.com/ipfs/go-libipfs/pinning/remote/client": "github.com/AstaFrode/boxo/pinning/remote/client",
		"github.com/ipfs/go-libipfs/path":                  "github.com/AstaFrode/boxo/path",
		"github.com/ipfs/go-libipfs/namesys":               "github.com/AstaFrode/boxo/namesys",
		"github.com/ipfs/go-libipfs/mfs":                   "github.com/AstaFrode/boxo/mfs",
		"github.com/ipfs/go-libipfs/provider":              "github.com/AstaFrode/boxo/provider",
		"github.com/ipfs/go-libipfs/pinning/pinner":        "github.com/AstaFrode/boxo/pinning/pinner",
		"github.com/ipfs/go-libipfs/keystore":              "github.com/AstaFrode/boxo/keystore",
		"github.com/ipfs/go-libipfs/filestore":             "github.com/AstaFrode/boxo/filestore",
		"github.com/ipfs/go-libipfs/ipns":                  "github.com/AstaFrode/boxo/ipns",
		"github.com/ipfs/go-libipfs/blockservice":          "github.com/AstaFrode/boxo/blockservice",
		"github.com/ipfs/go-libipfs/chunker":               "github.com/AstaFrode/boxo/chunker",
		"github.com/ipfs/go-libipfs/fetcher":               "github.com/AstaFrode/boxo/fetcher",
		"github.com/ipfs/go-libipfs/blockstore":            "github.com/AstaFrode/boxo/blockstore",
		"github.com/ipfs/go-libipfs/filestore/posinfo":     "github.com/AstaFrode/boxo/filestore/posinfo",
		"github.com/ipfs/go-libipfs/util":                  "github.com/AstaFrode/boxo/util",
		"github.com/ipfs/go-libipfs/datastore/dshelp":      "github.com/AstaFrode/boxo/datastore/dshelp",
		"github.com/ipfs/go-libipfs/verifcid":              "github.com/AstaFrode/boxo/verifcid",
		"github.com/ipfs/go-libipfs/exchange/offline":      "github.com/AstaFrode/boxo/exchange/offline",
		"github.com/ipfs/go-libipfs/routing":               "github.com/AstaFrode/boxo/routing",
		"github.com/ipfs/go-libipfs/exchange":              "github.com/AstaFrode/boxo/exchange",

		// Unmigrated things
		"github.com/ipfs/go-libipfs/blocks": "github.com/ipfs/go-block-format",
		"github.com/AstaFrode/boxo/blocks":  "github.com/ipfs/go-block-format",
	},
	Modules: []string{
		"github.com/ipfs/go-bitswap",
		"github.com/ipfs/go-ipfs-files",
		"github.com/ipfs/tar-utils",
		"gihtub.com/ipfs/go-block-format",
		"github.com/ipfs/interface-go-ipfs-core",
		"github.com/ipfs/go-unixfs",
		"github.com/ipfs/go-pinning-service-http-client",
		"github.com/ipfs/go-path",
		"github.com/ipfs/go-namesys",
		"github.com/ipfs/go-mfs",
		"github.com/ipfs/go-ipfs-provider",
		"github.com/ipfs/go-ipfs-pinner",
		"github.com/ipfs/go-ipfs-keystore",
		"github.com/ipfs/go-filestore",
		"github.com/ipfs/go-ipns",
		"github.com/ipfs/go-blockservice",
		"github.com/ipfs/go-ipfs-chunker",
		"github.com/ipfs/go-fetcher",
		"github.com/ipfs/go-ipfs-blockstore",
		"github.com/ipfs/go-ipfs-posinfo",
		"github.com/ipfs/go-ipfs-util",
		"github.com/ipfs/go-ipfs-ds-help",
		"github.com/ipfs/go-verifcid",
		"github.com/ipfs/go-ipfs-exchange-offline",
		"github.com/ipfs/go-ipfs-routing",
		"github.com/ipfs/go-ipfs-exchange-interface",
		"github.com/ipfs/go-libipfs",
	},
}

func ReadConfig(r io.Reader) (Config, error) {
	var config Config
	err := json.NewDecoder(r).Decode(&config)
	if err != nil {
		return Config{}, fmt.Errorf("reading and decoding config: %w", err)
	}
	return config, nil
}
