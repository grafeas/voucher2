package grafeasos

import (
	"errors"

	"github.com/Shopify/voucher/signer"
	grafeaspb "github.com/grafeas/client-go/0.1.0"
)

var errCannotAttest = errors.New("cannot create attestations, keyring is empty")

// Client implements voucher.MetadataClient, connecting to Grafeas.
type Client struct {
	grafeas        *grafeaspb.GrafeasV1Beta1ApiService // The client reference.
	keyring        signer.AttestationSigner            // The keyring used for signing metadata.
	binauthProject string                              // The project that Binauth Notes and Occurrences are written to.
	imageProject   string                              // The project that image information is stored.
}
