package tarus_container

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
	"github.com/nightlyone/lockfile"
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

func TestPull(t *testing.T) {
	image, err := pullBase("docker.io/library/ubuntu:20.04", "containers/.cache/bundles/linux")
	if err != nil {
		t.Fatal(err)
	}

	pythonInstall, err := tarball.LayerFromFile("containers/python/build/PrebuiltsPython-3.10.4.tar")
	if err != nil {
		t.Fatal(err)
	}

	image, err = mutate.AppendLayers(image, pythonInstall)
	if err != nil {
		t.Fatal(err)
	}
	var x = time.Now()
	f, err := os.OpenFile("containers/python/build/init.tar", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = f.Close()
	}()
	if err = crane.Export(image, f); err != nil {
		t.Fatal(fmt.Errorf("saving tarball %s: %v", "containers/python/build/init.tar", err))
	}
	fmt.Println("layer saved", time.Now().Sub(x))
}
