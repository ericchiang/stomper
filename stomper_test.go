package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestListLayers(t *testing.T) {
	imgFile := "tests/ubuntu_test.docker"
	requiresImage(t, imgFile)

	file, err := os.Open(imgFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	images, err := listImages(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 image got %d", len(images))
	}
	img := images[0]
	if img.repo != "ubuntu_test" || img.tag != "latest" {
		t.Errorf("expected ubuntu_test:latest got %s:%s", img.repo, img.tag)
	}
}

func TestSquashImage(t *testing.T) {
	imgFile := "tests/busybox_test.docker"
	requiresImage(t, imgFile)

	file, err := os.Open(imgFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := squash(file, ioutil.Discard, ""); err != nil {
		t.Fatal(err)
	}
}

func TestPrepairSquash(t *testing.T) {
	imgFile := "tests/busybox_test.docker"
	requiresImage(t, imgFile)

	file, err := os.Open(imgFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	images, err := listImages(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, img := range images {
		files, _, _, err := prepairSquash(file, img.layers)
		if err != nil {
			t.Error(err)
			continue
		}
		for _, name := range []string{"hi", "foo"} {
			if _, ok := files[name]; ok {
				t.Errorf("expected file %s to be deleted", name)
			}
		}
	}
}

func TestSquashLayers(t *testing.T) {
	imgFile := "tests/busybox_test.docker"
	requiresImage(t, imgFile)

	file, err := os.Open(imgFile)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	images, err := listImages(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, img := range images {
		files, _, _, err := prepairSquash(file, img.layers)
		if err != nil {
			t.Error(err)
			continue
		}
		err = squashLayers(file, ioutil.Discard, img.layers, files)
		if err != nil {
			t.Error(err)
			continue
		}
	}
}

func TestNewId(t *testing.T) {
	id, err := newId()
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 64 {
		t.Errorf("id must be of length 64, got %d", len(id))
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		start, exp []string
	}{
		{[]string{}, []string{}},
		{[]string{"1", "2"}, []string{"2", "1"}},
		{[]string{"4", "5", "6"}, []string{"6", "5", "4"}},
	}
	for _, test := range tests {
		reverse(test.start)
		if len(test.start) != len(test.exp) {
			t.Errorf("%s does not equal %s", test.start, test.exp)
			continue
		}

		for i, s := range test.start {
			if test.exp[i] != s {
				t.Errorf("%s does not equal %s", test.start, test.exp)
				break
			}
		}
	}
}

func requiresImage(t *testing.T, name string) {
	_, err := os.Stat(name)
	if err == nil {
		return
	}
	if os.IsNotExist(err) {
		t.Skipf("test requires image %s, skipping", name)
	}
	t.Fatal(err)
}
