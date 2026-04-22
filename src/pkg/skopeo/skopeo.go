/*
 * Copyright © 2023 – 2024 Red Hat Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package skopeo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"os"

	"github.com/containers/toolbox/pkg/architecture"
	"github.com/containers/toolbox/pkg/shell"
	"github.com/docker/go-units"
	"github.com/sirupsen/logrus"
)

type Layer struct {
	Size json.Number
}
type Image struct {
	Architecture string `json:"Architecture"`
	LayersData   []Layer
	NameFull     string
	Os           string `json:"Os"`
}

func (image *Image) GetSizeHuman() (string, error) {
	imageSizeFloat, err := image.GetSize()
	if err != nil {
		return "", err
	}

	imageSizeHuman := units.HumanSize(imageSizeFloat)
	return imageSizeHuman, nil
}

func (image *Image) GetSize() (float64, error) {
	var imageSizeFloat float64

	if image.LayersData == nil {
		return -1, errors.New("'skopeo inspect' did not have LayersData")
	}

	for _, layer := range image.LayersData {
		if layerSize, err := layer.Size.Float64(); err != nil {
			return -1, err
		} else {
			imageSizeFloat += layerSize
		}
	}

	return imageSizeFloat, nil
}

func (image *Image) VerifyArchitectureMatch(expectedArchID int) error {
	expectedArchName := architecture.GetArchNameOCI(expectedArchID)
	logrus.Debugf("Verifying image %s supports architecture %s", image.NameFull, expectedArchName)

	actualArchID, err := architecture.ParseArgArchValue(image.Architecture)
	if err != nil {
		return err
	}

	if actualArchID != expectedArchID {
		// Single-arch image mismatch
		return fmt.Errorf("image %s is a single-architecture image for %s, but %s was requested",
			image.NameFull, image.Architecture, expectedArchName)
	}

	logrus.Debugf("Architecture verification passed: %s", expectedArchName)
	return nil
}

func CopyOverrideArch(source, destination string, archID int, authfile string) error {

	destinationWithTransport := "containers-storage:" + destination
	sourceWithTransport := "docker://" + source
	args := []string{"copy", "--override-arch", architecture.GetArchNameOCI(archID)}

	if authfile != "" {
		args = append(args, []string{"--src-authfile", authfile}...)
	}

	args = append(args, sourceWithTransport, destinationWithTransport)

	if logrus.GetLevel() < logrus.DebugLevel {
		if err := shell.Run("skopeo", nil, nil, nil, args...); err != nil {
			return err
		}
	} else {
		if err := shell.Run("skopeo", nil, os.Stderr, nil, args...); err != nil {
			return err
		}
	}

	return nil
}

// func Inspect(ctx context.Context, target string) (*Image, error) {
// 	var stdout bytes.Buffer

// 	targetWithTransport := "docker://" + target
// 	args := []string{"inspect", "--format", "json", targetWithTransport}

// 	if err := shell.RunContext(ctx, "skopeo", nil, &stdout, nil, args...); err != nil {
// 		return nil, err
// 	}

// 	output := stdout.Bytes()
// 	var image Image
// 	if err := json.Unmarshal(output, &image); err != nil {
// 		return nil, err
// 	}

// 	return &image, nil
// }

func Inspect(ctx context.Context, target string, archID int, authfile string) (*Image, error) {
	var stdout bytes.Buffer

	targetWithTransport := "docker://" + target
	args := []string{"inspect", "--format", "json"}

	if !architecture.HasContainerNativeArch(archID) {
		archName := architecture.GetArchNameOCI(archID)
		args = append(args, []string{"--override-arch", archName}...)
	}

	if authfile != "" {
		args = append(args, []string{"--authfile", authfile}...)
	}

	args = append(args, targetWithTransport)

	if _, err := shell.RunContextWithExitCode2(ctx, "skopeo", nil, &stdout, nil, args...); err != nil {
		return nil, err
	}

	output := stdout.Bytes()
	var image Image
	if err := json.Unmarshal(output, &image); err != nil {
		return nil, err
	}

	image.NameFull = target

	return &image, nil
}

// func InspectWithOverrideArch(target string, archID int, authfile string) (*Image, error) {
// 	var stdout bytes.Buffer

// 	archName := architecture.GetArchNameOCI(archID)
// 	targetWithTransport := "docker://" + target
// 	args := []string{"inspect", "--override-arch", archName, "--format", "json"}

// 	if authfile != "" {
// 		args = append(args, []string{"--authfile", authfile}...)
// 	}

// 	args = append(args, targetWithTransport)

// 	if err := shell.Run("skopeo", nil, &stdout, nil, args...); err != nil {
// 		return nil, err
// 	}

// 	output := stdout.Bytes()
// 	var image Image
// 	if err := json.Unmarshal(output, &image); err != nil {
// 		return nil, err
// 	}

// 	return &image, nil
// }

// func VerifyArchitectureMatch(imageTarget string, expectedArchID int, authfile string) error {
// 	expectedArchName := architecture.GetArchNameOCI(expectedArchID)
// 	logrus.Debugf("Verifying image %s supports architecture %s", imageTarget, expectedArchName)

// 	image, err := InspectWithOverrideArch(imageTarget, expectedArchID, authfile)
// 	if err != nil {
// 		// Multi-arch image mismatch or other error related to the image inspection
// 		return fmt.Errorf("failed to verify: image %s does not support architecture %s",
// 			imageTarget, expectedArchName)
// 	}

// 	actualArchID, err := architecture.ParseArgArchValue(image.Architecture)
// 	if err != nil {
// 		return err
// 	}

// 	if actualArchID != expectedArchID {
// 		// Single-arch image mismatch
// 		return fmt.Errorf("image %s is a single-architecture image for %s, but %s was requested",
// 			imageTarget, image.Architecture, expectedArchName)
// 	}

// 	logrus.Debugf("Architecture verification passed: %s", expectedArchName)
// 	return nil
// }
