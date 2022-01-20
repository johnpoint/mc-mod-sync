package main

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/guonaihong/gout"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	rootCmd = &cobra.Command{}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("# All done! Have fun!")
}

func init() {
	rootCmd.AddCommand(gen, get) // API服务,grpc
	gen.Flags().StringVar(&outputPath, "output", "", "打包生成路径")
	get.Flags().StringVar(&configUrl, "url", "", "配置文件所在的url")
}

type modInfo struct {
	Name string `json:"name"`
	Md5  string `json:"md5"`
}

var (
	outputPath = ""
	configUrl  = ""
	gen        = &cobra.Command{
		Use:   "gen",
		Short: "生成当前文件夹下的mod打包文件",
		Long:  "生成当前文件夹下的mod打包文件",
		Run: func(cmd *cobra.Command, args []string) {
			random := RandomString()
			dir, err := ioutil.ReadDir(".")
			if err != nil {
				fmt.Println("! Error", err.Error())
				return
			}
			var modList []*modInfo
			err = os.Mkdir(fmt.Sprintf("tmp-%s", random), fs.ModePerm)
			if err != nil {
				fmt.Println("! Error", err.Error())
				return
			}
			fmt.Println("! Scanning")
			for _, v := range dir {
				if strings.Contains(v.Name(), ".disable") || !strings.Contains(v.Name(), ".jar") {
					continue
				}
				fileByte, err := ioutil.ReadFile(v.Name())
				if err != nil {
					fmt.Println("! Error", err.Error())
					continue
				}
				buffer := bytes.NewBuffer(fileByte)
				hash := md5.New()
				_, _ = io.Copy(hash, buffer)
				fileMD5 := hex.EncodeToString(hash.Sum(nil))
				fmt.Println(fileMD5[:7], v.Name())
				modList = append(modList, &modInfo{
					Name: v.Name(),
					Md5:  fileMD5,
				})
				f, err := os.Create(fmt.Sprintf("tmp-%s/%s", random, fileMD5))
				if err != nil {
					fmt.Println("! Error", err.Error())
					continue
				}
				_, err = f.Write(fileByte)
				if err != nil {
					fmt.Println("! Error", err.Error())
					continue
				}
				err = f.Close()
				if err != nil {
					fmt.Println("! Error", err.Error())
					continue
				}
			}
			jsonByte, err := jsoniter.Marshal(modList)
			if err != nil {
				fmt.Println("! Error", err.Error())
				return
			}
			f, err := os.Create(fmt.Sprintf("tmp-%s/update.json", random))
			if err != nil {
				fmt.Println("! Error", err.Error())
				return
			}
			defer f.Close()
			f.Write(jsonByte)
			if len(outputPath) == 0 {
				outputPath = "."
			}
			err = Zip(fmt.Sprintf("tmp-%s/", random), fmt.Sprintf("%s/mod.zip", outputPath))
			if err != nil {
				fmt.Println("! Error", err.Error())
				return
			}
			err = os.RemoveAll(fmt.Sprintf("tmp-%s", random))
			if err != nil {
				fmt.Println("! Error", err.Error())
				return
			}
		},
	}
	get = &cobra.Command{
		Use:   "get",
		Short: "获取指定目标的mod打包文件",
		Long:  "获取指定目标的mod打包文件",
		Run: func(cmd *cobra.Command, args []string) {
			var haveFileMd5 = make(map[string]string)
			var needDownloadMd5 = make(map[string]string)
			var needDisableMd5 = make(map[string]string)
			var needEnableMd5 = make(map[string]string)
			var modMap = make(map[string]string)
			baseUrl := strings.Replace(configUrl, "update.json", "", -1)
			var jsonByte []byte
			err := gout.GET(configUrl).BindBody(&jsonByte).Do()
			if err != nil {
				fmt.Println("! Error", err.Error())
				return
			}
			dir, err := ioutil.ReadDir(".")
			if err != nil {
				fmt.Println("! Error", err.Error())
				return
			}
			fmt.Println("! Scanning")
			for _, v := range dir {
				s, _ := os.Stat(v.Name())
				if s.IsDir() {
					continue
				}
				if !strings.Contains(v.Name(), ".jar") {
					continue
				}
				fileByte, err := ioutil.ReadFile(v.Name())
				if err != nil {
					fmt.Println("! Error", err.Error())
					continue
				}
				buffer := bytes.NewBuffer(fileByte)
				hash := md5.New()
				_, _ = io.Copy(hash, buffer)
				fileMD5 := hex.EncodeToString(hash.Sum(nil))
				haveFileMd5[fileMD5] = v.Name()
			}
			var modList []*modInfo
			err = jsoniter.Unmarshal(jsonByte, &modList)
			if err != nil {
				fmt.Println("! Error", err.Error())
				return
			}
			for _, v := range modList {
				modMap[v.Md5] = v.Name
				name, has := haveFileMd5[v.Md5]
				if !has {
					needDownloadMd5[v.Md5] = v.Name
					continue
				}
				if strings.Contains(name, ".disable") {
					needEnableMd5[v.Md5] = name
				}
			}
			for k, v := range haveFileMd5 {
				if _, has := modMap[k]; !has {
					if strings.Contains(v, ".disable") {
						continue
					}
					needDisableMd5[k] = v
				}
			}
			fmt.Println("! Disable")
			for k, v := range needDisableMd5 {
				fmt.Println(
					"\033[0;36mx",
					k[0:7],
					strings.Replace(v, ".disable", "", -1)+"\033[0m",
				)
				err := os.Rename(v, fmt.Sprintf("%s.disable", strings.Replace(v, ".disable", "", -1)))
				if err != nil {
					fmt.Println("! Error", err.Error())
					continue
				}
			}
			fmt.Println("! Enable")
			for k, v := range needEnableMd5 {
				fmt.Println(
					"\033[0;32m+",
					k[0:7],
					v+"\033[0m",
				)
				err := os.Rename(v, strings.Replace(v, ".disable", "", -1))
				if err != nil {
					fmt.Println("! Error", err.Error())
					continue
				}
			}
			fmt.Println("! Download")
			for k, v := range needDownloadMd5 {
				fmt.Println("\033[0;33m↓", k[0:7], v+"\033[0m")
				var fileByte []byte
				err := gout.GET(fmt.Sprintf("%s%s", baseUrl, k)).BindBody(&fileByte).Do()
				if err != nil {
					fmt.Println("! Error", err.Error())
					continue
				}
				f, err := os.Create(v)
				if err != nil {
					fmt.Println("! Error", err.Error())
					continue
				}
				_, err = f.Write(fileByte)
				if err != nil {
					fmt.Println("! Error", err.Error())
					continue
				}
				err = f.Close()
				if err != nil {
					fmt.Println("! Error", err.Error())
					continue
				}
			}
		},
	}
)

func Zip(srcFile string, destZip string) error {
	fmt.Println("! Packing")
	zipfile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, filepath.Dir(srcFile)+"/")
		// header.Name = path
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}
		return err
	})

	return err
}

func Unzip(zipFile string, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}

			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func RandomString() string {
	newUUID, _ := uuid.NewRandom()
	return strings.Replace(newUUID.String(), "-", "", -1)
}
