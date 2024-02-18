package goupd

import (
	"compress/bzip2"
	"fmt"
	"io"
	"net/http"
)

type Version struct {
	ProjectName  string
	Channel      string
	DateTag      string
	GitTag       string
	UpdatePrefix string
}

// GetUpdate returns update details for a given project.
//
// Deprecated
func GetUpdate(projectName, curTag, os, arch, channel string) (string, string, string, error) {
	v, err := GetLatest(projectName, channel)
	if err != nil {
		return "", "", "", err
	}
	if curTag != v.GitTag {
		// there is a different version, check if it supports the needed architecture
		err = v.CheckArch(os, arch)
		if err != nil {
			return "", "", "", err
		}
	}

	// success
	return v.URL(os, arch), v.DateTag, v.GitTag, nil
}

// GetLatest returns the latest version information for a given projectName and channel
func GetLatest(projectName, channel string) (*Version, error) {
	latest := HOST + projectName + "/LATEST"
	if channel != "" {
		// for example LATEST-testing
		latest += "-" + channel
	}
	updInfo, err := httpGetFields(latest)
	if err != nil {
		return nil, fmt.Errorf("failed to read latest version: %w", err)
	}
	if len(updInfo) != 3 {
		return nil, fmt.Errorf("failed to parse update data (%v)", updInfo)
	}

	res := &Version{
		ProjectName:  projectName,
		Channel:      channel,
		DateTag:      updInfo[0], // 20230518035112
		GitTag:       updInfo[1], // e894f37
		UpdatePrefix: updInfo[2], // packagename_stable_20230518035112_e894f37
	}

	return res, nil
}

// CheckArch checks if the provided version is compatible with the provided os and arch
func (v *Version) CheckArch(os, arch string) error {
	target := os + "_" + arch

	// check if compatible version is available
	archs, err := httpGetFields(HOST + v.ProjectName + "/" + v.UpdatePrefix + ".arch")
	if err != nil {
		return fmt.Errorf("failed to read arch info: %w", err)
	}

	for _, subarch := range archs {
		if subarch == target {
			return nil
		}
	}

	return fmt.Errorf("no version available for %s", target)
}

// URL generates the download url for the provided os and arch. Note that the URL will
// point to a compressed file, use Download() to instead directly receive decompressed
// data.
func (v *Version) URL(os, arch string) string {
	target := os + "_" + arch
	dlUrl := HOST + v.ProjectName + "/" + v.UpdatePrefix + "/" + v.ProjectName + "_" + target + ".bz2"

	return dlUrl
}

// readCloser is a simple struct to mix a Reader and a Closer, so these can be different
// objects.
type readCloser struct {
	io.Reader
	io.Closer
}

// Download returns a ReadCloser that allows reading the updated executable data. It will
// handle any decompression that might be needed, so the data can be read directly. Make
// sure to close the returned ReadCloser after usage.
func (v *Version) Download(os, arch string) (io.ReadCloser, error) {
	resp, err := http.Get(v.URL(os, arch))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch update: %w", err)
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to fetch update: bad HTTP status %s", resp.Status)
	}
	r := bzip2.NewReader(resp.Body)

	return &readCloser{Reader: r, Closer: resp.Body}, nil
}
