package grafeasos

import (
	"context"
	"errors"

	"github.com/Shopify/voucher"
	"github.com/Shopify/voucher/attestation"
	"github.com/Shopify/voucher/signer"
	"github.com/docker/distribution/reference"
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

// CanAttest returns true if the client can create and sign attestations.
func (g *Client) CanAttest() bool {
	return nil != g.keyring
}

// NewPayloadBody returns a payload body appropriate for this MetadataClient.
func (g *Client) NewPayloadBody(reference reference.Canonical) (string, error) {
	payload, err := attestation.NewPayload(reference).ToString()
	if err != nil {
		return "", err
	}

	return payload, err
}

// AddAttestationToImage adds a new attestation with the passed Attestation
// to the image described by ImageData.
func (g *Client) AddAttestationToImage(ctx context.Context, reference reference.Canonical, payload voucher.Attestation) (interface{}, error) {
	if !g.CanAttest() {
		return nil, errCannotAttest
	}

	signed, err := voucher.SignAttestation(g.keyring, payload)
	if nil != err {
		return nil, err
	}

	binauthProjectPath := "projects/" + g.binauthProject

	contentType := grafeaspb.SIMPLE_SIGNING_JSON_AttestationPgpSignedAttestationContentType

	attestation := grafeaspb.V1beta1attestationDetails{Attestation: &grafeaspb.AttestationAttestation{PgpSignedAttestation: &grafeaspb.AttestationPgpSignedAttestation{Signature: signed.Signature,
		PgpKeyId: signed.KeyID, ContentType: &contentType}}}

	occurrence := g.getCreateOccurrence(reference, payload.CheckName, &attestation, binauthProjectPath)
	occ, _, err := g.grafeas.CreateOccurrence(ctx, binauthProjectPath, occurrence)

	if isAttestionExistsErr(err) {
		err = nil
		occ = grafeaspb.V1beta1Occurrence{}
	}

	return &occ, err

}

func (g *Client) getCreateOccurrence(reference reference.Canonical, parentNoteID string, attestation *grafeaspb.V1beta1attestationDetails, binauthProjectPath string) grafeaspb.V1beta1Occurrence {
	noteName := binauthProjectPath + "/notes/" + parentNoteID

	resource := grafeaspb.V1beta1Resource{
		Uri: "https://" + reference.Name() + "@" + reference.Digest().String(),
	}

	noteKind := grafeaspb.ATTESTATION_V1beta1NoteKind

	occurrence := grafeaspb.V1beta1Occurrence{Resource: &resource, NoteName: noteName, Kind: &noteKind, Attestation: attestation}

	return occurrence
}
