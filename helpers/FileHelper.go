package helpers

import "io/ioutil"


// Function for helping loading files (especially used for loading html files)

func LoadFile(fileName string) (string, error) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
