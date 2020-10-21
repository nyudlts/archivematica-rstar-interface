package main

import (
	"bufio"
	"crypto/sha256"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

	//check that the tagmanifest=sha256.txt is present
	tagManifest := filepath.Join(*packageptr, "tagmanifest-sha256.txt")
	err = fileExists(tagManifest); if err != nil {
		panic(err)
	}

	//remove bag-info.txt from tagmanifest-sha256.txt
	removeBagInfoFromTagManifest(tagManifest)

	//get sha256 as HexString for bag-info.txt
	bagInfoSha, err := getSha256(bagInfo, "bag-info.txt")
	if err != nil {
		panic(err)
	}

	//write sha256 entry to tagmanifest-sha256.txt
	err = appendToFile(tagManifest, bagInfoSha); if err != nil {
		panic(err)
	}

	//validate tag manifest
	result, err := validate(tagManifest); if err != nil {
		panic(err)
	}

	if result != true {
		panic(fmt.Errorf("Tagmanifest validation failed"))
	}

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
	_, err = bagInfoFile.WriteString("\n" + string(transferInfoContents)); if err != nil {
		return err
	}
	return nil
}

func removeBagInfoFromTagManifest(tagmanifestFile string) {
	tags := []string{}
	tagmanifest, err := os.Open(tagmanifestFile); if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(tagmanifest)
	for scanner.Scan() {
		tags = append(tags, scanner.Text() + "\n")
	}

	if scanner.Err() != nil {
		panic(scanner.Err())
	}

	err = tagmanifest.Close()
	if err != nil {
		panic(err)
	}

	err = os.Remove(tagmanifestFile)
	if err != nil {
		panic(err)
	}

	tagmanifest, err = os.Create(tagmanifestFile)
	if err != nil {
		panic(err)
	}
	defer tagmanifest.Close()

	writer := bufio.NewWriter(tagmanifest)
	bag := regexp.MustCompile(`bag-info.txt`)
	for _, tag := range tags {
		if bag.MatchString(tag) != true {
			writer.WriteString(tag)
		}
	}
	writer.Flush()
}

func getSha256(file string, entry string) (string, error) {
	bagBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	bagInfoSha256 := fmt.Sprintf("%x %s", sha256.Sum256(bagBytes), entry)
	return bagInfoSha256, nil
}

func appendToFile(fileLoc string, message string) error {
	file, err := os.OpenFile(fileLoc, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	_, err = file.WriteString(message); if err != nil {
		return err
	}
	return nil
}

func validate(tagManifestLoc string) (bool, error) {
	tagManifest, err := os.Open(tagManifestLoc)
	if err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(tagManifest)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), " ")
		storedHash := tokens[0]
		filename := tokens[1]
		fileBytes, err := ioutil.ReadFile(filepath.Join(*packageptr, filename))
		if err != nil {
			return false, err
		}
		fileSha256 := fmt.Sprintf("%x", sha256.Sum256(fileBytes))
		if storedHash != fileSha256 {
			return false, nil
		}
	}
	return true, nil
}