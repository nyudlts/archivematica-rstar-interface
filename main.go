package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var packageptr = flag.String("package","", "Location of Package to be transferred")

func main() {
	//parse flags
	flag.Parse()

	//check if IP exists
	err := fileExists(*packageptr); if err != nil {
		panic(err)
	}

	//get the path of the transfer-info.txt file
	transferInfo, err := getTransferInfo(*packageptr)
	if err != nil {
		panic(err)
	}

	//check that the transfer-info.txt is present
	err = fileExists(transferInfo); if err != nil {
		panic(err)
	}

	//check that the bag-info.txt is present
	bagInfo := filepath.Join(*packageptr, "bag-info.txt")
	err = fileExists(bagInfo); if err != nil {
		panic(err)
	}

	//append the contents of transfer-info.txt to bag-info.txt
	appendTransfeInfoToBagInfo(transferInfo, bagInfo)

	//remove bag-info.txt from tagmanifest-sha256.txt

	//get sha256 as HexString for bag-info.txt

	//write sha256 entry to tagmanifest-sha256.txt

	//validate bag

	//run uploader
}

func fileExists(ip string) error {
	if _, err := os.Stat(ip); os.IsNotExist(err) {
		return err
	}
	return nil
}

func getTransferInfo(ip string) (string, error) {
	transfers := filepath.Join(ip, "data", "objects", "metadata", "transfers")
	var transferinfo string
	err := filepath.Walk(transfers,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == "transfer-info.txt" {
				transferinfo = path
			}
			return nil
		})

	if err != nil {
		return "", err
	}

	if transferinfo != "" {
		return transferinfo, nil
	}

	return "", fmt.Errorf("no transfer-info.txt found in IP")

}

func appendTransfeInfoToBagInfo(transferinfo string, baginfo string) error {
	transferInfoContents, err := ioutil.ReadFile(transferinfo)
	if err != nil {
		return err
	}
	bagInfoFile, err := os.OpenFile(baginfo, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer bagInfoFile.Close()

	_, err = bagInfoFile.Write(transferInfoContents); if err != nil {
		return err
	}
	return nil
}