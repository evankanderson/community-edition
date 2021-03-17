// Copyright 2020-2021 VMware Tanzu Community Edition contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	yaml "github.com/ghodss/yaml"
)

const (
	// DefaultTagVersion used after tagging a GA release
	DefaultTagVersion string = "dev.1"
	// DevFullPathFilename filename
	DevFullPathFilename string = "hack/DEV_BUILD_VERSION.yaml"
	// NewVersionFullPathFilename filename
	NewVersionFullPathFilename string = "hack/NEW_BUILD_VERSION"

	// NumberOfSemVerSeparators is 3
	NumberOfSemVerSeparators int = 3
	// NumberOfSeparatorsInDevTag is 2
	NumberOfSeparatorsInDevTag int = 2
)

var (
	// ErrInvalidVersionFormat is Invalid version format
	ErrInvalidVersionFormat = errors.New("invalid version format")
	// ErrDataReaderFailed is Datawriter is empty
	ErrDataReaderFailed = errors.New("datareader is empty")
	// ErrDataWriterFailed is Datawriter is empty
	ErrDataWriterFailed = errors.New("datawriter is empty")
)

type Version struct {
	Version string `json:"version"`
}

// Release version
func newRelease(current string) error {
	newVersion, err := incrementRelease(current)
	if err != nil {
		fmt.Printf("incrementRelease failed. Err: %v\n", err)
		return err
	}

	err = saveRelease(newVersion)
	if err != nil {
		fmt.Printf("saveDev failed. Err: %v\n", err)
		return err
	}

	return nil
}

func incrementRelease(tag string) (string, error) {
	items := strings.Split(tag, ".")
	if len(items) != NumberOfSemVerSeparators {
		fmt.Printf("Split version failed\n")
		return "", ErrInvalidVersionFormat
	}

	ver, err := strconv.Atoi(items[1])
	if err != nil {
		fmt.Printf("String to int failed\n")
		return "", ErrInvalidVersionFormat
	}

	newMajor := items[0]
	newMinor := ver + 1
	newVersionStr := fmt.Sprintf("%s.%d.0", newMajor, newMinor)
	fmt.Printf("incrementRelease: %s\n", newVersionStr)

	return newVersionStr, nil
}

func saveRelease(version string) error {
	// write the file
	fileWrite, err := os.OpenFile(NewVersionFullPathFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Printf("Open for write failed. Err: %v\n", err)
		return err
	}

	datawriter := bufio.NewWriter(fileWrite)
	if datawriter == nil {
		fmt.Printf("Datawriter creation failed\n")
		return ErrDataWriterFailed
	}

	_, err = datawriter.Write([]byte(version))
	if err != nil {
		fmt.Printf("datawriter.Write error. Err: %v\n", err)
		return err
	}
	datawriter.Flush()

	err = fileWrite.Close()
	if err != nil {
		fmt.Printf("fileWrite.Close failed. Err: %v\n", err)
		return err
	}

	return nil
}

// Release version

// Dev version
func resetDev() error {
	return saveDev(DefaultTagVersion)
}

func bumpDev() error {
	version, err := getTag()
	if err != nil {
		fmt.Printf("getTag failed. Err: %v\n", err)
		return err
	}

	newVersion, err := incrementDev(version)
	if err != nil {
		fmt.Printf("incrementDev failed. Err: %v\n", err)
		return err
	}

	err = saveDev(newVersion)
	if err != nil {
		fmt.Printf("saveDev failed. Err: %v\n", err)
		return err
	}

	return nil
}

func getTag() (string, error) {
	fileRead, err := os.OpenFile(DevFullPathFilename, os.O_RDONLY, 0755)
	if err != nil {
		fmt.Printf("Open for read failed. Err: %v\n", err)
		return "", err
	}

	dataReader := bufio.NewReader(fileRead)
	if dataReader == nil {
		fmt.Printf("Datareader creation failed\n")
		return "", ErrDataReaderFailed
	}

	byFile, err := io.ReadAll(dataReader)
	if err != nil {
		fmt.Printf("ReadAll failed. Err: %v\n", err)
		return "", err
	}

	version := &Version{}

	err = yaml.Unmarshal(byFile, version)
	if err != nil {
		fmt.Printf("Unmarshal failed. Err: %v\n", err)
		return "", err
	}

	return version.Version, nil
}

func incrementDev(tag string) (string, error) {
	items := strings.Split(tag, ".")
	if len(items) != NumberOfSeparatorsInDevTag {
		fmt.Printf("Split version failed\n")
		return "", ErrInvalidVersionFormat
	}

	ver, err := strconv.Atoi(items[1])
	if err != nil {
		fmt.Printf("String to int failed\n")
		return "", ErrInvalidVersionFormat
	}

	newVersion := ver + 1
	newVersionStr := fmt.Sprintf("dev.%d", newVersion)
	fmt.Printf("incrementDev: %s\n", newVersionStr)

	return newVersionStr, nil
}

func saveDev(tag string) error {
	version := &Version{
		Version: tag,
	}

	byRaw, err := yaml.Marshal(version)
	if err != nil {
		fmt.Printf("yaml.Marshal error. Err: %v\n", err)
		return err
	}
	fmt.Printf("BYTES:\n\n")
	fmt.Printf("%s\n", string(byRaw))

	// write the file
	fileWrite, err := os.OpenFile(DevFullPathFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Printf("Open for write failed. Err: %v\n", err)
		return err
	}

	datawriter := bufio.NewWriter(fileWrite)
	if datawriter == nil {
		fmt.Printf("Datawriter creation failed\n")
		return ErrDataWriterFailed
	}

	_, err = datawriter.Write(byRaw)
	if err != nil {
		fmt.Printf("datawriter.Write error. Err: %v\n", err)
		return err
	}
	datawriter.Flush()

	err = fileWrite.Close()
	if err != nil {
		fmt.Printf("fileWrite.Close failed. Err: %v\n", err)
		return err
	}

	return nil
}

// Dev version

func main() {
	// flags
	var tag string
	flag.StringVar(&tag, "tag", "", "The current release tag")

	var release bool
	flag.BoolVar(&release, "release", false, "Is this a release")

	flag.Parse()
	// flags

	if release {
		if tag == "" {
			fmt.Printf("Must supply -tag when -release is set\n")
			return
		}

		fmt.Printf("Cutting GA release, so resetting\n")
		err := resetDev()
		if err != nil {
			fmt.Printf("resetDev failed. Err: %v\n", err)
			return
		}

		err = newRelease(tag)
		if err != nil {
			fmt.Printf("newRelease failed. Err: %v\n", err)
			return
		}
	} else {
		fmt.Printf("Cutting RC release, so bumping\n")
		err := bumpDev()
		if err != nil {
			fmt.Printf("bumpDev failed. Err: %v\n", err)
			return
		}
	}

	fmt.Printf("Succeeded\n")
}