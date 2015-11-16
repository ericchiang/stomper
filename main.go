package main

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	flag "github.com/ericchiang/stomper/Godeps/_workspace/src/github.com/ogier/pflag"
)

func fatalf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(2)
}

func main() {
	var (
		verbose  bool
		useStdin bool
		outfile  string
		tag      string
	)

	flag.BoolVarP(&verbose, "verbose", "v", false, "print progress while squashing.")
	flag.BoolVarP(&useStdin, "stdin", "i", false, "read stdin, instead of first argument.")
	flag.StringVarP(&outfile, "outfile", "o", "", "write to a file, instead of stdout.")
	flag.StringVarP(&tag, "tag", "t", "", "tag for squashed image.")
	flag.Parse()
	if len(flag.Args()) == 0 && !useStdin {
		flag.Usage()
		os.Exit(2)
	}

	var out io.Writer
	if outfile == "" {
		out = os.Stdout
	} else {
		f, err := os.OpenFile(outfile, os.O_WRONLY|os.O_CREATE|os.O_CREATE, 0644)
		if err != nil {
			fatalf("error: %v\n", err)
		}
		defer f.Close()
		out = f
	}

	var err error
	if useStdin {
		err = squashWithBuffer(os.Stdin, out, tag)
	} else {
		f, err := os.OpenFile(flag.Args()[0], os.O_RDONLY, 0644)
		if err != nil {
			fatalf("error: %v\n", err)
		}
		defer f.Close()
		err = squash(f, out, tag)
	}
	if err != nil {
		fatalf("error: %v\n", err)
	}
}

func squashWithBuffer(r io.Reader, w io.Writer, tag string) (err error) {
	tempfile, err := ioutil.TempFile("", "stomper_")
	if err != nil {
		return err
	}
	defer os.Remove(tempfile.Name())
	images, err := listImages(io.TeeReader(r, tempfile))
	if err != nil {
		return fmt.Errorf("failed to list layers: %v", err)
	}
	return squashArchive(tempfile, w, images, tag)
}

func squash(r io.ReadSeeker, w io.Writer, tag string) (err error) {
	images, err := listImages(r)
	if err != nil {
		return fmt.Errorf("failed to list layers: %v", err)
	}

	return squashArchive(r, w, images, tag)
}

func squashArchive(r io.ReadSeeker, w io.Writer, images []image, tag string) (err error) {
	if len(images) == 0 {
		return fmt.Errorf("no images found in archive")
	}
	if tag != "" {
		repos := map[string]bool{}
		for _, img := range images {
			if repos[img.repo] {
				return fmt.Errorf("tag flag specified with multiple tags in repo %s", img.repo)
			}
			repos[img.repo] = true
		}
	}

	tw := tar.NewWriter(w)
	defer func() {
		terr := tw.Close()
		if err == nil {
			err = terr
		}
	}()

	createdAt := time.Now()

	repos := map[string]map[string]string{}
	for _, img := range images {
		id, err := squashImage(r, tw, img, createdAt)
		if err != nil {
			return fmt.Errorf("squash image: %v", err)
		}
		if _, ok := repos[img.repo]; !ok {
			repos[img.repo] = map[string]string{}
		}
		t := tag
		if t == "" {
			t = img.tag
		}
		repos[img.repo][t] = id
	}

	data, err := json.Marshal(&repos)
	if err != nil {
		return err
	}
	hdr := &tar.Header{
		Name:     "repositories",
		Mode:     0644,
		Typeflag: tar.TypeReg,
		Size:     int64(len(data)),
		ModTime:  createdAt,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err = tw.Write(data)
	return err
}

func squashImage(r io.ReadSeeker, tw *tar.Writer, img image, createdAt time.Time) (id string, err error) {
	files, json, version, err := prepairSquash(r, img.layers)
	if err != nil {
		return "", fmt.Errorf("prepairing: %v", err)
	}

	if id, err = newId(); err != nil {
		return "", err
	}

	f, err := ioutil.TempFile("", "stomper_")
	if err != nil {
		return "", err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	data, err := prepairJSON(json, id, files, createdAt)
	if err != nil {
		return "", fmt.Errorf("prepair JSON: %v", err)
	}
	if err := squashLayers(r, f, img.layers, files); err != nil {
		return "", fmt.Errorf("squashing layers: %v", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return "", err
	}
	info, err := f.Stat()
	if err != nil {
		return "", err
	}

	tarFiles := []struct {
		Name string
		Type byte
		Size int64
		Body io.Reader
		Mode int64
	}{
		{id + "/", tar.TypeDir, 0, nil, 0755},
		{id + "/VERSION", tar.TypeReg, int64(len(version)), bytes.NewReader(version), 0644},
		{id + "/json", tar.TypeReg, int64(len(data)), bytes.NewReader(data), 0644},
		{id + "/layer.tar", tar.TypeReg, info.Size(), f, 0644},
	}
	for _, tf := range tarFiles {
		hdr := &tar.Header{
			Name:     tf.Name,
			Mode:     tf.Mode,
			Typeflag: tf.Type,
			Size:     tf.Size,
			ModTime:  createdAt,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return "", fmt.Errorf("failed to write header for %s: %v", tf.Name, err)
		}
		if tf.Size != 0 {
			if _, err := io.Copy(tw, tf.Body); err != nil {
				return "", err
			}
		}
	}
	return id, nil
}
