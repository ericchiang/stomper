package main

import (
	"archive/tar"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strings"
	"time"
)

type image struct {
	repo   string
	tag    string
	layers []string
}

type fileInfo struct {
	layerId string
	size    int64
}

type layer struct {
	Id     string `json:"id"`
	Parent string `json:"parent"`
}

func newId() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func prepairJSON(jsonMeta []byte, id string, files map[string]fileInfo, createAt time.Time) ([]byte, error) {
	meta := make(map[string]interface{})
	if err := json.Unmarshal(jsonMeta, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal: %v", err)
	}
	// sanity check
	mustHave := []string{"id", "created"}
	for _, field := range mustHave {
		if _, ok := meta[field]; !ok {
			return nil, fmt.Errorf("json metadata does not have a '%s' field %s", field, jsonMeta)
		}
	}
	var size int64 = 0
	for _, info := range files {
		size += info.size
	}
	meta["Size"] = size
	meta["id"] = id
	meta["created"] = createAt

	// if digests are present, delete them
	for _, field := range []string{"parent", "parent_id", "layer_id"} {
		delete(meta, field)
	}
	return json.Marshal(&meta)
}

func prepairSquash(rs io.ReadSeeker, layers []string) (files map[string]fileInfo, json, version []byte, err error) {
	files = make(map[string]fileInfo)
	json = nil
	for _, layerId := range layers {
		var layer *tar.Reader
		var newJSON []byte
		newJSON, version, layer, err = readLayer(rs, layerId)
		if err != nil {
			err = fmt.Errorf("read layers: %v", err)
			return
		}
		if json != nil {
			json, err = merge(json, newJSON)
			if err != nil {
				err = fmt.Errorf("merge: %v", err)
				return
			}
		} else {
			json = newJSON
		}
		for {
			var h *tar.Header
			if h, err = layer.Next(); err == io.EOF {
				err = nil
				break
			}
			if err != nil {
				return
			}

			dir, file := path.Split(h.Name)
			if !strings.HasPrefix(file, ".wh.") {
				files[h.Name] = fileInfo{layerId, h.Size}
				continue
			}

			name := path.Join(dir, strings.TrimPrefix(file, ".wh."))
			if _, ok := files[name]; !ok {
				name = name + "/"
				if _, ok := files[name]; !ok {
					err = fmt.Errorf("whiteout file '%s' found without existing file '%s'", h.Name, name)
					return
				}
			}
			delete(files, name)
		}
	}
	return
}

func squashLayers(rs io.ReadSeeker, w io.Writer, layers []string, files map[string]fileInfo) (err error) {
	tw := tar.NewWriter(w)
	defer func() {
		terr := tw.Close()
		if err == nil {
			err = terr
		}
	}()
	for _, layerId := range layers {
		_, _, layer, err := readLayer(rs, layerId)
		if err != nil {
			return err
		}
		for {
			var h *tar.Header
			if h, err = layer.Next(); err == io.EOF {
				err = nil
				break
			}
			if err != nil {
				return err
			}
			if files[h.Name].layerId != layerId {
				continue
			}
			if err := tw.WriteHeader(h); err != nil {
				return err
			}
			if h.Size > 0 {
				if _, err := io.Copy(tw, layer); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func listImages(r io.Reader) ([]image, error) {

	layers, repositories, err := listLayers(r)
	if err != nil {
		return nil, err
	}
	parentOf := func(id string) (string, bool) {
		for _, layer := range layers {
			if layer.Id == id {
				return layer.Parent, true
			}
		}
		return "", false
	}

	images := []image{}
	for repo, imgs := range repositories {
		for tag, id := range imgs {
			img := image{
				repo:   repo,
				tag:    tag,
				layers: []string{id},
			}
			var parent string
			for {
				var found bool
				parent, found = parentOf(id)
				if !found {
					return nil, fmt.Errorf("could not find image %s", id)
				}
				if parent == "" {
					break
				}
				img.layers = append(img.layers, parent)
				id = parent
			}
			reverse(img.layers)

			images = append(images, img)
		}
	}
	return images, nil
}

func listLayers(r io.Reader) (layers []layer, repositories map[string]map[string]string, err error) {
	repositories = map[string]map[string]string{}

	tr := tar.NewReader(r)
	layers = []layer{}
	for {
		var h *tar.Header
		h, err = tr.Next()
		if err == io.EOF {
			return nil, nil, errors.New("no repositories file in image archive")
		}
		if err != nil {
			return nil, nil, err
		}
		if h.Name == "repositories" {
			var data []byte
			data, err = ioutil.ReadAll(tr)
			if err != nil {
				return nil, nil, err
			}
			if err := json.Unmarshal(data, &repositories); err != nil {
				return nil, nil, fmt.Errorf("could not decode repositories: %v", err)
			}
			return
		}

		if !strings.HasSuffix(h.Name, "/json") {
			continue
		}
		if h.Typeflag != tar.TypeReg {
			return nil, nil, fmt.Errorf("expected record %s to be a file", h.Name)
		}
		var l layer
		if err := json.NewDecoder(tr).Decode(&l); err != nil {
			return nil, nil, fmt.Errorf("failed to decode %s %v", h.Name, err)
		}
		if l.Id == "" {
			return nil, nil, fmt.Errorf("metadata file %s had no layer id", h.Name)
		}
		layers = append(layers, l)
	}

}

func nextFile(tr *tar.Reader, name string) (io.Reader, error) {
	h, err := tr.Next()
	if err == io.EOF {
		return nil, io.ErrUnexpectedEOF
	}
	if err != nil {
		return nil, err
	}
	if h.Name != name {
		return nil, fmt.Errorf("expected file %s got %s", name, h.Name)
	}
	if h.Typeflag != tar.TypeReg {
		return nil, fmt.Errorf("expected %s to be a regular file", name)
	}
	return tr, nil
}

func readFile(tr *tar.Reader, name string) ([]byte, error) {
	r, err := nextFile(tr, name)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}

func readLayer(r io.ReadSeeker, id string) (json, version []byte, layer *tar.Reader, err error) {
	if _, err = r.Seek(0, 0); err != nil {
		return
	}

	tr := tar.NewReader(r)
	for {
		var h *tar.Header
		h, err = tr.Next()
		if err == io.EOF {
			return nil, nil, nil, fmt.Errorf("layer not found")
		}
		if err != nil {
			return
		}
		if h.Typeflag != tar.TypeDir || h.Name != id+"/" {
			continue
		}
		if version, err = readFile(tr, id+"/VERSION"); err != nil {
			return
		}
		if json, err = readFile(tr, id+"/json"); err != nil {
			return
		}
		var r io.Reader
		r, err = nextFile(tr, id+"/layer.tar")
		if err != nil {
			return
		}
		layer = tar.NewReader(r)
		return
	}
}

func reverse(s []string) {
	n := len(s)
	for i := 0; i < (n / 2); i++ {
		j := n - (i + 1)
		s[i], s[j] = s[j], s[i]
	}
}
