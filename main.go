//author :Ally Dale(vipaly@gmail.com)
//date: 2016-08-24

//tool installgithub is used to download the latest version of GitHub desktop offline install files
//refer: https://desktop.github.com
package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vipally/cmdline"
)

var (
	root_url    = "http://github-windows.s3.amazonaws.com" //下载根目录
	break_point = true
	local_root  = "."
	curl        = "./CURL.EXE"
	root_file   = "GitHub.application"
)

func main() {
	cmdline.Summary("Command installgithub is used to download the latest version of GitHub desktop offline install files")
	cmdline.Details("More information refer: https://desktop.github.com")
	cmdline.StringVar(&root_url, "r", "root", root_url, false, "root_url of GitHub desktop")
	cmdline.StringVar(&local_root, "d", "dir", ".", false, "local root dir for download")
	cmdline.BoolVar(&break_point, "b", "break_point", break_point, false, "if download from last break_point")
	file := cmdline.String("f", "file", "", false, "single file path to download")
	//*file = "Application Files\\GitHub_3_2_0_0\\GitHub.exe.manifest"
	cmdline.Parse()
	if *file != "" {
		dn_file(*file, false)
		return
	} else {
		dn_from_root(break_point)
	}
}

type File struct {
	Type string
	Path string
	Size int
}

func get_dn_list(file string) (r []*File, err error) {
	full_path := local_dir(file)
	if f, err2 := os.Open(full_path); err == nil {
		d := xml.NewDecoder(f)
		for t, err3 := d.Token(); err3 == nil; t, err3 = d.Token() {
			switch token := t.(type) {
			case xml.StartElement:
				name := token.Name.Local
				if name == "dependentAssembly" {
					var nf File
					for _, attr := range token.Attr {
						switch attr.Name.Local {
						case "dependencyType":
							nf.Type = attr.Value
						case "codebase":
							nf.Path = attr.Value
						case "size":
							nf.Size, _ = strconv.Atoi(attr.Value)
						}
					}
					if nf.Type == "install" {
						r = append(r, &nf)
					}
				}
			case xml.EndElement:
			case xml.CharData:
			default:
			}
		}
	} else {
		err = err2
	}
	return
}

func dn_from_root(brk bool) error {
	dn_file(root_file, false)
	if l, e := get_dn_list(root_file); e == nil {
		for _, v := range l {
			dn_file(v.Path, false)
			dir := filepath.Dir(v.Path)
			if l2, e2 := get_dn_list(v.Path); e2 == nil {
				n := len(l2)
				for i, v2 := range l2 {
					v2.Path = dir + "\\" + v2.Path + ".deploy"
					fmt.Printf("list%d/%d: %s\n", i+1, n, v2.Path)
				}
				for i, v2 := range l2 {
					fmt.Printf("%d/%d %s\n", i+1, n, v2.Path)
					dn_file(v2.Path, brk)
				}
				fmt.Printf("\n\n\n!!!!!!!!!!!!!!Download finished, click [%s] to stat install!!!!!!!!!!!!!\n", root_file)
			} else {
				return e2
			}
		}
	} else {
		return e
	}
	return nil
}

func check_file(file string) bool {
	if f, e := os.Open(file); e == nil {
		f.Close()
		return true
	}
	return false
}

func dn_file(file string, brk bool) error {
	url := full_url(file)
	local := local_dir(file)
	mk_dir(file)
	fmt.Println("downloding:", url)
	if brk && check_file(local) {
		fmt.Println("exist and skip", url)
	} else {
		cmd := exec.Command(curl, "-o", local, url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		cmd.Wait()
		fmt.Println("finish", url)
	}

	return nil
}

func full_url(file string) string {
	url := root_url + "/" + file
	url = strings.Replace(url, " ", "%20", -1)
	url = strings.Replace(url, "\\", "/", -1)
	return url
}
func local_dir(file string) string {
	return local_root + "/" + file
}

func mk_dir(file string) {
	d := filepath.Dir(file)
	os.MkdirAll(d, os.ModeDir)
}
