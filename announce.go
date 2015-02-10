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

	c, _ := config.ReadDefault(configFile);

	var site, _ = c.String("announce", "site");
	var announceUrl, _ = c.String("announce", "url");
	var tthPath, _ = c.String("announce", "tthsum");

	args := os.Args[1:]

	if len(args) == 8 {
		fmt.Println("SABnzbd mode");

		var downloadPath string = args[0];
		var processingStatus string = args[6];

		if processingStatus == "0" {
			fmt.Println("Processing completed successfully");

			announce(site, announceUrl, downloadPath, tthPath);
		}
	}
}

func announce(site string, announceUrl string, downloadPath string, tthPath string) {
	var largestFile os.FileInfo = getLargestFile(downloadPath);
	var filename string = largestFile.Name();
	var size string = strconv.FormatInt(largestFile.Size(), 10);
	var tth string = calculateTth(tthPath, downloadPath+"/"+largestFile.Name());
	var magnet string = fmt.Sprintf("magnet:?xt=urn:tree:tiger:%s&xl=%s&dn=%s", tth, size, filename);

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
