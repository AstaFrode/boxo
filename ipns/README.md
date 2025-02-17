# IPNS

> A reference implementation of the IPNS Record and Verification specification.

## Documentation

- Go Documentation: https://pkg.go.dev/github.com/ipfs/boxo/ipns
- IPNS Record Specification: https://specs.ipfs.tech/ipns/ipns-record/

## Example

Here's an example on how to create an IPNS Record:

```go
import (
	"crypto/rand"
	"time"

	"github.com/AstaFrode/boxo/ipns"
	"github.com/AstaFrode/boxo/path"
	ic "github.com/AstaFrode/go-libp2p/core/crypto"
)

func main() {
	// Create a private key to sign your IPNS record. Most of the time, you
	// will want to retrieve an already-existing key from Kubo, for example.
	sk, _, err := ic.GenerateEd25519Key(rand.Reader)
	if err != nil {
		panic(err)
	}

	// Define the path this record will point to.
	path := path.FromString("/ipfs/bafkqac3jobxhgidsn5rww4yk")

	// Until when the record is valid.
	eol := time.Now().Add(time.Hour)

	// For how long should caches cache the record.
	ttl := time.Second * 20

	record, err := ipns.NewRecord(sk, path, 1, eol, ttl)
	if err != nil {
		panic(err)
	}

	// Now you have an IPNS Record.
}
```
