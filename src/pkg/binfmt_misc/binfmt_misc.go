/*
 * Copyright © 2019 – 2025 Red Hat Inc.
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

package binfmt_misc

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/containers/toolbox/pkg/architecture"
	"github.com/containers/toolbox/pkg/shell"
	"github.com/containers/toolbox/pkg/utils"
	"github.com/sirupsen/logrus"
)

type Registration struct {
	Status      bool
	Name        string
	MagicType   string
	Offset      string
	Magic       []byte
	Mask        []byte
	Interpreter string
	Flags       string
}

const (
	defaultOffset  = "0"
	defaultFlags   = "FC"
	binfmtMiscPath = "/proc/sys/fs/binfmt_misc"
)

// Add this at the package level in binfmt_misc.go (after imports, before functions)
var defaultRegistrations = map[int]Registration{
	architecture.AARCH64ArchID: {
		Name:        "qemu-aarch64",
		MagicType:   "M",
		Offset:      defaultOffset,
		Magic:       architecture.GetArchELFMagic(architecture.AARCH64ArchID),
		Mask:        architecture.GetArchELFMask(architecture.AARCH64ArchID),
		Interpreter: "",
		Flags:       defaultFlags,
	},
	architecture.PPC64LEArchID: {
		Name:        "qemu-ppc64le",
		MagicType:   "M",
		Offset:      defaultOffset,
		Magic:       architecture.GetArchELFMagic(architecture.PPC64LEArchID),
		Mask:        architecture.GetArchELFMask(architecture.PPC64LEArchID),
		Interpreter: "",
		Flags:       defaultFlags,
	},
	architecture.X86_64ArchID: {
		Name:        "qemu-x86_64",
		MagicType:   "M",
		Offset:      defaultOffset,
		Magic:       architecture.GetArchELFMagic(architecture.X86_64ArchID),
		Mask:        architecture.GetArchELFMask(architecture.X86_64ArchID),
		Interpreter: "",
		Flags:       defaultFlags,
	},
}

func (r *Registration) register() error {
	logrus.Debugf("Registering binfmt_misc for %s", r.Name)

	// if err := r.Validate(); err != nil {
	// 	return fmt.Errorf("registration validation failed: %w", err)
	// }

	regString := r.buildRegistrationString()
	logrus.Debugf("Registration string: %s", regString)

	if err := os.WriteFile("/proc/sys/fs/binfmt_misc/register", []byte(regString), 0200); err != nil {
		return fmt.Errorf("failed to register binfmt_misc handler: %w", err)
	}

	// return r.Verify()
	return nil
}

func (r *Registration) fixInterpreterPath() {
	if !strings.HasPrefix(r.Interpreter, "/run/host/") {
		r.Interpreter = filepath.Join("/run/host", r.Interpreter)
	}
}

func (r *Registration) buildRegistrationString() string {
	return fmt.Sprintf(":%s:%s:%s:%s:%s:%s:%s",
		r.Name, r.MagicType, r.Offset,
		bytesToEscapedString(r.Magic),
		bytesToEscapedString(r.Mask),
		r.Interpreter, r.Flags)
}

func bytesToEscapedString(bytes []byte) string {
	var result strings.Builder
	for _, b := range bytes {
		result.WriteString(fmt.Sprintf("\\x%02x", b))
	}
	return result.String()
}

func MountBinfmtMisc() error {
	args := []string{
		"binfmt_misc",
		"-t",
		"binfmt_misc",
		"/proc/sys/fs/binfmt_misc",
	}

	var stdout bytes.Buffer

	if err := shell.Run("mount", nil, &stdout, nil, args...); err != nil {
		return fmt.Errorf("failed to mount binfmt_misc: %w", err)
	}

	logrus.Debugf("Result of mount command: %s", stdout.String())

	return nil
}

func RegisterBinfmtMisc(archID int, interpreterPath string) error {
	reg := getHardcodedRegistration(archID, interpreterPath)
	if reg == nil {
		logrus.Debugf("Could not find binfmt_misc registration for: %s", architecture.GetArchName(archID))
		return fmt.Errorf("no hardcoded registration available for architecture %s", architecture.GetArchName(archID))

		// TODO: Fallback to parsing the values from the host registration file??
		//			How to provide the path to the host registration file??

		//logrus.Debug("Trying to read registration from the host file system as fallback")
		//
		// Fallback to parsing the values from the host registration file
		// reg, err := binfmt_misc.GetRegistration(archID)
		// if err != nil {
		// 	return fmt.Errorf("no hardcoded registration available for architecture %s", arch)
		// }
	}

	reg.fixInterpreterPath()

	if err := reg.register(); err != nil {
		return err
	}

	return nil
}

// TODO
func getHardcodedRegistration(archID int, interpreterPath string) *Registration {
	if reg, exists := defaultRegistrations[archID]; exists {
		regCopy := reg // Create a copy to avoid returning pointer to map value
		regCopy.Interpreter = interpreterPath
		return &regCopy
	}
	// TODO: return an error instead of nil?
	return nil
}

// ---------------------- Parsing registration from binfmt_misc -------------------------

// The path should be like "/run/host/proc/sys/fs/binfmt_misc/qemu-aarch64"
func GetRegistration(archID int) (*Registration, error) {
	defaultReg, exists := defaultRegistrations[archID]
	if !exists {
		return nil, fmt.Errorf("no information available for architecture %s", architecture.GetArchName(archID))
	}

	name := defaultReg.Name
	filePath := fmt.Sprintf("/run/host/proc/sys/fs/binfmt_misc/%s", name)

	logrus.Debugf("Reading binfmt_misc registration from %s", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	fields, err := parseBinfmtFields(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	reg, err := buildRegistrationFromFields(name, fields)
	if err != nil {
		return nil, fmt.Errorf("failed to build registration from %s: %w", filePath, err)
	}

	return reg, nil
}

// func GetSupportedArchitectures() (map[int]bool, error) {
// 	if !utils.PathExists(binfmtMiscPath) {
// 		err := fmt.Errorf("Error: binfmt_misc not available on this system. Path does not exist: %s", binfmtMiscPath)
// 		return nil, err
// 	}

// 	entries, err := os.ReadDir(binfmtMiscPath)
// 	if err != nil {
// 		err := fmt.Errorf("Error reading %s: %w\n", binfmtMiscPath, err)
// 		return nil, err
// 	}

// 	supportedArchs := make(map[int]bool)

// 	for _, entry := range entries {
// 		name := entry.Name()

// 		if name == "register" || name == "status" {
// 			continue
// 		}

// 		entryPath := filepath.Join(binfmtMiscPath, name)

// 		data, err := os.ReadFile(entryPath)
// 		if err != nil {
// 			logrus.Debugf("Error reading file '%s': %v\n", name, err)
// 			continue
// 		}

// 		fields, err := parseBinfmtFields(string(data))
// 		if err != nil {
// 			logrus.Debugf("failed to parse %s: %v", entryPath, err)
// 			continue
// 		}

// 		reg, err := buildRegistrationFromFields(name, fields)
// 		if err != nil {
// 			logrus.Debugf("failed to build registration from %s: %v", entryPath, err)
// 			continue
// 		}

// 		matchedArchID := architecture.NotSpecifiedArchID

// 		for archID, magic := range architecture.GetArchELFMagicAll() {
// 			if bytes.Equal(magic, reg.Magic) {
// 				matchedArchID = archID
// 			}
// 		}

// 		if matchedArchID == architecture.NotSpecifiedArchID {
// 			continue
// 		}

// 		if reg.Status && strings.Contains(reg.Interpreter, "qemu") {
// 			supportedArchs[matchedArchID] = true
// 		}
// 	}

// 	return supportedArchs, nil
// }

func IsArchSupported(archID int) (bool, error) {
	if !utils.PathExists(binfmtMiscPath) {
		err := fmt.Errorf("Error: binfmt_misc not available on this system. Path does not exist: %s", binfmtMiscPath)
		return false, err
	}

	entries, err := os.ReadDir(binfmtMiscPath)
	if err != nil {
		err := fmt.Errorf("Error reading %s: %w\n", binfmtMiscPath, err)
		return false, err
	}

	archIDSupported := false

	for _, entry := range entries {
		name := entry.Name()

		if name == "register" || name == "status" {
			continue
		}

		entryPath := filepath.Join(binfmtMiscPath, name)

		data, err := os.ReadFile(entryPath)
		if err != nil {
			logrus.Debugf("Error reading file '%s': %v\n", name, err)
			continue
		}

		fields, err := parseBinfmtFields(string(data))
		if err != nil {
			logrus.Debugf("failed to parse %s: %v", entryPath, err)
			continue
		}

		reg, err := buildRegistrationFromFields(name, fields)
		if err != nil {
			logrus.Debugf("failed to build registration from %s: %v", entryPath, err)
			continue
		}

		if !bytes.Equal(architecture.GetArchELFMagic(archID), reg.Magic) {
			continue
		}

		if reg.Status && strings.Contains(reg.Interpreter, "qemu") {
			archIDSupported = true
			break
		}
	}

	return archIDSupported, nil
}

func parseBinfmtFields(content string) (map[string]string, error) {
	fields := make(map[string]string)
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var key, value string
		if line == "enabled" || line == "disabled" {
			key = "status"
			value = line
		} else if strings.Contains(line, ":") {
			// parsing format "flags: FC"
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key = strings.TrimSpace(parts[0])
				value = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, " ") {
			// parsing format "interpreter /usr/bin/qemu-aarch64-static"
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				key = strings.TrimSpace(parts[0])
				value = strings.TrimSpace(parts[1])
			}
		}

		if key == "" {
			return nil, fmt.Errorf("invalid field format on line %d: %s", lineNum+1, line)
		}

		fields[key] = value
	}

	return fields, nil
}

func buildRegistrationFromFields(name string, fields map[string]string) (*Registration, error) {
	reg := &Registration{
		Name: name,
	}

	if status, exists := fields["status"]; exists && status == "enabled" {
		reg.Status = true
	} else {
		reg.Status = false
	}

	if err := validateAndSetInterpreter(reg, fields); err != nil {
		return nil, err
	}

	if _, exists := fields["magic"]; exists {
		reg.MagicType = "M"

		if err := validateAndSetMagic(reg, fields); err != nil {
			return nil, err
		}

		if err := validateAndSetMask(reg, fields); err != nil {
			return nil, err
		}
		// } else if _, exists := fields["extension"]; exists {
		// 	reg.MagicType = "E"
		// 	if err := validateAndSetExtension(reg, fields); err != nil {
		// 		return nil, err
		// 	}
	} else {
		return nil, errors.New("missing required field: magic or extension")
	}

	setOptionalFields(reg, fields)

	return reg, nil
}

func validateAndSetInterpreter(reg *Registration, fields map[string]string) error {
	interpreter, exists := fields["interpreter"]
	if !exists {
		return errors.New("missing required field: interpreter")
	}

	if interpreter == "" {
		return errors.New("interpreter field is empty")
	}

	if !filepath.IsAbs(interpreter) {
		return fmt.Errorf("interpreter must be an absolute path, got: %s", interpreter)
	}

	reg.Interpreter = interpreter
	return nil
}

func validateAndSetMagic(reg *Registration, fields map[string]string) error {
	magic, exists := fields["magic"]
	if !exists {
		return errors.New("missing required field: magic")
	}

	if magic == "" {
		return errors.New("magic field is empty")
	}

	magicByte, err := hex.DecodeString(magic)
	if err != nil {
		return err
	}

	reg.Magic = magicByte
	return nil
}

func validateAndSetMask(reg *Registration, fields map[string]string) error {
	mask, exists := fields["mask"]
	if !exists {
		return errors.New("missing required field: mask")
	}

	if mask == "" {
		return errors.New("mask field is empty")
	}

	maskByte, err := hex.DecodeString(mask)
	if err != nil {
		return err
	}

	reg.Mask = maskByte
	return nil
}

func setOptionalFields(reg *Registration, fields map[string]string) {
	if offset, exists := fields["offset"]; exists && offset != "" {
		reg.Offset = offset
	} else {
		reg.Offset = "0"
	}

	if flags, exists := fields["flags"]; exists && flags != "" {
		if strings.Contains(flags, "C") {
			reg.Flags = flags
		} else {
			reg.Flags = flags + "C"
		}
	} else {
		reg.Flags = "FC"
	}
}

// func validateHexString(hexStr string) error {
// 	cleaned := strings.ReplaceAll(hexStr, " ", "")

// 	if len(cleaned)%2 != 0 {
// 		return errors.New("hex string must have even length")
// 	}

// 	for i, char := range cleaned {
// 		if !((char >= '0' && char <= '9') ||
// 			(char >= 'a' && char <= 'f') ||
// 			(char >= 'A' && char <= 'F')) {
// 			return fmt.Errorf("invalid hex character '%c' at position %d", char, i)
// 		}
// 	}

// 	return nil
// }
