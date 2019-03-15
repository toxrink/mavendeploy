package main

import (
	"archive/zip"
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Jar struct {
	groupId    string
	artifactId string
	version    string
	packaging  string
	file       string
	name       string
}

var maven_url string
var maven_id string
var maven_user string = "admin"
var maven_pwd string = "admin123"
var silence bool

func init() {
	flag.StringVar(&maven_url, "url", "http://192.168.119.209:8081/repository/3rd/", "maven repository")
	flag.StringVar(&maven_id, "id", "nexus", "repository id")
	flag.BoolVar(&silence, "silence", false, "automic generate deploy info")
}

func main() {
	flag.Parse()
	var jars = getDeployJars()
	maven_url = confirm("url", maven_url, false)
	maven_id = confirm("id", maven_id, false)
	if '/' != maven_url[len(maven_url)-1] {
		maven_url += "/"
	}

	for _, jar := range jars {
		fmt.Printf("********** %s **********\n", jar.file)
		//deploy(jar)
		httpDeploy(jar)
	}

	for _, jar := range jars {
		fmt.Printf(`<!-- %s -->
<dependency>
    <groupId>%s</groupId>
    <artifactId>%s</artifactId>
    <version>%s</version>
</dependency>
`, maven_url, jar.groupId, jar.artifactId, jar.version)
	}

	fmt.Printf("press any key to finish ...")
	bufio.NewReader(os.Stdin).ReadString(byte('\n'))
}

func deploy(jar *Jar) {
	jar.groupId = confirm("groupId", jar.groupId, silence)
	jar.artifactId = confirm("artifactId", jar.artifactId, silence)
	jar.version = confirm("version", jar.version, silence)

	var params []string
	params = append(params, "deploy:deploy-file")
	params = append(params, "-DgroupId="+jar.groupId)
	params = append(params, "-DartifactId="+jar.artifactId)
	params = append(params, "-Dversion="+jar.version)
	params = append(params, "-Dpackaging=jar")
	params = append(params, "-Dfile="+jar.file)
	params = append(params, "-Durl="+maven_url)
	params = append(params, "-DrepositoryId="+maven_id)

	cmd := exec.Command("mvn", params...)
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func confirm(k string, v string, c bool) string {
	var ret = v
	if c {
		return v
	}
s1:
	fmt.Printf("%s is: %v ? [y/n]\t", k, ret)
	in := bufio.NewReader(os.Stdin)
	cc, _ := in.ReadString(byte('\n'))
	cc = strings.TrimSpace(cc)
	if "" != cc {
		switch cc[0] {
		case 'y', 'Y':
			return ret
		case 'n', 'N':
			fmt.Printf("please input new value:\t")
			ret, _ = in.ReadString(byte('\n'))
			ret = strings.TrimSpace(ret)
			goto s1
		default:
			ret = v
			goto s1
		}
	} else {
		fmt.Printf("%s can not be empty!!!\n", k)
		ret = v
		goto s1
	}
}

func getDeployJars() []*Jar {
	var jars []*Jar
	dir := filepath.Dir(os.Args[0])
	fi, _ := ioutil.ReadDir(dir)
	for _, f := range fi {
		if strings.HasSuffix(f.Name(), ".jar") {
			jars = append(jars, getJar(f.Name(), filepath.Join(dir, f.Name())))
		}
	}
	return jars
}

func getJar(name string, path string) *Jar {
	var jar Jar
	jar.name = name
	jar.file = path
	i := strings.LastIndex(name, "-")
	if i != -1 {
		jar.artifactId = name[:i]
		jar.version = name[i+1 : len(name)-4]
	} else {
		jar.artifactId = name[:len(name)-4]
		jar.version = "1.0"
	}
	jar.groupId = getGroupId(path)
	return &jar
}

func getGroupId(path string) string {
	r, _ := zip.OpenReader(path)
	defer r.Close()
	tmp := map[string]int{}
	for _, f := range r.File {
		t := f.FileHeader.Name
		if strings.HasPrefix(t, "META-INF/maven/") && t != "META-INF/maven/" && strings.HasSuffix(t, "/") {
			return t[15 : len(t)-1]
		}
		if strings.HasSuffix(t, ".class") {
			key := t[:strings.LastIndex(t, "/")]
			key = strings.Replace(key, "/", ".", -1)
			tmp[key] = 0
		}
	}

	short := ""
	for k, _ := range tmp {
		if "" == short {
			short = k
			continue
		}
		if len(k) < len(short) {
			short = k
		}
	}

	mark := len(tmp)
	for mark > 0 {
		for k, _ := range tmp {
			if strings.HasPrefix(k, short) {
				mark--
			} else {
				mark = len(tmp)
				i := strings.LastIndex(short, ".")
				if i == -1 {
					return short
				}
				short = short[:i]
			}
		}
	}

	if strings.HasPrefix(short, "com.vrv.vap") {
		return "com.vrv.vap"
	}

	return short
}
