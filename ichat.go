package iChatTool

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/JamieSinn/go-plist"
	"github.com/beevik/guid"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

type PListArray struct {
	Array []PListKV
}

type PListKV struct {
	Key   string
	Value interface{}
}

type AttachedImage struct {
	ImageType  string
	ImageBytes []byte
	Img        image.Image
}

func (kv *PListKV) Print(tabs int) {
	prefix := ""
	if tabs == 0 {
		prefix = ""
	} else {
		prefix = strings.Repeat("\t", tabs)
	}

	fmt.Println(prefix + kv.Key)

	switch t := kv.Value.(type) {
	case PListArray:
		pl := PListArray{}
		pl = t
		for i := 0; i < len(pl.Array); i++ {
			pl.Array[i].Print(tabs + 1)
		}
	case PListKV:
		fmt.Println(prefix + t.Key)
		fmt.Print(prefix + "\t")
		fmt.Print(t.Value)
		fmt.Print("\n")
		break
	case string:
		decoded, err := base64.StdEncoding.DecodeString(t)
		if err == nil {
			size := len(decoded)
			fmt.Println(prefix + "\t" + "Base64String - size: " + strconv.Itoa(size))
			if size < 0x1000 {
				break
			}
			attachedimage, err := parseImage(decoded)
			if err != nil {
				break
			}
			fmt.Println(prefix + "\tFound Image of type: " + attachedimage.ImageType)
		}
		fmt.Println(prefix + "\t" + t)
	default:
		fmt.Print(prefix + "\t")
		fmt.Print(t)
		fmt.Print("\n")
	}
}

func (kv *PListKV) ExtractImages(tabs int) AttachedImage {
	prefix := ""
	if tabs == 0 {
		prefix = ""
	} else {
		prefix = strings.Repeat("\t", tabs)
	}
	switch t := kv.Value.(type) {
	case PListArray:
		pl := PListArray{}
		pl = t
		for i := 0; i < len(pl.Array); i++ {
			ret := pl.Array[i].ExtractImages(tabs + 1)
			if ret.ImageType != "" {
				return ret
			}
		}
	case PListKV:
		break
	case string:
		decoded, err := base64.StdEncoding.DecodeString(t)
		if err == nil {
			size := len(decoded)
			fmt.Println(prefix + "\t" + "Base64String - size: " + strconv.Itoa(size))
			if size < 0x1000 {
				break
			}
			attachedimage, err := parseImage(decoded)
			if err != nil {
				break
			}
			fmt.Println(prefix + "\tFound Image of type: " + attachedimage.ImageType)
			return attachedimage
		}
	default:
		break
	}
	return AttachedImage{}
}

func parseImage(data []byte) (AttachedImage, error) {
	// Seems like all the data starts at 0x1000 regardless of filetype
	imgbytes := data[0x1000 : len(data)-0x4E]

	img, ext, err := image.Decode(bytes.NewReader(imgbytes))
	if err == nil {
		return AttachedImage{
			ImageType:  ext,
			ImageBytes: imgbytes,
			Img:        img,
		}, nil
	}
	return AttachedImage{}, err
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func extractPList(i interface{}) PListArray {
	return readPlistInterface("^$", i)
}

func readPlistInterface(key string, i interface{}) PListArray {
	v := reflect.ValueOf(i)
	switch t := v.Interface().(type) {
	case []byte:
		str := base64.StdEncoding.EncodeToString(t)
		return PListArray{[]PListKV{{key, str}}}
	case []interface{}:
		var vals []PListKV
		for i := 0; i < len(t); i++ {
			vals = append(vals, readPlistInterface(key, t[i]).Array...)
		}
		return PListArray{Array: vals}
	default:
		if v.Kind() == reflect.Map {
			var ret PListKV
			ret.Key = key
			var data = PListArray{}
			for _, key := range v.MapKeys() {
				val := v.MapIndex(key)
				data.Array = append(data.Array, readPlistInterface(key.String(), val.Interface()).Array...)
			}
			ret.Value = data
			return PListArray{[]PListKV{ret}}
		}
		return PListArray{[]PListKV{{key, t}}}
	}
}

func dumpPlist(i interface{}, tabs int) {
	var prefix string
	if tabs == 0 {
		prefix = ""
	} else {
		prefix = strings.Repeat("\t", tabs)
	}

	v := reflect.ValueOf(i)
	switch t := v.Interface().(type) {
	case []byte:
		str := base64.StdEncoding.EncodeToString(t)
		fmt.Println(prefix + "BASE64 String - length " + strconv.Itoa(len(t)))
		fmt.Println(prefix + "Starts with: " + str[:len(str)-(len(str)-4)])
		break
	case string:
		g, err := guid.ParseString(t)
		if err == nil {
			fmt.Println("GUID: " + g.String())
			break
		}
		fmt.Println(prefix + t)

		break
	case map[interface{}]interface{}:
		fmt.Println(prefix + "Interface map")
		//DumpPlist(v)
		break
	case []interface{}:
		fmt.Println(prefix + "Interface array - length " + strconv.Itoa(len(t)))
		for i := 0; i < len(t); i++ {
			dumpPlist(t[i], tabs+1)
		}
		break
	case bool:
		fmt.Println(prefix + "Bool: " + strconv.FormatBool(t))
		break
	case int64:
		fmt.Println(prefix + "Int64:" + strconv.Itoa(int(t)))
		break
	case uint64:
		fmt.Println(prefix + "uInt64:" + strconv.Itoa(int(t)))
		break
	case float64:
		fmt.Println(prefix + "float64:" + strconv.Itoa(int(t)))
		break
	default:
		if v.Kind() == reflect.Map {
			for _, key := range v.MapKeys() {
				val := v.MapIndex(key)
				fmt.Println(prefix + "\tkey: " + key.String())
				dumpPlist(val.Interface(), tabs+2)
			}
		} else {
			fmt.Println(prefix + "TYPE: " + v.Kind().String())
			fmt.Println(prefix + strconv.Itoa(int(v.Uint())))
		}
	}
}

func DumpPList(path string) {
	dumpPlist(readPlistFile(path), 0)
}

func readPlistFile(path string) (output map[string]interface{}) {
	data, err := ioutil.ReadFile(path)
	check(err)
	decoder := plist.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&output)
	check(err)
	return
}

func ExtractData(path string) {
	extract := extractPList(readPlistFile(path))
	for i := 0; i < len(extract.Array); i++ {
		extract.Array[i].Print(0)
	}
}

func ExtractImages(path string) (images []AttachedImage) {
	extract := extractPList(readPlistFile(path))
	for i := 0; i < len(extract.Array); i++ {
		images = append(images, extract.Array[i].ExtractImages(0))
	}
	return
}
