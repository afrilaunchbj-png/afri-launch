package pptx

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

func TestBuild(t *testing.T) {
	images := [][]byte{
		[]byte("fake-png-1"),
		[]byte("fake-png-2"),
		[]byte("fake-png-3"),
	}
	data, err := Build(images, 12192000, 6858000)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip open: %v", err)
	}

	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}

	required := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"ppt/presentation.xml",
		"ppt/_rels/presentation.xml.rels",
		"ppt/slideMasters/slideMaster1.xml",
		"ppt/slideLayouts/slideLayout1.xml",
		"ppt/theme/theme1.xml",
		"ppt/slides/slide1.xml",
		"ppt/slides/slide2.xml",
		"ppt/slides/slide3.xml",
		"ppt/media/image1.png",
		"ppt/media/image2.png",
		"ppt/media/image3.png",
	}
	for _, r := range required {
		if !names[r] {
			t.Errorf("missing part %q", r)
		}
	}

	slide1 := readZip(t, zr, "ppt/slides/slide1.xml")
	if !bytes.Contains(slide1, []byte(`r:embed="rId2"`)) {
		t.Errorf("slide1.xml missing image embed")
	}
	rels := readZip(t, zr, "ppt/slides/_rels/slide2.xml.rels")
	if !bytes.Contains(rels, []byte(`Target="../media/image2.png"`)) {
		t.Errorf("slide2 rels missing image2")
	}
	pres := readZip(t, zr, "ppt/presentation.xml")
	if !bytes.Contains(pres, []byte(`cx="12192000" cy="6858000"`)) {
		t.Errorf("presentation missing slide size")
	}
}

func readZip(t *testing.T, zr *zip.Reader, name string) []byte {
	t.Helper()
	for _, f := range zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open %s: %v", name, err)
			}
			defer rc.Close()
			b, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}
			return b
		}
	}
	t.Fatalf("part %q not found", name)
	return nil
}
