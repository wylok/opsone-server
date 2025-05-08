package kits

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func LenToSubNetMask(subnet int) string {
	var buff bytes.Buffer
	for i := 0; i < subnet; i++ {
		buff.WriteString("1")
	}
	for i := subnet; i < 32; i++ {
		buff.WriteString("0")
	}
	masker := buff.String()
	a, _ := strconv.ParseUint(masker[:8], 2, 64)
	b, _ := strconv.ParseUint(masker[8:16], 2, 64)
	c, _ := strconv.ParseUint(masker[16:24], 2, 64)
	d, _ := strconv.ParseUint(masker[24:32], 2, 64)
	resultMask := fmt.Sprintf("%v.%v.%v.%v", a, b, c, d)
	return resultMask
}

func ExcludeNetName(name string, extend []string) bool {
	result := true
	extend = append(extend, []string{"lo", "docker", "cni", "tunl", "vir", "cali", "flannel", "br", "vnet", "veth", "kube-ip"}...)
	for _, n := range extend {
		if strings.Contains(name, n) == true {
			result = false
			break
		}
	}
	return result
}
