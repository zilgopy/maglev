package maglev

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"testing"
)

func Test(t *testing.T) {

	backends := make([]backend, 5)

	for i := range backends {

		backends[i] = backend{
			IP:     net.ParseIP("192.168.1." + strconv.Itoa(i+1)),
			Weight: i,
		}
	}
	fmt.Println(backends)
	fmt.Println("initialize")
	m, err := NewMaglev(backends)

	if err != nil {
		log.Fatalf("not able to create new maglev instance , error : %v", err)
	}

	print(m.weights)

	look := m.Lookup
	kv := make(map[int]int)

	for _, v := range m.Lookup {

		count, ok := kv[v]
		if !ok {
			kv[v] = 1
			continue
		}
		kv[v] = count + 1

	}
	fmt.Println(kv)

	m.RemoveNode(net.ParseIP("192.168.1.1"))

	look1 := m.Lookup
	count := 0
	for i, v := range look {
		if v == 0 && look[i] != look1[i] {
			count++

		}

	}
	fmt.Println(count)
}
