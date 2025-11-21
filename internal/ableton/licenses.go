package ableton

import (
	"crypto/dsa"
	"fmt"
	"math/rand"
	"os"
)

func GenerateLicense(privateKey dsa.PrivateKey, hwid string, edition, version int, addonsLimit int) ([]string, error) {
	var results []string

	l1, err := generateLicenseSingle(privateKey, hwid, edition, version<<4)
	if err != nil {
		return nil, fmt.Errorf("generate license: %v", err)
	}
	results = append(results, *l1)

	for i := 0x40; i <= addonsLimit; i++ {
		l2, err := generateLicenseSingle(privateKey, hwid, i, 0x10)
		if err != nil {
			return nil, fmt.Errorf("generate license [%d]: %v", i, err)
		}
		results = append(results, *l2)
	}

	for i := 0x8000; i <= 0x80FF; i++ {
		l3, err := generateLicenseSingle(privateKey, hwid, i, 0x10)
		if err != nil {
			return nil, fmt.Errorf("generate license [%d]: %v", i, err)
		}
		results = append(results, *l3)
	}

	return results, nil
}

func WriteAuthorizationFile(license []string, fileName string) error {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, l := range license {
		if _, err := f.WriteString(l + "\n"); err != nil {
			return err
		}
	}

	return nil
}

func generateLicenseSingle(key dsa.PrivateKey, hwid string, edition, version int) (*string, error) {
	format := "%s,%02X,%02X,Standard,%s"
	serial := generateSerial()

	l := fmt.Sprintf(format, serial, edition, version, hwid)

	signature, err := signDSA(key, l)
	if err != nil {
		return nil, fmt.Errorf("sign license: %v", err)
	}

	l = fmt.Sprintf(format, serial, edition, version, signature)

	return &l, nil
}

func generateSerial() string {
	groups := []int{
		rand.Intn(0x1000) + 0x3000,
		rand.Intn(0x10000),
		rand.Intn(0x10000),
		rand.Intn(0x10000),
		rand.Intn(0x10000),
	}

	for i := range groups {
		groups[i] = fixGroupChecksum(i, groups[i])
	}

	d := overallChecksum(groups)

	return fmt.Sprintf("%04X-%04X-%04X-%04X-%04X-%04X", groups[0], groups[1], groups[2], groups[3], groups[4], d)
}

func fixGroupChecksum(groupNumber, n int) int {
	checksum := (n >> 4 & 0xf) ^
		(n >> 5 & 0x8) ^
		(n >> 9 & 0x7) ^
		(n >> 11 & 0xe) ^
		(n >> 15 & 0x1) ^
		groupNumber
	return (n & 0xfff0) | checksum
}

func overallChecksum(groups []int) int {
	r := 0
	for i := 0; i < 20; i++ {
		g, digit := i/4, i%4
		v := (groups[g] >> (digit * 8)) & 0xff
		r ^= v << 8
		for j := 0; j < 8; j++ {
			r <<= 1
			if r&0x10000 != 0 {
				r ^= 0x8005
			}
		}
	}
	return r & 0xffff
}
