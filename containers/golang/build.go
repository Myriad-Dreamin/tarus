package container_build_golang

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type GolangBuildJob struct {
	saveFile  string
	saveFile2 string
	checksum  string
}

func New() *GolangBuildJob {
	var job = new(GolangBuildJob)
	saveDir := "containers/golang/build/downloads"
	job.saveFile = filepath.Join(saveDir, "go1.18.tar.gz")
	job.saveFile2 = filepath.Join(saveDir, "go1.18-prebuilts.tar")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		panic(err)
	}
	job.checksum = "e85278e98f57cdb150fe8409e6e5df5343ecb13cebf03a5d5ff12bd55a80264f"

	return job
}

func (job *GolangBuildJob) makeCmd(c string, args ...string) *exec.Cmd {
	var cmd = exec.Command(c, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func (job *GolangBuildJob) DownloadBinary() {

	var url = "https://go.dev/dl/go1.18.linux-amd64.tar.gz"
	cmd := job.makeCmd("wget", url, "-O", job.saveFile)
	cmd.Env = append(cmd.Env, "http_proxy=http://172.23.96.1:10809", "https_proxy=http://172.23.96.1:10809")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func pathRewrite(tr *tar.Reader, src, dst string) *tar.Reader {
	pr, pw := io.Pipe()
	r, w := tar.NewReader(pr), tar.NewWriter(pw)

	var worker = func() error {
		// Iterate through the files in the archive.
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				// end of tar archive
				_ = w.Close()
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			if !strings.HasPrefix(hdr.Name, src) {
				return fmt.Errorf("prefix mismatch want %s, got %s", src, hdr.Name)
			}
			hdr.Name = dst + strings.TrimPrefix(hdr.Name, src)
			if err = w.WriteHeader(hdr); err != nil {
				return err
			}
			_, err = io.Copy(w, tr)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	}
	go func() {
		_ = pw.CloseWithError(worker())
	}()

	return r
}

func teeTar(tw *tar.Writer, tr *tar.Reader) error {
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			_ = tw.Close()
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		_ = tw.WriteHeader(hdr)
		_, err = io.Copy(tw, tr)
		if err == nil {
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func Build() {
	job := New()
	if !job.CheckedBinary() {
		job.DownloadBinary()
		if !job.CheckedBinary() {
			fmt.Println("invalid binary downloaded")
			return
		}
	}

	// Open a file
	f, err := os.Open(job.saveFile)
	if err != nil {
		log.Fatal(err)
	}
	f2, err := os.OpenFile(job.saveFile2, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// Create an xz Reader
	r, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}
	// Create a tar Reader
	tr := tar.NewReader(r)
	tr = pathRewrite(tr, "go", "opt/prebuilts/go-1.18")
	tw := tar.NewWriter(f2)
	if err = teeTar(tw, tr); err != nil {
		log.Fatal(err)
	}
	_ = f.Close()
	_ = f2.Close()
}

func (job *GolangBuildJob) CheckedBinary() bool {
	if _, err := os.Stat(job.saveFile); err == nil {
		f, err := os.Open(job.saveFile)
		if err != nil {
			panic(err)
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)
		h := sha256.New()
		_, _ = io.Copy(h, f)
		h2 := hex.EncodeToString(h.Sum(nil))
		return h2 == job.checksum
	}
	return true
}
