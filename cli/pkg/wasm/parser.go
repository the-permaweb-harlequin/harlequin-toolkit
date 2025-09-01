package wasm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// WasmExport represents an exported function, memory, global, or table
type WasmExport struct {
	Name string
	Type string // "function", "memory", "global", "table"
	Index uint32
}

// WasmImport represents an imported function, memory, global, or table
type WasmImport struct {
	Module string
	Name   string
	Type   string // "function", "memory", "global", "table"
}

// WasmInfo contains metadata extracted from a WASM binary
type WasmInfo struct {
	StackSize       uint32
	InitialMemory   uint32
	MaxMemory       uint32
	Target          string // "wasm32" or "wasm64"
	CustomSections map[string][]byte
	Exports         []WasmExport
	Imports         []WasmImport
	FunctionCount   uint32
	GlobalCount     uint32
	TableCount      uint32
}

// WASM binary format constants
const (
	WASM_MAGIC   = 0x6d736100 // "\0asm"
	WASM_VERSION = 0x01000000 // version 1

	// Section types
	SECTION_TYPE     = 1
	SECTION_IMPORT   = 2
	SECTION_FUNCTION = 3
	SECTION_TABLE    = 4
	SECTION_MEMORY   = 5
	SECTION_GLOBAL   = 6
	SECTION_EXPORT   = 7
	SECTION_START    = 8
	SECTION_ELEMENT  = 9
	SECTION_CODE     = 10
	SECTION_DATA     = 11
	SECTION_CUSTOM   = 0
)

// ParseWasmBinary extracts metadata from a WASM binary
func ParseWasmBinary(data []byte) (*WasmInfo, error) {
	reader := bytes.NewReader(data)
	info := &WasmInfo{
		CustomSections: make(map[string][]byte),
		Exports:        make([]WasmExport, 0),
		Imports:        make([]WasmImport, 0),
		Target:         "wasm32", // default assumption
	}

	// Read magic number
	var magic uint32
	if err := binary.Read(reader, binary.LittleEndian, &magic); err != nil {
		return nil, fmt.Errorf("failed to read magic number: %w", err)
	}
	if magic != WASM_MAGIC {
		return nil, fmt.Errorf("invalid WASM magic number: 0x%x", magic)
	}

	// Read version
	var version uint32
	if err := binary.Read(reader, binary.LittleEndian, &version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}
	// Accept version 1 in various formats
	if version != WASM_VERSION && version != 0x01 && version != 1 {
		return nil, fmt.Errorf("unsupported WASM version: 0x%x", version)
	}

	// Parse sections
	for {
		sectionType, err := readByte(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read section type: %w", err)
		}

		sectionSize, err := readLEB128(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read section size: %w", err)
		}

		sectionData := make([]byte, sectionSize)
		if _, err := io.ReadFull(reader, sectionData); err != nil {
			return nil, fmt.Errorf("failed to read section data: %w", err)
		}

		switch sectionType {
		case SECTION_IMPORT:
			if err := parseImportSection(sectionData, info); err != nil {
				return nil, fmt.Errorf("failed to parse import section: %w", err)
			}
		case SECTION_FUNCTION:
			if err := parseFunctionSection(sectionData, info); err != nil {
				return nil, fmt.Errorf("failed to parse function section: %w", err)
			}
		case SECTION_TABLE:
			if err := parseTableSection(sectionData, info); err != nil {
				return nil, fmt.Errorf("failed to parse table section: %w", err)
			}
		case SECTION_MEMORY:
			if err := parseMemorySection(sectionData, info); err != nil {
				return nil, fmt.Errorf("failed to parse memory section: %w", err)
			}
		case SECTION_GLOBAL:
			if err := parseGlobalSection(sectionData, info); err != nil {
				return nil, fmt.Errorf("failed to parse global section: %w", err)
			}
		case SECTION_EXPORT:
			if err := parseExportSection(sectionData, info); err != nil {
				return nil, fmt.Errorf("failed to parse export section: %w", err)
			}
		case SECTION_CUSTOM:
			if err := parseCustomSection(sectionData, info); err != nil {
				return nil, fmt.Errorf("failed to parse custom section: %w", err)
			}
		}
	}

	return info, nil
}

// parseMemorySection extracts memory configuration
func parseMemorySection(data []byte, info *WasmInfo) error {
	reader := bytes.NewReader(data)

	// Read number of memory entries
	count, err := readLEB128(reader)
	if err != nil {
		return err
	}

	if count > 0 {
		// Read memory limits
		limits, err := readByte(reader)
		if err != nil {
			return err
		}

		// Read initial memory size (in 64KB pages)
		initial, err := readLEB128(reader)
		if err != nil {
			return err
		}
		info.InitialMemory = initial * 65536 // Convert pages to bytes

		// If limits indicate maximum is present
		if limits == 0x01 {
			maximum, err := readLEB128(reader)
			if err != nil {
				return err
			}
			info.MaxMemory = maximum * 65536 // Convert pages to bytes

			// Infer target architecture from maximum memory pages
			info.Target = InferTargetFromMemory(maximum)
		} else {
			// No maximum specified, infer from initial memory size
			// If initial memory is close to or at WASM32 limit, likely WASM32
			// Otherwise, use initial as a hint but default to WASM32
			info.Target = InferTargetFromMemory(initial)
		}
	}

	return nil
}

// parseImportSection extracts import information
func parseImportSection(data []byte, info *WasmInfo) error {
	reader := bytes.NewReader(data)

	// Read number of imports
	count, err := readLEB128(reader)
	if err != nil {
		return err
	}

	for i := uint32(0); i < count; i++ {
		// Read module name
		moduleLen, err := readLEB128(reader)
		if err != nil {
			return err
		}
		moduleBytes := make([]byte, moduleLen)
		if _, err := io.ReadFull(reader, moduleBytes); err != nil {
			return err
		}

		// Read import name
		nameLen, err := readLEB128(reader)
		if err != nil {
			return err
		}
		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(reader, nameBytes); err != nil {
			return err
		}

		// Read import type
		importType, err := readByte(reader)
		if err != nil {
			return err
		}

		var typeStr string
		switch importType {
		case 0x00:
			typeStr = "function"
			// Skip function type index
			if _, err := readLEB128(reader); err != nil {
				return err
			}
		case 0x01:
			typeStr = "table"
			// Skip table type (element type + limits)
			if _, err := readByte(reader); err != nil {
				return err
			}
			if _, err := readLEB128(reader); err != nil { // limits flag
				return err
			}
			if _, err := readLEB128(reader); err != nil { // initial
				return err
			}
		case 0x02:
			typeStr = "memory"
			// Skip memory limits
			if _, err := readLEB128(reader); err != nil { // limits flag
				return err
			}
			if _, err := readLEB128(reader); err != nil { // initial
				return err
			}
		case 0x03:
			typeStr = "global"
			// Skip global type
			if _, err := readByte(reader); err != nil { // value type
				return err
			}
			if _, err := readByte(reader); err != nil { // mutability
				return err
			}
		}

		info.Imports = append(info.Imports, WasmImport{
			Module: string(moduleBytes),
			Name:   string(nameBytes),
			Type:   typeStr,
		})
	}

	return nil
}

// parseFunctionSection counts functions
func parseFunctionSection(data []byte, info *WasmInfo) error {
	reader := bytes.NewReader(data)

	// Read number of functions
	count, err := readLEB128(reader)
	if err != nil {
		return err
	}

	info.FunctionCount = count
	return nil
}

// parseTableSection counts tables
func parseTableSection(data []byte, info *WasmInfo) error {
	reader := bytes.NewReader(data)

	// Read number of tables
	count, err := readLEB128(reader)
	if err != nil {
		return err
	}

	info.TableCount = count
	return nil
}

// parseGlobalSection counts globals
func parseGlobalSection(data []byte, info *WasmInfo) error {
	reader := bytes.NewReader(data)

	// Read number of globals
	count, err := readLEB128(reader)
	if err != nil {
		return err
	}

	info.GlobalCount = count
	return nil
}

// parseExportSection extracts export information
func parseExportSection(data []byte, info *WasmInfo) error {
	reader := bytes.NewReader(data)

	// Read number of exports
	count, err := readLEB128(reader)
	if err != nil {
		return err
	}

	for i := uint32(0); i < count; i++ {
		// Read export name
		nameLen, err := readLEB128(reader)
		if err != nil {
			return err
		}
		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(reader, nameBytes); err != nil {
			return err
		}

		// Read export type
		exportType, err := readByte(reader)
		if err != nil {
			return err
		}

		// Read export index
		index, err := readLEB128(reader)
		if err != nil {
			return err
		}

		var typeStr string
		switch exportType {
		case 0x00:
			typeStr = "function"
		case 0x01:
			typeStr = "table"
		case 0x02:
			typeStr = "memory"
		case 0x03:
			typeStr = "global"
		default:
			typeStr = "unknown"
		}

		info.Exports = append(info.Exports, WasmExport{
			Name:  string(nameBytes),
			Type:  typeStr,
			Index: index,
		})
	}

	return nil
}

// parseCustomSection extracts custom section data
func parseCustomSection(data []byte, info *WasmInfo) error {
	reader := bytes.NewReader(data)

	// Read section name
	nameLen, err := readLEB128(reader)
	if err != nil {
		return err
	}

	nameBytes := make([]byte, nameLen)
	if _, err := io.ReadFull(reader, nameBytes); err != nil {
		return err
	}
	name := string(nameBytes)

	// Read remaining data
	remaining := make([]byte, len(data)-int(nameLen)-getLEB128Size(nameLen))
	if _, err := io.ReadFull(reader, remaining); err != nil && err != io.EOF {
		return err
	}

	info.CustomSections[name] = remaining

	// Try to extract stack size from known custom sections
	if name == "target_features" || name == "linking" {
		// These sections might contain stack size info
		// This is a simplified approach - real parsing would be more complex
		if len(remaining) >= 4 {
			stackSize := binary.LittleEndian.Uint32(remaining[:4])
			if stackSize > 1024 && stackSize < 100*1024*1024 { // Reasonable range
				info.StackSize = stackSize
			}
		}
	}

	return nil
}

// readByte reads a single byte from the reader
func readByte(reader *bytes.Reader) (byte, error) {
	var b byte
	err := binary.Read(reader, binary.LittleEndian, &b)
	return b, err
}

// readLEB128 reads an unsigned LEB128 integer
func readLEB128(reader *bytes.Reader) (uint32, error) {
	var result uint32
	var shift uint
	for {
		b, err := readByte(reader)
		if err != nil {
			return 0, err
		}

		result |= (uint32(b) & 0x7F) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
		if shift >= 32 {
			return 0, fmt.Errorf("LEB128 overflow")
		}
	}
	return result, nil
}

// getLEB128Size returns the size in bytes of a LEB128 encoded number
func getLEB128Size(value uint32) int {
	if value == 0 {
		return 1
	}
	size := 0
	for value > 0 {
		value >>= 7
		size++
	}
	return size
}

// FormatMemorySize formats a memory size in bytes to a human-readable string
func FormatMemorySize(bytes uint32) string {
	if bytes == 0 {
		return "0"
	}

	kb := bytes / 1024
	if kb < 1024 {
		return fmt.Sprintf("%d KB", kb)
	}

	mb := kb / 1024
	if mb < 1024 {
		return fmt.Sprintf("%d MB", mb)
	}

	gb := mb / 1024
	return fmt.Sprintf("%d GB", gb)
}

// InferTargetFromMemory infers the target architecture (wasm32/wasm64) from memory page limits.
// WASM32 has a hard limit of 65,536 pages (64KB each = 4GB total address space).
// WASM64 can theoretically have larger memory limits beyond this constraint.
func InferTargetFromMemory(memoryPages uint32) string {
	// WASM32 maximum: 65,536 pages × 64KB = 4GB address space
	const WASM32_MAX_PAGES = 65536

	if memoryPages > WASM32_MAX_PAGES {
		return "wasm64"
	}
	return "wasm32"
}
