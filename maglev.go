package maglev

import (
	"errors"
	"fmt"
	"net"
	"sort"

	"github.com/minio/highwayhash"
)

var KEY1 = []byte("Albert Einstein's IQ is Number 1")
var KEY2 = []byte("Ein8's IQ is 8 times of Einstein")

type backend struct {
	IP     net.IP
	Weight int
}

type Maglev struct {
	n           int
	ips         []net.IP
	weights     []int
	permutation [][]uint64
	Lookup      []int
}

const M = 65537

func NewMaglev(backends []backend) (*Maglev, error) {

	n := len(backends)
	if n == 0 {
		return nil, errors.New("backend list is empty")

	}
	maglev := new(Maglev)
	maglev.n = n
	var ips []net.IP
	var weights []int

	for _, backend := range backends {
		ips = append(ips, backend.IP)
		weights = append(weights, backend.Weight)
	}
	maglev.ips = ips
	maglev.weights = weights
	maglev.genPermutation()
	maglev.weightedPopulate()
	return maglev, nil
}

func (m *Maglev) calcWeights() []int {
	var total int

	for _, v := range m.weights {

		total += v
	}

	var allocated int
	w := make([]int, m.n)
	for j, v := range m.weights {

		weight := v * M / total
		w[j] = weight
		allocated += weight

	}

	indices := make([]int, m.n)
	for i := range m.weights {
		indices[i] = i
	}

	sort.Slice(indices, func(i, j int) bool {
		return m.weights[indices[i]] > m.weights[indices[j]]
	})
	i := 0
	for remains := M - allocated; remains > 0; remains-- {

		w[indices[i]] += 1
		i += 1

	}
	fmt.Println(w)

	return w

}

func (m *Maglev) genPermutation() {

	permutation := make([][]uint64, 0)

	for _, ip := range m.ips {

		offset := highwayhash.Sum64(ip, KEY1) % M
		skip := highwayhash.Sum64(ip, KEY2)%(M-1) + 1
		p := make([]uint64, M)
		var j uint64
		for j = 0; j < M; j++ {
			p[j] = (offset + j*skip) % M
		}
		permutation = append(permutation, p)

	}
	m.permutation = permutation

}

func (m *Maglev) weightedPopulate() {
	w := m.calcWeights()

	next := make([]int, m.n)

	entry := make([]int, M)

	for i := range entry {
		entry[i] = -1

	}

	n := 0

	for {

		for i := 0; i < m.n; i++ {

			if w[i] == 0 {
				continue
			}

			c := m.permutation[i][next[i]]

			for entry[c] >= 0 {
				next[i] = next[i] + 1
				c = m.permutation[i][next[i]]

			}
			entry[c] = i
			next[i] = next[i] + 1
			n++
			w[i]--
			if n == M {
				m.Lookup = entry
				return
			}

		}

	}

}

func (m *Maglev) AddNode(b backend) error {

	for _, v := range m.ips {

		if b.IP.Equal(v) {
			return errors.New("existing entries")
		}

	}

	if m.n > 99 {
		return errors.New("backend number reached to 100,rejected")

	}

	m.n += 1
	m.ips = append(m.ips, b.IP)
	m.weights = append(m.weights, b.Weight)
	m.genPermutation()
	m.weightedPopulate()
	return nil

}

func (m *Maglev) RemoveNode(ip net.IP) error {

	for i, v := range m.ips {

		if ip.Equal(v) {
			m.ips = append(m.ips[:i], m.ips[i+1:]...)
			m.n--
			m.weights = append(m.weights[:i], m.weights[i+1:]...)
			m.genPermutation()
			m.weightedPopulate()
			return nil

		}

	}
	return errors.New("ip not found")

}
