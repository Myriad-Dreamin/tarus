package tarus_container

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	oci_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge/oci"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/nightlyone/lockfile"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func getLocalRegistry() {

}

func withFlock(f lockfile.Lockfile, op func() error) error {
	err := f.TryLock()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer f.Unlock()

	return op()
}

func pullRemote(ref name.Reference) (v1.Image, error) {
	var options []remote.Option
	proxy := os.Getenv("http_proxy")
	if len(proxy) != 0 {
		proxyUrl, _ := url.Parse(proxy)
		options = append(options, remote.WithTransport(&http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}))
	}
	options = append(options, remote.WithAuthFromKeychain(authn.DefaultKeychain))

	image, err := remote.Image(ref, options...)
	if err != nil {
		return nil, fmt.Errorf("pull image %v", err)
	}
	return image, nil
}

func pullBase(src string, path string) (v1.Image, error) {
	ref, err := name.NewTag(src)
	if err != nil {
		return nil, fmt.Errorf("parsing reference %q: %v", src, err)
	}

	cacheRef, err := filepath.Abs("containers/.cache")
	if err != nil {
		return nil, fmt.Errorf("invalid container cache path %s: %v", cacheRef, err)
	}

	p, err := layout.FromPath(path)
	if err != nil {
		p, err = layout.Write(path, empty.Index)
		if err != nil {
			return nil, err
		}
	}

	h := sha256.New()
	h.Write([]byte(ref.Name()))
	h2 := hex.EncodeToString(h.Sum(nil))
	href := filepath.Join(cacheRef, "refs", h2[:2], h2[2:4], h2[4:])

	flockFile := filepath.Join(cacheRef, "index.lock")
	locker, err := lockfile.New(flockFile)
	if err != nil {
		return nil, fmt.Errorf("invalid container cache lock %s: %v", cacheRef, err)
	}

	var image v1.Image
	if err = withFlock(locker, func() error {
		st, err := os.Stat(href)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		timeDiff := time.Now().Sub(st.ModTime())
		if timeDiff > time.Hour*24 {
			return nil
		}

		b, err := os.ReadFile(href)
		if err != nil {
			return err
		}

		imageDigest, err := v1.NewHash(string(b))
		if err != nil {
			return err
		}
		image, _ = p.Image(imageDigest)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("container cache query failed %s: %v", cacheRef, err)
	}

	var imageDigest v1.Hash
	if image == nil {
		image, err = pullRemote(ref)
		if err != nil {
			return nil, fmt.Errorf("pull remote image %s: %v", path, err)
		}
		imageDigest, err = image.Digest()
		if err != nil {
			return nil, fmt.Errorf("digest image %s: %v", path, err)
		}

		if err = withFlock(locker, func() error {
			err = os.MkdirAll(filepath.Dir(href), 0755)
			if err != nil {
				return err
			}
			err = os.WriteFile(href, []byte(imageDigest.String()), 0644)
			if err != nil {
				return err
			}

			return nil
		}); err != nil {
			return nil, fmt.Errorf("container cache query failed %s: %v", cacheRef, err)
		}
	}
	image = cache.Image(image, cache.NewFilesystemCache(filepath.Join(cacheRef, "crane")))
	if len(imageDigest.Hex) == 0 {
		imageDigest, err = image.Digest()
		if err != nil {
			return nil, fmt.Errorf("digest image %s: %v", path, err)
		}
	}

	_, err = p.Image(imageDigest)
	if err == nil {
		return image, nil
	}
	var x = time.Now()
	if err = crane.SaveOCI(image, path); err != nil {
		return nil, fmt.Errorf("saving tarball %s: %v", path, err)
	}
	fmt.Println("layer saved", time.Now().Sub(x))

	return image, nil
}

func createStdImg(image v1.Image) (*v1.Manifest, error) {
	b, err := image.RawConfigFile()
	if err != nil {
		return nil, err
	}

	cfgHash, cfgSize, err := v1.SHA256(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	m := &v1.Manifest{
		SchemaVersion: 2,
		MediaType:     types.DockerManifestSchema2,
		Config: v1.Descriptor{
			MediaType: types.DockerConfigJSON,
			Size:      cfgSize,
			Digest:    cfgHash,
		},
	}

	return m, nil
}

func TestPull(t *testing.T) {
	image, err := pullBase("docker.io/library/ubuntu:20.04", "containers/.cache/bundles/linux")
	if err != nil {
		t.Fatal(err)
	}

	pythonInstall, err := tarball.LayerFromFile("containers/python/build/PrebuiltsPython-3.10.4.tar")
	if err != nil {
		t.Fatal(err)
	}

	image, err = mutate.Append(image, mutate.Addendum{
		Layer: pythonInstall,
		History: v1.History{
			Author:     "Myriad-Dreamin",
			Created:    v1.Time{Time: time.Now()},
			CreatedBy:  "tarus-container append env prebuilts/python-3.10.4.tar",
			Comment:    "auto generated",
			EmptyLayer: false,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var x = time.Now()
	client, err := oci_judge.NewContainerdServer()
	if err != nil {
		t.Fatal(err)
	}

	target := "kcr.skyline.io/library/python-judge:3.10.4"
	ref, err := name.NewTag(target)
	if err != nil {
		t.Fatal(fmt.Errorf("parsing reference %q: %v", target, err))
	}

	pr, pw := io.Pipe()
	go func() {
		_ = pw.CloseWithError(tarball.MultiWrite(map[name.Tag]v1.Image{
			ref: image,
		}, pw))
	}()

	if err = client.ImportOCIArchiveR(context.Background(), pr, target); err != nil {
		t.Fatal(fmt.Errorf("saving archive to daemon %v", err))
	}
	fmt.Println("layer saved", time.Now().Sub(x))
}
