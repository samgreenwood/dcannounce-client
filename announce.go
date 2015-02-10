package main

import "os"
import "os/exec"
import "fmt"
import "io/ioutil"
import "path"
import "strconv"
import "net/http"
import "net/url"
import "path/filepath"
import "strings"
import "github.com/robfig/config"

func main() {

	filename := os.Args[0];
	dirname := path.Dir(filename);
	abspath, _ := filepath.Abs(dirname);

	configFile := abspath + "/" + "announce.conf";

	c, err := config.ReadDefault(configFile);

	if err != nil {
		fmt.Println("Error loading configuration file " + configFile);
		os.Exit(0);
	}

	var site, _ = c.String("announce", "site");
	var announceUrl, _ = c.String("announce", "url");
	var tthPath, _ = c.String("announce", "tthsum");

	args := os.Args[1:]

	var ran = false;

	if len(args) == 8 {
		fmt.Println("SABnzbd mode");

		var downloadPath string = args[0];
		var processingStatus string = args[6];

		if processingStatus == "0" {
			fmt.Println("Processing completed successfully");

			// first we get the largest file in the directory as thats what we want to announce
			var largestFile = getLargestFile(downloadPath);

			// then we work out the full path to the file
			var filePath string = downloadPath + "/" + largestFile.Name();

			announce(site, announceUrl, filePath, tthPath);
		}

		ran = true;
	}

	if len(args) == 6 {
		fmt.Println("Sickbeard mode");

		var arg string = args[0];
		var filePath string = strings.Replace(arg, ",", "", -1);

		announce(site, announceUrl, filePath, tthPath);

		ran = true;
	}

	if ran == false {
		fmt.Println("Invalid number of arguments");
		os.Exit(0);
	}

}

func announce(site string, announceUrl string, filePath string, tthPath string) {
	file, _ := os.Stat(filePath);

	var filename string = file.Name();
	var size string = strconv.FormatInt(file.Size(), 10);
	var tth string = calculateTth(tthPath, filePath);
	var magnet string = fmt.Sprintf("magnet:?xt=urn:tree:tiger:%s&xl=%s&dn=%s", tth, size, url.QueryEscape(filename));

	fmt.Println("Sitename: " + site);
	fmt.Println("Filename: " + filename);
	fmt.Println("Size: " + size);
	fmt.Println("TTH: " + tth);
	fmt.Println("Magnet: " + magnet);

	apiUrl := announceUrl;
	resource := "/announce";

	form := url.Values{}
	form.Add("site", site)
	form.Add("filename", filename)
	form.Add("size", size)
	form.Add("tth", tth)

	u, _ := url.ParseRequestURI(apiUrl);
	u.Path = resource;
	urlStr := fmt.Sprintf("%v", u);

	client := &http.Client{};

	r, _ := http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()));
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client.Do(r);
}

func getLargestFile(path string) os.FileInfo {
	files, _ := ioutil.ReadDir(path);

	var largestFile = files[0];

	for _, f := range files {
		if f.Size() > largestFile.Size() {
			largestFile = f;
		}
	}

	fmt.Println("Largest File Found: " + largestFile.Name());

	return largestFile;
}

func calculateTth(tthPath string, filePath string) string {

	out, err := exec.Command(tthPath, filePath).Output()

	if err != nil {
		fmt.Println(err);
		fmt.Println("Error calculating TTH");
	}

	var outputString string = (string(out[:]));

	return outputString[0:39];
}
