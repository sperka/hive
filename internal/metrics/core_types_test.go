package metrics

import (
	"reflect"
	"testing"
)

func TestRegistryCoreTypesUseLogicalIDs(t *testing.T) {
	// ioreg -a encodes device-tree data as base64. CPU entries may arrive out of order.
	registry := []byte(`<plist version="1.0"><array><dict><key>IORegistryEntryChildren</key><array>
		<dict><key>logical-cpu-id</key><integer>1</integer><key>cluster-type</key><data>UAAA</data></dict>
		<dict><key>logical-cpu-id</key><integer>0</integer><key>cluster-type</key><data>RQ==</data></dict>
		<dict><key>cluster-type</key><data>UA==</data></dict>
	</array></dict></array></plist>`)
	groups := []CoreGroup{{Name: "Performance", Count: 1}, {Name: "Efficiency", Count: 1}}
	if got := registryCoreTypes(registry, groups); !reflect.DeepEqual(got, map[int]string{0: "Efficiency", 1: "Performance"}) {
		t.Fatalf("core types by logical ID = %v", got)
	}
	groups[0].Count++
	if got := registryCoreTypes(registry, groups); len(got) != 0 {
		t.Fatalf("accepted incomplete topology: %v", got)
	}
}
