package main

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"regexp"
)

func getValueInEtcd(key string, adrr string) (string, error) {
	machines := []string{adrr}
	client := etcd.NewClient(machines)
	resp, err := client.Get(key, false, true)
	if err != nil {
		fmt.Println("err:", err)
		return "", nil
	} else {
		fmt.Println("Key:", resp.Node.Key, "Value:", resp.Node.Value)
		return resp.Node.Value, nil
	}
}

func matchRegex(oriKey string) (string, bool) {
	reg := regexp.MustCompile(`(?:).*docker`)
	valDocker := reg.FindAllString(oriKey, -1)
	if len(valDocker) == 0 {
		return oriKey, false
	} else {
		tempArray := []byte(valDocker[0])
		realKey := string(tempArray[:len(tempArray)-7])
		return realKey, true
	}

}
