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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/containers/toolbox/pkg/shell"
	"github.com/containers/toolbox/pkg/utils"
	"github.com/sirupsen/logrus"
)

type Registration struct {
	Name        string
	MagicType   string
	Offset      string
	Magic       string
	Mask        string
	Interpreter string
	Flags       string
}

// Add this at the package level in binfmt_misc.go (after imports, before functions)
var defaultRegistrations = map[int]Registration{
	utils.ARM64ArchID: {
		Name:        "qemu-aarch64",
		MagicType:   "M",
		Offset:      "0",
		Magic:       "\\x7fELF\\x02\\x01\\x01\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x02\\x00\\xb7\\x00",
		Mask:        "\\xff\\xff\\xff\\xff\\xff\\xff\\xff\\x00\\xff\\xff\\xff\\xff\\xff\\xff\\xff\\xff\\xfe\\xff\\xff\\xff",
		Interpreter: "/run/host/usr/bin/qemu-aarch64-static",
		Flags:       "FC",
	},
	utils.PPC64LEArchID: {
		Name:        "qemu-ppc64le",
		MagicType:   "M",
		Offset:      "0",
		Magic:       "\\x7fELF\\x02\\x01\\x01\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x02\\x00\\x15\\x00",
		Mask:        "\\xff\\xff\\xff\\xff\\xff\\xff\\xff\\x00\\xff\\xff\\xff\\xff\\xff\\xff\\xff\\xff\\xfe\\xff\\xff\\x00",
		Interpreter: "/run/host/usr/bin/qemu-ppc64le-static",
		Flags:       "FC",
	},
	utils.X86_64ArchID: {
		Name:        "qemu-x86_64",
		MagicType:   "M",
		Offset:      "0",
		Magic:       "\\x7fELF\\x02\\x01\\x01\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x00\\x02\\x00\\x3e\\x00",
		Mask:        "\\xff\\xff\\xff\\xff\\xff\\xfe\\xfe\\x00\\xff\\xff\\xff\\xff\\xff\\xff\\xff\\xff\\xfe\\xff\\xff\\xff",
		Interpreter: "/run/host/usr/bin/qemu-x86_64-static",
		Flags:       "FC",
	},
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

func RegisterBinfmtMisc(archID int) error {
	reg := getHardcodedRegistration(archID)
	if reg == nil {
		logrus.Debugf("Could not find binfmt_misc registration for: %s", utils.GetArchName(archID))
		return fmt.Errorf("no hardcoded registration available for architecture %s", utils.GetArchName(archID))

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
func getHardcodedRegistration(archID int) *Registration {
	if reg, exists := defaultRegistrations[archID]; exists {
		regCopy := reg // Create a copy to avoid returning pointer to map value
		return &regCopy
	}
	// TODO: return an error instead of nil?
	return nil
}

// The path should be like "/run/host/proc/sys/fs/binfmt_misc/qemu-aarch64"
func GetRegistration(archID int) (*Registration, error) {
	defaultReg, exists := defaultRegistrations[archID]
	if !exists {
		return nil, fmt.Errorf("no information available for architecture %s", utils.GetArchName(archID))
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

func (r *Registration) register() error {
	logrus.Debugf("Registering binfmt_misc for %s", r.Name)

	// if err := r.Validate(); err != nil {
	// 	return fmt.Errorf("registration validation failed: %w", err)
	// }

	regString := r.BuildRegistrationString()
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

func (r *Registration) BuildRegistrationString() string {
	return fmt.Sprintf(":%s:%s:%s:%s:%s:%s:%s",
		r.Name, r.MagicType, r.Offset, r.Magic, r.Mask, r.Interpreter, r.Flags)
}

// func (r *Registration) Validate() error {
// 	if r.Name == "" {
// 		return errors.New("registration name cannot be empty")
// 	}

// 	if r.Interpreter == "" {
// 		return errors.New("interpreter cannot be empty")
// 	}

// 	if !filepath.IsAbs(r.Interpreter) {
// 		return fmt.Errorf("interpreter must be an absolute path, got: %s", r.Interpreter)
// 	}

// 	if r.Magic == "" {
// 		return errors.New("magic cannot be empty")
// 	}

// 	if r.Mask == "" {
// 		return errors.New("mask cannot be empty")
// 	}

// 	if r.MagicType != "M" && r.MagicType != "E" {
// 		return fmt.Errorf("invalid magic type: %s (must be M or E)", r.MagicType)
// 	}

// 	return nil
// }

// func (r *Registration) Verify() error {
// 	regFile := fmt.Sprintf("/proc/sys/fs/binfmt_misc/%s", r.Name)
// 	data, err := os.ReadFile(regFile)
// 	if err != nil {
// 		return fmt.Errorf("failed to verify registration by reading %s: %w", regFile, err)
// 	}

// 	content := string(data)
// 	if strings.Contains(content, "enabled") {
// 		logrus.Debugf("Successfully registered and enabled %s", r.Name)
// 		return nil
// 	}

// 	return fmt.Errorf("registration appears to have failed for %s", r.Name)
// }

func parseBinfmtFields(content string) (map[string]string, error) {
	fields := make(map[string]string)
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "enabled" || line == "disabled" {
			continue
		}

		var key, value string
		if strings.Contains(line, ":") {
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

	if err := validateHexString(magic); err != nil {
		return fmt.Errorf("invalid magic format: %w", err)
	}

	reg.Magic = hexToEscaped(magic)
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

	if err := validateHexString(mask); err != nil {
		return fmt.Errorf("invalid mask format: %w", err)
	}

	reg.Mask = hexToEscaped(mask)
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

func validateHexString(hexStr string) error {
	cleaned := strings.ReplaceAll(hexStr, " ", "")

	if len(cleaned)%2 != 0 {
		return errors.New("hex string must have even length")
	}

	for i, char := range cleaned {
		if !((char >= '0' && char <= '9') ||
			(char >= 'a' && char <= 'f') ||
			(char >= 'A' && char <= 'F')) {
			return fmt.Errorf("invalid hex character '%c' at position %d", char, i)
		}
	}

	return nil
}

func hexToEscaped(hexStr string) string {
	var result strings.Builder

	cleaned := strings.ReplaceAll(hexStr, " ", "")

	for i := 0; i < len(cleaned); i += 2 {
		if i+2 <= len(cleaned) {
			result.WriteString("\\x")
			result.WriteString(strings.ToLower(cleaned[i : i+2]))
		}
	}

	return result.String()
}
