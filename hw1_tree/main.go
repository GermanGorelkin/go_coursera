package main

import (
	"fmt"
	_ "fmt"
	"io"
	_ "log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, startPath string, printFiles bool) (err error) {
	rootFolder := newFolder(startPath)

	err = filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
		segments := strings.Split(path, string(filepath.Separator))
		if info.IsDir() {
			getFolder(rootFolder, segments)

		} else {
			f := getFolder(rootFolder, segments[:len(segments)-1])
			f.Files = append(f.Files, &FSNode{Name: info.Name(), Size: info.Size()})
		}

		return err
	})

	printTree(out, rootFolder.Folders[startPath], "", printFiles)

	//for name, folder := range rootFolder.Folders {
	//	fmt.Println(name, folder)
	//}

	//fmt.Println(rootFolder.Folders[startPath].Folders["zline"].Folders["lorem"].Files)

	//printFolders(rootFolder.Folders[startPath], "")

	return err
}

func printTree(out io.Writer, f *FSNode, prefix string, printFiles bool) {
	fl := make([]*FSNode, 0, len(f.Folders)+len(f.Files))
	for _, v := range f.Folders {
		fl = append(fl, v)
	}

	if printFiles {
		for _, v := range f.Files {
			fl = append(fl, v)
		}
	}

	sort.Slice(fl, func(i, j int) bool {
		return fl[i].Name < fl[j].Name
	})

	var subPrefix string

	for i, node := range fl {

		if i != len(fl)-1 {
			subPrefix = fmt.Sprintf("%v├───", prefix)
		} else {
			subPrefix = fmt.Sprintf("%v└───", prefix)
		}
		fmt.Fprintf(out, "%v%v\n", subPrefix, node)

		if i != len(fl)-1 {
			subPrefix = strings.Replace(subPrefix, "├───", "│\t", -1)
		} else {
			subPrefix = strings.Replace(subPrefix, "└───", "\t", -1)
		}

		if node.IsDir {
			printTree(out, f.Folders[node.Name], subPrefix, printFiles)
		}

		//subPrefix := fmt.Sprintf("%v\t├───", subPrefix)
		//printTree(f.Folders[name], subPrefix)

		//subPrefix := strings.Replace(prefix, "├───", "│", -1)
		//subPrefix = fmt.Sprintf("%v\t├───", subPrefix)
		//
		//if i == len(fl)-1 {
		//	subPrefix = strings.Replace(subPrefix, "├───", "└───", -1)
		//	subPrefix = strings.Replace(subPrefix, "│", "", -1)
		//}

		//printTree(f.Folders[name], subPrefix)

		//if i == len(fl)-1 {
		//	prefix =strings.Replace(prefix, "├───", "└───", -1)
		//}
		//fmt.Printf("%v%v\n", prefix, name)
		//
		//subPrefix := prefix
		////subPrefix :=strings.Replace(prefix, "├───", "│", 1)
		//if i != len(fl)-1 {
		//	subPrefix =strings.Replace(subPrefix, "├───", "│", -1)
		//	subPrefix += "\t├───"
		//} else {
		//	//subPrefix =strings.Replace(subPrefix, "├───", "\t", -1)
		//	subPrefix += "\t└───"
		//}
		//fmt.Printf("%v%v\n", subPrefix, name)
		//
		//printTree(f.Folders[name], subPrefix)
	}
}

//type File struct {
//	Name string
//	Size int64
//}

type FSNode struct {
	Name    string
	Size    int64
	IsDir   bool
	Files   []*FSNode
	Folders map[string]*FSNode
}

func (f *FSNode) String() string {
	if f.IsDir {
		return f.Name
	}

	var size string
	if f.Size > 0 {
		size = fmt.Sprintf("%vb", f.Size)
	} else {
		size = "empty"
	}

	return fmt.Sprintf("%v (%v)", f.Name, size)
}

func newFolder(name string) *FSNode {
	return &FSNode{Name: name, Files: []*FSNode{}, Folders: make(map[string]*FSNode), IsDir: true}
}

func getFolder(f *FSNode, path []string) *FSNode {
	var folderSearch *FSNode
	folderRoot := f
	var ok bool
	for _, segment := range path {
		if folderSearch, ok = folderRoot.Folders[segment]; ok {
			folderRoot = folderSearch
		} else {
			folderSearch = newFolder(segment)
			folderRoot.Folders[segment] = folderSearch
			folderRoot = folderSearch
		}
	}

	return folderSearch
}

//func printFolders(f *FSNode, path string) {
//	if len(f.Folders) > 0 {
//		for name, folder := range f.Folders {
//			newPath := path
//			newPath += fmt.Sprintf("%v/", name)
//			printFolders(folder, newPath)
//		}
//	} else {
//		fmt.Printf("%v\n", path)
//	}
//}

//func printFolders(f *Folder) string {
//	var str string
//	if len(f.Folders) > 0 {
//		for name, folder := range f.Folders {
//			str += fmt.Sprintf("%v/", name)
//			str += printFolders(folder)
//		}
//	} else {
//		str = "\n"
//	}
//	return str
//}
