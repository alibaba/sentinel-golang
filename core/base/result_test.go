package base

import (
	"fmt"
	"strconv"
	"testing"
)

/*
BenchmarkBlockType_String/Switch
BenchmarkBlockType_String/Switch-4         	633041816	         1.89 ns/op
BenchmarkBlockType_String/Slice
BenchmarkBlockType_String/Slice-4          	1000000000	         0.387 ns/op
BenchmarkBlockType_String/Map
BenchmarkBlockType_String/Map-4            	55307811	        20.0 ns/op
*/

func BenchmarkBlockType_String(b *testing.B) {
	b.ReportAllocs()

	//BlockTypeNew1 := BlockType(6)
	//RegistryBlockType(BlockTypeNew1, "new1")
	BlockTypeTmp := BlockTypeFlow

	b.Run("Switch", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			BlockTypeTmp.stringSwitch()
		}
	})
	b.Run("Slice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			BlockTypeTmp.string()
		}
	})
	b.Run("Map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			BlockTypeTmp.stringMap()
		}
	})
}

/*
BenchmarkBlockType_Registry_String/Slice-4         	1000000000	         1.16 ns/op
*/
func BenchmarkBlockType_Registry_String(b *testing.B) {
	b.ReportAllocs()

	BlockTypeNew1 := BlockType(6)
	RegistryBlockType(BlockTypeNew1, "new1")
	BlockTypeTmp := BlockTypeNew1

	b.Run("Registry_String", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			BlockTypeTmp.string()
		}
	})
}

func (t BlockType) stringSwitch() string {
	switch t {
	case BlockTypeUnknown:
		return "Unknown"
	case BlockTypeFlow:
		return "FlowControl"
	case BlockTypeIsolation:
		return "BlockTypeIsolation"
	case BlockTypeCircuitBreaking:
		return "CircuitBreaking"
	case BlockTypeSystemFlow:
		return "System"
	case BlockTypeHotSpotParamFlow:
		return "HotSpotParamFlow"
	default:
		return fmt.Sprintf("%d", t)
	}
}

var (
	blockTypeNames = []string{
		BlockTypeUnknown:          "Unknown",
		BlockTypeFlow:             "FlowControl",
		BlockTypeIsolation:        "BlockTypeIsolation",
		BlockTypeCircuitBreaking:  "CircuitBreaking",
		BlockTypeSystemFlow:       "System",
		BlockTypeHotSpotParamFlow: "HotSpotParamFlow",
	}
	blockTypeErr = fmt.Errorf("block type err")
)

func (t BlockType) stringMap() string {
	name, ok := blockTypeMap[t]
	if ok {
		return name
	}
	return fmt.Sprintf("%d", t)
}

func (t BlockType) string() string {
	if int(t) >= len(blockTypeNames) {
		return strconv.Itoa(int(t))
	}
	return blockTypeNames[t]
}

func TestRegistryBlockType(t *testing.T) {
	type args struct {
		blockType BlockType
		desc      string
	}

	var (
		New1BlockType = BlockType(100)
		New2BlockType = BlockType(200)
	)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "New1",
			args: struct {
				blockType BlockType
				desc      string
			}{blockType: New1BlockType, desc: "New1"},
			wantErr: false,
		},
		{
			name: "New2",
			args: struct {
				blockType BlockType
				desc      string
			}{blockType: New2BlockType, desc: "New2"},
			wantErr: false,
		},
		{
			name: "existed",
			args: struct {
				blockType BlockType
				desc      string
			}{blockType: BlockTypeHotSpotParamFlow, desc: "BlockTypeHotSpotParamFlow"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RegistryBlockType(tt.args.blockType, tt.args.desc); (err != nil) != tt.wantErr {
				t.Errorf("RegistBlockType() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				if tt.args.blockType.String() != tt.args.desc {
					t.Errorf("RegistBlockType() string :%s not equal to desc:%s", tt.args.blockType.String(), tt.args.desc)
				}
			}
		})
	}
}
