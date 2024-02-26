package goupd

import (
	"compress/bzip2"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

type Version struct {
	ProjectName  string // project's name
	Channel      string // project channel (typically, branch name)
	DateTag      string // version's date tag
	GitTag       string // version's git tag (first 7 digits of git hash)
	UpdatePrefix string // internally used prefix
}

// GetUpdate returns update details for a given project.
//
// Deprecated: You should be using GetLatest instead
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

// IsCurrent returns true if the version matches the currently running program
func (v *Version) IsCurrent() bool {
	// we only check GitTag
	return v.GitTag == GIT_TAG
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

// Install will download the update and replace the currently running executable data
func (v *Version) Install() error {
	return v.SaveAs(self_exe)
}

// SaveAs will download the update to the given local filename, performing an atomic
// replacement of the file
func (v *Version) SaveAs(fn string) error {
	r, err := v.Download(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	defer r.Close()

	// install updated file (in io.Reader)
	exe, err := filepath.EvalSymlinks(fn)
	if err != nil {
		return fmt.Errorf("failed to find download target: %w", err)
	}

	// decompose executable
	dir := filepath.Dir(exe)
	name := filepath.Base(exe)

	// copy data in new file
	newPath := filepath.Join(dir, "."+name+".new")
	fp, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create new file: %w", err)
	}
	defer fp.Close()

	_, err = io.Copy(fp, r)
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}
	err = fp.Close()
	if err != nil {
		// delayed error because disk full?
		return fmt.Errorf("close failed: %w", err)
	}

	// move files
	oldPath := filepath.Join(dir, "."+name+".old")

	err = os.Rename(exe, oldPath)
	if err != nil {
		return fmt.Errorf("update rename failed: %w", err)
	}

	err = os.Rename(newPath, exe)
	if err != nil {
		// rename failed, revert previous rename (hopefully successful)
		os.Rename(oldPath, exe)
		return fmt.Errorf("update second rename failed: %w", err)
	}

	// attempt to remove old
	err = os.Remove(oldPath)
	if err != nil {
		// hide it since remove failed
		hideFile(oldPath)
	}

	return nil
}
