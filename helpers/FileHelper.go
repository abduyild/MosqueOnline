package helpers

import "io/ioutil"

<<<<<<< HEAD
// Function for helping loading files (especially used for loading html files)
=======
>>>>>>> 81a31ff736a38c51807974c39203cc754ae74309
func LoadFile(fileName string) (string, error) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
