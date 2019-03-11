package clair

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	v1 "github.com/coreos/clair/api/v1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	digest "github.com/opencontainers/go-digest"
	"golang.org/x/oauth2"
)

var errNoLayers = errors.New("no layers in image, vulnerabilities have not been populated")

// sendLayerToClair sends a layer from the passed repository (with the passed LayerReference).
func sendLayerToClair(config Config, tokenSrc oauth2.TokenSource, layerRef LayerReference) (err error) {
	var token *oauth2.Token

	token, err = tokenSrc.Token()
	if nil != err {
		return
	}

	layer := AddAuthorization(layerRef.GetLayer(), token)
	data := map[string]v1.Layer{
		"Layer": layer,
	}

	var buffer bytes.Buffer

	err = json.NewEncoder(&buffer).Encode(data)
	if nil != err {
		return
	}

	request, err := http.NewRequest(http.MethodPost, GetNewLayerURI(config.Hostname), &buffer)
	if nil != err {
		return
	}

	if config.UseBasicAuth() {
		config.UpdateRequest(request)
	}

	resp, err := http.DefaultClient.Do(request)
	if nil != err {
		return
	}

	defer resp.Body.Close()

	if 300 <= resp.StatusCode {
		err = fmt.Errorf("pushing layer to clair failed: %s", getErrorFromResponse(resp))
		return
	}

	err = resp.Body.Close()

	return
}

// getLayerFromClair gets the description of the Layer with the passed digest from Clair,
// using the passed digest.
func getLayerFromClair(config Config, digest digest.Digest) (layer v1.Layer, err error) {
	request, err := http.NewRequest(http.MethodGet, GetLayerURI(config.Hostname, digest), nil)
	if nil != err {
		return
	}

	if config.UseBasicAuth() {
		config.UpdateRequest(request)
	}

	resp, err := http.DefaultClient.Do(request)
	if nil != err {
		return
	}

	defer resp.Body.Close()

	if 300 <= resp.StatusCode {
		err = fmt.Errorf("getting layer from clair failed: %s", getErrorFromResponse(resp))
		return
	}

	var layerEnv v1.LayerEnvelope

	err = json.NewDecoder(resp.Body).Decode(&layerEnv)
	if nil != layerEnv.Layer {
		layer = *layerEnv.Layer
	}

	return
}

// getClairVulnerabilities gets a map[string]v1.Vulnerability from Clair, so that we can convert
// them to Voucher Vulnerabilities all at once.
func getClairVulnerabilities(manifest schema2.Manifest, config Config, tokenSrc oauth2.TokenSource, image reference.Canonical) (map[string]v1.Vulnerability, error) {
	vulns := make(map[string]v1.Vulnerability)
	var err error

	parent := digest.Digest("")

	for _, imageLayer := range manifest.Layers {
		current := imageLayer.Digest
		// send the current layer to Clair
		if err = sendLayerToClair(config, tokenSrc, NewLayerReference(image, current, parent)); nil != err {
			return vulns, err
		}

		parent = current
	}

	if "" != string(parent) {
		var layer v1.Layer

		// according to the Clair API, we can just get the vulnerabilities from the last
		// layer checked by Clair. The parent digest would have been updated at the end
		// of the manifest.Layers loop.
		if layer, err = getLayerFromClair(config, parent); nil != err {
			return vulns, err
		}

		for _, feature := range layer.Features {
			for _, vul := range feature.Vulnerabilities {
				vulns[vul.Name] = vul
			}
		}

	} else {
		return vulns, errNoLayers
	}

	return vulns, err
}
