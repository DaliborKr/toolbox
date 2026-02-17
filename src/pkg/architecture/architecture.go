/*
 * Copyright © 2019 – 2026 Red Hat Inc.
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

package architecture

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	NotSpecifiedArchID = iota
	ARM64ArchID
	PPC64LEArchID
	X86_64ArchID
)

var archELFMagic = map[int][]byte{
	ARM64ArchID:   {0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0xb7, 0x00},
	PPC64LEArchID: {0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x15, 0x00},
	X86_64ArchID:  {0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x3e, 0x00},
}

var archELFMask = map[int][]byte{
	ARM64ArchID:   {0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0xff},
	PPC64LEArchID: {0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0x00},
	X86_64ArchID:  {0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xfe, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0xff},
}

var archNames = map[int]string{
	NotSpecifiedArchID: "",
	ARM64ArchID:        "arm64",
	PPC64LEArchID:      "ppc64le",
	X86_64ArchID:       "x86_64",
}

var (
	HostArchID int
)

// TODO: Add support for other architectures
//   - see command "go tool dist list"
var supportedArgArchValues = map[string]int{
	"arm64":   ARM64ArchID,
	"aarch64": ARM64ArchID,
	"ppc64le": PPC64LEArchID,
	"x86_64":  X86_64ArchID,
	"amd64":   X86_64ArchID,
}

func GetArchELFMagicAll() map[int][]byte {
	return archELFMagic
}

func GetArchELFMagic(archID int) []byte {
	return archELFMagic[archID]
}

func GetArchELFMask(archID int) []byte {
	return archELFMask[archID]
}

func GetArchName(arch int) string {
	if arch == NotSpecifiedArchID {
		logrus.Warnf("Getting arch name for not specified architecture")
		return archNames[arch]
	}
	return archNames[arch]
}

// TODO is this really necessary??
// the "go dist list -json" what architectures can go compile to
// func GetGoSupportedArchitectures() (map[int]bool, error) {
// 	type platform struct {
// 		GOOS    string `json:"GOOS"`
// 		GOARCH  string `json:"GOARCH"`
// 		CgoSupp bool   `json:"CgoSupported"`
// 	}

// 	var stdout bytes.Buffer

// 	err := shell.Run("go", nil, &stdout, nil, "tool", "dist", "list", "-json")

// 	if err != nil {
// 		return nil, err
// 	}

// 	data := stdout.Bytes()
// 	var platforms []platform
// 	if err := json.Unmarshal(data, &platforms); err != nil {
// 		return nil, err
// 	}

// 	supportedArchs := make(map[int]bool)

// 	for _, p := range platforms {
// 		archID, _ := ParseArgArchValue(p.GOARCH)

// 		if archID != NotSpecifiedArchID && p.CgoSupp && p.GOOS == "linux" {
// 			supportedArchs[archID] = true
// 		}
// 	}

// 	return supportedArchs, nil
// }

func ParseArgArchValue(value string) (int, error) {
	archID, exists := supportedArgArchValues[value]
	if !exists {
		return NotSpecifiedArchID, fmt.Errorf("architecture '%s' is not supported", value)
	}

	return archID, nil
}
