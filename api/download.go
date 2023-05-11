package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/nokia/docker-registry-client/registry"
	jww "github.com/spf13/jwalterweatherman"
)

const defaultRegistryBase = "https://registry-1.docker.io"

const (
	MediaTypeLayer = "application/vnd.docker.image.rootfs.diff.tar.gzip"

	// MediaTypeForeignLayer is the mediaType used for layers that must be
	// downloaded from foreign URLs.
	MediaTypeForeignLayer = "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip"

	// MediaTypeUncompressedLayer is the mediaType used for layers which
	// are not compressed.
	MediaTypeUncompressedLayer = "application/vnd.docker.image.rootfs.diff.tar"
)

type DownloadOpts struct {
	RegistryBase     string
	RegistryUsername string
	RegistryPassword string
	KeepLayers       bool
	UnpackMode       string
}

func getTargetImageSha(hub *registry.Registry, repo, tag, arch string) (string, error) {
	manifestList, err := hub.ManifestList(repo, tag)
	if err != nil {
		jww.WARN.Println(err)
		return "", err
	}
	manifests := manifestList.Manifests

	for _, m := range manifests {
		jww.INFO.Println("Manifest for arch ", m.Platform.Architecture)
		if arch == m.Platform.Architecture {
			return string(m.Digest), nil
		}
	}

	jww.WARN.Fatalln(fmt.Errorf("Could not found %s digest", arch))
	return "", fmt.Errorf("Could not found %s digest", arch)
}

func DownloadAndUnpackImage(sourceImage, output, arch string, opts *DownloadOpts) error {

	if opts.RegistryBase == "" {
		opts.RegistryBase = defaultRegistryBase
	}

	var TempDir = os.Getenv("TEMP_LAYER_FOLDER")
	if TempDir == "" {
		TempDir = "layers"
	}
	err := os.MkdirAll(TempDir, os.ModePerm)
	if err != nil {
		return err
	}
	if opts.KeepLayers == false {
		defer os.RemoveAll(TempDir)
	}

	if sourceImage != "" && strings.Contains(sourceImage, ":") {
		parts := strings.Split(sourceImage, ":")
		if parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("Bad usage. Image should be in this format: foo/my-image:latest")
		}
	}

	tagPart := "latest"
	repoPart := sourceImage
	parts := strings.Split(sourceImage, ":")
	if len(parts) > 1 {
		repoPart = parts[0]
		tagPart = parts[1]
	}

	jww.INFO.Println("Unpacking", repoPart, "tag", tagPart, "in", output)
	os.MkdirAll(output, os.ModePerm)
	username := opts.RegistryUsername
	password := opts.RegistryPassword
	hub, err := registry.New(opts.RegistryBase, username, password)
	if err != nil {
		jww.ERROR.Fatalln(err)
		return err
	}

	tagSha, err := getTargetImageSha(hub, repoPart, tagPart, arch)
	layers_sha := make([]string, 0)
	if err != nil {
		jww.WARN.Println(err)
		jww.INFO.Println("try manifest v1")
		// ref from https://docs.docker.com/registry/spec/manifest-v2-2/#manifest-list-field-descriptions
		// if v2 schema failed, failback to schema v1
		manifest, err := hub.ManifestV1(repoPart, tagPart)
		if err != nil {
			jww.ERROR.Fatalln(err)
			return err
		}

		layers := manifest.FSLayers
		for _, l := range layers {
			jww.INFO.Println("Layer ", l)
			// or obtain the digest from an existing manifest's FSLayer list
			s := string(l.BlobSum)
			i := strings.Index(s, ":")
			enc := s[i+1:]
			reader, err := hub.DownloadBlob(repoPart, l.BlobSum)
			layers_sha = append(layers_sha, enc)

			if reader != nil {
				defer reader.Close()
			}
			if err != nil {
				return err
			}

			where := path.Join(TempDir, enc)
			err = os.MkdirAll(where, os.ModePerm)
			if err != nil {
				jww.ERROR.Println(err)
				return err
			}

			out, err := os.Create(path.Join(where, "layer.tar"))
			if err != nil {
				return err
			}
			defer out.Close()
			if _, err := io.Copy(out, reader); err != nil {
				fmt.Println(err)
				return err
			}
		}
	} else {
		// try with manifestV2 schema
		manifest, err := hub.Manifest(repoPart, tagSha)
		if err != nil {
			jww.WARN.Fatalln(err)
			return err
		}

		layers := manifest.References()
		for _, l := range layers {
			jww.INFO.Println("Layer ", l)
			if l.MediaType != MediaTypeLayer {
				continue
			}
			// or obtain the digest from an existing manifest's FSLayer list
			s := string(l.Digest)
			i := strings.Index(s, ":")
			enc := s[i+1:]
			reader, err := hub.DownloadBlob(repoPart, l.Digest)
			layers_sha = append(layers_sha, enc)

			if reader != nil {
				defer reader.Close()
			}
			if err != nil {
				return err
			}

			where := path.Join(TempDir, enc)
			err = os.MkdirAll(where, os.ModePerm)
			if err != nil {
				jww.ERROR.Println(err)
				return err
			}

			out, err := os.Create(path.Join(where, "layer.tar"))
			if err != nil {
				return err
			}
			defer out.Close()
			if _, err := io.Copy(out, reader); err != nil {
				fmt.Println(err)
				return err
			}
		}
	}

	jww.INFO.Println("Download complete")

	export, err := CreateExport(TempDir)
	if err != nil {
		fmt.Println(err)
		return err
	}

	jww.INFO.Println("Unpacking...")

	err = export.UnPackLayers(layers_sha, output, opts.UnpackMode)
	if err != nil {
		jww.INFO.Fatal(err)
		return err
	}

	jww.INFO.Println("Done")
	return nil
}

func CreateExport(layers string) (*Export, error) {

	export := &Export{
		Entries: map[string]*ExportedImage{},
		Path:    layers,
	}

	dirs, err := ioutil.ReadDir(export.Path)
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {

		if !dir.IsDir() {
			continue
		}

		entry := &ExportedImage{
			Path:         filepath.Join(export.Path, dir.Name()),
			LayerTarPath: filepath.Join(export.Path, dir.Name(), "layer.tar"),
			LayerDirPath: filepath.Join(export.Path, dir.Name(), "layer"),
		}

		export.Entries[dir.Name()] = entry
	}

	return export, err
}
