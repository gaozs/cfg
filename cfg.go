// This is ini file parse:
//
// use cfg.ParseIniFile(FilePathName) to parse a file,
// use cfg.GetAttr("group name", "item name") to get an item,
//
// a "main" group with "version" item must be existed in INI file, caller should check this is not "" to ensure INI file is successfully parsed.
package cfg

import (
	"bufio"
	"io"
	"log"
	"os"
	"regexp"
)

// groupAttrs[Attr Name]=Attr Value
type groupAttrs map[string]string

// configs[group name]=groupAttrs
type configs map[string]groupAttrs

var cfgs configs

//parse a ini file
func ParseIniFile(f string) {
	iniFile, err := os.Open(f)
	if err != nil {
		log.Printf("Open INI file(%s) error:%s\n", f, err)
		cfgs = nil
		return
	}
	defer iniFile.Close()
	ParseIni(iniFile)

}

// GetAttr(group name, attr name)
func GetAttr(gname string, aname string) string {
	if cfgs == nil {
		return ""
	}
	gattrs, ok := cfgs[gname]
	if !ok {
		return ""
	}

	attr, ok := gattrs[aname]
	if !ok {
		return ""
	}
	return attr
}

//parse ini from io stream
func ParseIni(source io.Reader) {
	var validStrs = [...]string{
		`^[ \t]*\r?\n?$`,                                               // blank line
		`^[ \t]*#(.*)\r?\n?$`,                                          // comment
		`^[ \t]*\[[ \t]*([a-zA-Z][a-zA-Z0-9_]*)[ \t]*\][ \t]*\r?\n?$`,  // group
		`^[ \t]*([a-zA-Z][a-zA-Z0-9_]*)[ \t]*=[ \t]*(.*)[ \t]*\r?\n?$`, // key value
	}
	var validRes [len(validStrs)]*regexp.Regexp

	for idx := range validRes {
		validRes[idx] = regexp.MustCompile(validStrs[idx])
	}

	var line int = 0
	var curGroup string
	var res []string

	cfgs = make(configs)

	// ensure or generate bufio reader
	reader, ok := source.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(source)
	}

	for {
		input, err := reader.ReadString('\n') // Read line
		if err != nil && err != io.EOF {
			log.Println("Read line err:", err)
			cfgs = nil
			return
		}

		line++
		for idx := range validRes {
			res = validRes[idx].FindStringSubmatch(input)
			if res != nil {
				switch idx {
				case 0:
					//blank line, just ignore
				case 1:
					//comment line, just ignore
				case 2:
					curGroup = res[1]
					if _, ok := cfgs[curGroup]; !ok {
						cfgs[curGroup] = make(groupAttrs)
					}
				case 3:
					if _, ok := cfgs[curGroup]; !ok {
						cfgs[curGroup] = make(groupAttrs)
					}
					cfgs[curGroup][res[1]] = res[2]
				default:
					panic("should not reach here")
				}
				break //matched one reg, break for to process next line
			}
		}

		if res == nil {
			// Error, the input line doesn't match INI format
			log.Printf("INI Line(%d) illegal:%s", line, input)
			cfgs = nil
			return
		}

		if err == io.EOF {
			return
		}
	}
}
