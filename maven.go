package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func httpDeploy(jar *Jar) {
	deployJarFile(jar)
	deployPom(jar)
}

func deployJarFile(jar *Jar) {
	//------------jar--------------
	body, _ := os.Open(jar.file)
	defer body.Close()
	url := strings.Join([]string{maven_url + strings.Replace(jar.groupId, ".", "/", -1),
		jar.artifactId,
		jar.version,
		jar.name}, "/")
	put(url, body)

	//------------md5--------------
	bb, _ := ioutil.ReadAll(body)
	url = strings.Join([]string{maven_url + strings.Replace(jar.groupId, ".", "/", -1),
		jar.artifactId,
		jar.version,
		jar.name + ".md5"}, "/")
	putString(url, md5string(bb))

	//------------sha1--------------
	url = strings.Join([]string{maven_url + strings.Replace(jar.groupId, ".", "/", -1),
		jar.artifactId,
		jar.version,
		jar.name + ".sha1"}, "/")
	putString(url, sha1string(bb))
}

func deployPom(jar *Jar) {
	//------------pom--------------
	body := `<?xml version="1.0" encoding="UTF-8"?>
<project xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd" xmlns="http://maven.apache.org/POM/4.0.0"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <modelVersion>4.0.0</modelVersion>
  <groupId>` + jar.groupId + `</groupId>
  <artifactId>` + jar.artifactId + `</artifactId>
  <version>` + jar.version + `</version>
</project>
`
	url := strings.Join([]string{maven_url + strings.Replace(jar.groupId, ".", "/", -1),
		jar.artifactId,
		jar.version,
		jar.artifactId + "-" + jar.version + ".pom"}, "/")
	putString(url, body)
	
	//------------md5--------------
	bb := bytes.NewBufferString(body).Bytes()
	url = strings.Join([]string{maven_url + strings.Replace(jar.groupId, ".", "/", -1),
		jar.artifactId,
		jar.version,
		jar.artifactId + "-" + jar.version + ".pom.md5"}, "/")
	putString(url, md5string(bb))
	
	//------------sha1--------------
	url = strings.Join([]string{maven_url + strings.Replace(jar.groupId, ".", "/", -1),
		jar.artifactId,
		jar.version,
		jar.artifactId + "-" + jar.version + ".pom.sha1"}, "/")
	putString(url, sha1string(bb))
}

func putString(url, body string) {
	put(url, bytes.NewBufferString(body))
}

func put(url string, body io.Reader) {
	fmt.Println(url)

	var client http.Client
	req, err := http.NewRequest("PUT", url, body)
	if nil != err {
		fmt.Println(err)
	}
	req.SetBasicAuth(maven_user, maven_pwd)
	client.Do(req)
}

func md5string(b []byte) string {
	md5n := md5.New()
	md5n.Write(b)
	return hex.EncodeToString(md5n.Sum(nil))
}

func sha1string(b []byte) string {
	sha1n := sha1.New()
	sha1n.Write(b)
	return hex.EncodeToString(sha1n.Sum(nil))
}
