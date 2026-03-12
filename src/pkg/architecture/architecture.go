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
	"debug/elf"
	"fmt"
	"strings"

	"github.com/containers/toolbox/pkg/utils"
	"github.com/sirupsen/logrus"
)

type Architecture struct {
	ID         int
	NameBinfmt string
	NameOCI    string
	Aliases    []string
	ELFMagic   []byte
	ELFMask    []byte

	BinfmtFlags     string
	BinfmtName      string
	BinfmtMagicType string
	BinfmtOffset    string
}

const (
	NotSpecifiedArchID = iota
	AARCH64ArchID
	PPC64LEArchID
	X86_64ArchID
)

var supportedArchitectures = map[int]Architecture{
	AARCH64ArchID: {
		ID:         AARCH64ArchID,
		NameBinfmt: "aarch64",
		NameOCI:    "arm64",
		Aliases:    []string{"aarch64", "arm64"},
		ELFMagic:   []byte{0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0xb7, 0x00},
		ELFMask:    []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0xff},
	},
	PPC64LEArchID: {
		ID:         PPC64LEArchID,
		NameBinfmt: "ppc64le",
		NameOCI:    "ppc64le",
		Aliases:    []string{"ppc64le"},
		ELFMagic:   []byte{0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x15, 0x00},
		ELFMask:    []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0x00},
	},
	X86_64ArchID: {
		ID:         X86_64ArchID,
		NameBinfmt: "x86_64",
		NameOCI:    "amd64",
		Aliases:    []string{"x86_64", "amd64"},
		ELFMagic:   []byte{0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x3e, 0x00},
		ELFMask:    []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xfe, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0xff},
	},
}

// var archELFMagic = map[int][]byte{
// 	AARCH64ArchID: {0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0xb7, 0x00},
// 	PPC64LEArchID: {0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x15, 0x00},
// 	X86_64ArchID:  {0x7f, 0x45, 0x4c, 0x46, 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x3e, 0x00},
// }

// var archELFMask = map[int][]byte{
// 	AARCH64ArchID: {0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0xff},
// 	PPC64LEArchID: {0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0x00},
// 	X86_64ArchID:  {0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xfe, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe, 0xff, 0xff, 0xff},
// }

// var archNamesBinfmt = map[int]string{
// 	NotSpecifiedArchID: "",
// 	AARCH64ArchID:      "aarch64",
// 	PPC64LEArchID:      "ppc64le",
// 	X86_64ArchID:       "x86_64",
// }

// var archNamesOCI = map[int]string{
// 	NotSpecifiedArchID: "",
// 	AARCH64ArchID:      "arm64",
// 	PPC64LEArchID:      "ppc64le",
// 	X86_64ArchID:       "amd64",
// }

// // TODO: Add support for other architectures
// //   - see command "go tool dist list"
// var supportedArgArchValues = map[string]int{
// 	"arm64":   AARCH64ArchID,
// 	"aarch64": AARCH64ArchID,
// 	"ppc64le": PPC64LEArchID,
// 	"x86_64":  X86_64ArchID,
// 	"amd64":   X86_64ArchID,
// }

var (
	HostArchID             int
	supportedArgArchValues map[string]int
)

func init() {
	supportedArgArchValues = make(map[string]int)
	for archID, arch := range supportedArchitectures {
		for _, alias := range arch.Aliases {
			supportedArgArchValues[alias] = archID
		}
	}
}

// Currently no refence
func GetArchELFMagicAll() map[int][]byte {
	result := make(map[int][]byte)
	for archID, arch := range supportedArchitectures {
		result[archID] = arch.ELFMagic
	}
	return result
}

// Currently no refence
func GetArchELFMagic(archID int) []byte {
	if arch, exists := supportedArchitectures[archID]; exists {
		return arch.ELFMagic
	}
	return nil
}

// Currently no refence
func GetArchELFMask(archID int) []byte {
	if arch, exists := supportedArchitectures[archID]; exists {
		return arch.ELFMask
	}
	return nil
}

func GetArchNameBinfmt(arch int) string {
	if arch == NotSpecifiedArchID {
		logrus.Warnf("Getting arch name for not specified architecture")
		return ""
	}
	if archObj, exists := supportedArchitectures[arch]; exists {
		return archObj.NameBinfmt
	}
	return ""
}

func GetArchNameOCI(arch int) string {
	if arch == NotSpecifiedArchID {
		logrus.Warnf("Getting arch name for not specified architecture")
		return ""
	}
	if archObj, exists := supportedArchitectures[arch]; exists {
		return archObj.NameOCI
	}
	return ""
}

func GetArchitecture(archID int) (Architecture, bool) {
	arch, exists := supportedArchitectures[archID]
	return arch, exists
}

func ImageReferenceGetArchFromTag(image string) int {
	tag := utils.ImageReferenceGetTag(image)

	if tag == "" {
		return NotSpecifiedArchID
	}

	i := strings.LastIndexByte(tag, '-')
	if i == -1 {
		return NotSpecifiedArchID
	}

	archInTag := tag[i+1:]

	for archID, arch := range supportedArchitectures {
		if arch.NameBinfmt == archInTag {
			return archID
		}
	}

	return NotSpecifiedArchID
}

func IsArchSupported(archID int, inContainer bool) (string, error) {
	archName := GetArchNameBinfmt(archID)
	archNameDebug := GetArchNameOCI(archID)
	logrus.Debugf("Checking QEMU emulation support for architecture %s", archNameDebug)

	inContainerPathPrefix := ""

	if inContainer {
		inContainerPathPrefix = "/run/host"
	}

	qemuBinaryPossiblePaths := []string{
		fmt.Sprintf("%s/usr/bin/qemu-%s-static", inContainerPathPrefix, archName),
		fmt.Sprintf("%s/usr/bin/qemu-%s", inContainerPathPrefix, archName),
	}

	qemuBinfmtPossiblePaths := []string{
		fmt.Sprintf("%s/proc/sys/fs/binfmt_misc/qemu-%s", inContainerPathPrefix, archName),
		fmt.Sprintf("%s/proc/sys/fs/binfmt_misc/qemu-%s-static", inContainerPathPrefix, archName),
	}

	qemuBinaryExists := false
	foundInterpreterPath := ""
	for _, qemuPath := range qemuBinaryPossiblePaths {
		if isStaticELF := isStaticallyLinkedELF(qemuPath); isStaticELF {
			qemuBinaryExists = true
			foundInterpreterPath = qemuPath
			break
		}
	}

	if !qemuBinaryExists {
		err := fmt.Errorf("The host system does not have the required support: No %s statically linked QEMU emulator binary found in '/usr/bin/'", archNameDebug)
		return "", err
	}

	for _, binfmtPath := range qemuBinfmtPossiblePaths {
		if utils.PathExists(binfmtPath) {
			logrus.Debugf("Architecture %s is supported", archName)
			return foundInterpreterPath, nil
		}
	}

	err := fmt.Errorf("The host system does not have the required support: No %s binfmt_misc registration found", archNameDebug)
	return "", err
}

func isStaticallyLinkedELF(filePath string) bool {
	if !utils.PathExists(filePath) {
		logrus.Debugf("File '%s' does not exist\n", filePath)
		return false
	}

	f, err := elf.Open(filePath)
	if err != nil {
		logrus.Debugf("File '%s' is not an ELF file\n", filePath)
		return false
	}
	defer f.Close()

	// Check for PT_INTERP program header
	for _, prog := range f.Progs {
		if prog.Type == elf.PT_INTERP {
			// Has interpreter = dynamically linked
			logrus.Debugf("File '%s' is dynamically linked\n", filePath)
			return false
		}
	}

	// No interpreter = statically linked
	return true
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
