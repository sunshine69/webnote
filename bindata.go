// Package main Code generated by go-bindata. (@generated) DO NOT EDIT.
// sources:
// assets/media/ckeditor/CHANGES.md
// assets/media/ckeditor/LICENSE.md
// assets/media/ckeditor/README.md
// assets/media/ckeditor/build-config.js
// assets/media/ckeditor/ckeditor.js
// assets/media/ckeditor/config.js
// assets/media/ckeditor/contents.css
// assets/media/ckeditor/styles.js
// assets/media/js/ajax.js
// assets/media/js/inlineeditor-source.js
// assets/media/js/inlineeditor.js
// assets/templates/cred_search_results.html
// assets/templates/footer.html
// assets/templates/frontpage.html
// assets/templates/head_menu.html
// assets/templates/header.html
// assets/templates/list_attachment.html
// assets/templates/list_note_attachment.html
// assets/templates/login.html
// assets/templates/noteview1.html
// assets/templates/noteview2.html
// assets/templates/noteview3.html
// assets/templates/search_result.html
// assets/templates/searchuser.html
// assets/templates/upload.html
// assets/templates/userform.html
// assets/templates/userpref_form.html
// assets/templates/userpref_list_form.html
package main

import (
	"bytes"
	"net/http"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// bindataRead reads the given file from disk. It returns an error on failure.
func bindataRead(path, name string) ([]byte, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset %s at %s: %v", name, path, err)
	}
	return buf, err
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// ModTime return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}


type assetFile struct {
	*bytes.Reader
	name            string
	childInfos      []os.FileInfo
	childInfoOffset int
}

type assetOperator struct{}

// Open implement http.FileSystem interface
func (f *assetOperator) Open(name string) (http.File, error) {
	var err error
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	content, err := Asset(name)
	if err == nil {
		return &assetFile{name: name, Reader: bytes.NewReader(content)}, nil
	}
	children, err := AssetDir(name)
	if err == nil {
		childInfos := make([]os.FileInfo, 0, len(children))
		for _, child := range children {
			childPath := filepath.Join(name, child)
			info, errInfo := AssetInfo(filepath.Join(name, child))
			if errInfo == nil {
				childInfos = append(childInfos, info)
			} else {
				childInfos = append(childInfos, newDirFileInfo(childPath))
			}
		}
		return &assetFile{name: name, childInfos: childInfos}, nil
	} else {
		// If the error is not found, return an error that will
		// result in a 404 error. Otherwise the server returns
		// a 500 error for files not found.
		if strings.Contains(err.Error(), "not found") {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
}

// Close no need do anything
func (f *assetFile) Close() error {
	return nil
}

// Readdir read dir's children file info
func (f *assetFile) Readdir(count int) ([]os.FileInfo, error) {
	if len(f.childInfos) == 0 {
		return nil, os.ErrNotExist
	}
	if count <= 0 {
		return f.childInfos, nil
	}
	if f.childInfoOffset+count > len(f.childInfos) {
		count = len(f.childInfos) - f.childInfoOffset
	}
	offset := f.childInfoOffset
	f.childInfoOffset += count
	return f.childInfos[offset : offset+count], nil
}

// Stat read file info from asset item
func (f *assetFile) Stat() (os.FileInfo, error) {
	if len(f.childInfos) != 0 {
		return newDirFileInfo(f.name), nil
	}
	return AssetInfo(f.name)
}

// newDirFileInfo return default dir file info
func newDirFileInfo(name string) os.FileInfo {
	return &bindataFileInfo{
		name:    name,
		size:    0,
		mode:    os.FileMode(2147484068), // equal os.FileMode(0644)|os.ModeDir
		modTime: time.Time{}}
}

// AssetFile return a http.FileSystem instance that data backend by asset
func AssetFile() http.FileSystem {
	return &assetOperator{}
}

// assetsMediaCkeditorChangesMd reads file data from disk. It returns an error on failure.
func assetsMediaCkeditorChangesMd() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/media/ckeditor/CHANGES.md"
	name := "assets/media/ckeditor/CHANGES.md"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsMediaCkeditorLicenseMd reads file data from disk. It returns an error on failure.
func assetsMediaCkeditorLicenseMd() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/media/ckeditor/LICENSE.md"
	name := "assets/media/ckeditor/LICENSE.md"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsMediaCkeditorReadmeMd reads file data from disk. It returns an error on failure.
func assetsMediaCkeditorReadmeMd() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/media/ckeditor/README.md"
	name := "assets/media/ckeditor/README.md"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsMediaCkeditorBuildConfigJs reads file data from disk. It returns an error on failure.
func assetsMediaCkeditorBuildConfigJs() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/media/ckeditor/build-config.js"
	name := "assets/media/ckeditor/build-config.js"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsMediaCkeditorCkeditorJs reads file data from disk. It returns an error on failure.
func assetsMediaCkeditorCkeditorJs() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/media/ckeditor/ckeditor.js"
	name := "assets/media/ckeditor/ckeditor.js"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsMediaCkeditorConfigJs reads file data from disk. It returns an error on failure.
func assetsMediaCkeditorConfigJs() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/media/ckeditor/config.js"
	name := "assets/media/ckeditor/config.js"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsMediaCkeditorContentsCss reads file data from disk. It returns an error on failure.
func assetsMediaCkeditorContentsCss() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/media/ckeditor/contents.css"
	name := "assets/media/ckeditor/contents.css"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsMediaCkeditorStylesJs reads file data from disk. It returns an error on failure.
func assetsMediaCkeditorStylesJs() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/media/ckeditor/styles.js"
	name := "assets/media/ckeditor/styles.js"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsMediaJsAjaxJs reads file data from disk. It returns an error on failure.
func assetsMediaJsAjaxJs() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/media/js/ajax.js"
	name := "assets/media/js/ajax.js"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsMediaJsInlineeditorSourceJs reads file data from disk. It returns an error on failure.
func assetsMediaJsInlineeditorSourceJs() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/media/js/inlineeditor-source.js"
	name := "assets/media/js/inlineeditor-source.js"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsMediaJsInlineeditorJs reads file data from disk. It returns an error on failure.
func assetsMediaJsInlineeditorJs() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/media/js/inlineeditor.js"
	name := "assets/media/js/inlineeditor.js"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesCred_search_resultsHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesCred_search_resultsHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/cred_search_results.html"
	name := "assets/templates/cred_search_results.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesFooterHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesFooterHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/footer.html"
	name := "assets/templates/footer.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesFrontpageHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesFrontpageHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/frontpage.html"
	name := "assets/templates/frontpage.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesHead_menuHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesHead_menuHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/head_menu.html"
	name := "assets/templates/head_menu.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesHeaderHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesHeaderHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/header.html"
	name := "assets/templates/header.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesList_attachmentHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesList_attachmentHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/list_attachment.html"
	name := "assets/templates/list_attachment.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesList_note_attachmentHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesList_note_attachmentHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/list_note_attachment.html"
	name := "assets/templates/list_note_attachment.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesLoginHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesLoginHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/login.html"
	name := "assets/templates/login.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesNoteview1Html reads file data from disk. It returns an error on failure.
func assetsTemplatesNoteview1Html() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/noteview1.html"
	name := "assets/templates/noteview1.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesNoteview2Html reads file data from disk. It returns an error on failure.
func assetsTemplatesNoteview2Html() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/noteview2.html"
	name := "assets/templates/noteview2.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesNoteview3Html reads file data from disk. It returns an error on failure.
func assetsTemplatesNoteview3Html() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/noteview3.html"
	name := "assets/templates/noteview3.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesSearch_resultHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesSearch_resultHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/search_result.html"
	name := "assets/templates/search_result.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesSearchuserHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesSearchuserHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/searchuser.html"
	name := "assets/templates/searchuser.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesUploadHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesUploadHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/upload.html"
	name := "assets/templates/upload.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesUserformHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesUserformHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/userform.html"
	name := "assets/templates/userform.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesUserpref_formHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesUserpref_formHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/userpref_form.html"
	name := "assets/templates/userpref_form.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// assetsTemplatesUserpref_list_formHtml reads file data from disk. It returns an error on failure.
func assetsTemplatesUserpref_list_formHtml() (*asset, error) {
	path := "/home/stevek/src/webnote-go/assets/templates/userpref_list_form.html"
	name := "assets/templates/userpref_list_form.html"
	bytes, err := bindataRead(path, name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(path)
	if err != nil {
		err = fmt.Errorf("Error reading asset info %s at %s: %v", name, path, err)
	}

	a := &asset{bytes: bytes, info: fi}
	return a, err
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"assets/media/ckeditor/CHANGES.md":           assetsMediaCkeditorChangesMd,
	"assets/media/ckeditor/LICENSE.md":           assetsMediaCkeditorLicenseMd,
	"assets/media/ckeditor/README.md":            assetsMediaCkeditorReadmeMd,
	"assets/media/ckeditor/build-config.js":      assetsMediaCkeditorBuildConfigJs,
	"assets/media/ckeditor/ckeditor.js":          assetsMediaCkeditorCkeditorJs,
	"assets/media/ckeditor/config.js":            assetsMediaCkeditorConfigJs,
	"assets/media/ckeditor/contents.css":         assetsMediaCkeditorContentsCss,
	"assets/media/ckeditor/styles.js":            assetsMediaCkeditorStylesJs,
	"assets/media/js/ajax.js":                    assetsMediaJsAjaxJs,
	"assets/media/js/inlineeditor-source.js":     assetsMediaJsInlineeditorSourceJs,
	"assets/media/js/inlineeditor.js":            assetsMediaJsInlineeditorJs,
	"assets/templates/cred_search_results.html":  assetsTemplatesCred_search_resultsHtml,
	"assets/templates/footer.html":               assetsTemplatesFooterHtml,
	"assets/templates/frontpage.html":            assetsTemplatesFrontpageHtml,
	"assets/templates/head_menu.html":            assetsTemplatesHead_menuHtml,
	"assets/templates/header.html":               assetsTemplatesHeaderHtml,
	"assets/templates/list_attachment.html":      assetsTemplatesList_attachmentHtml,
	"assets/templates/list_note_attachment.html": assetsTemplatesList_note_attachmentHtml,
	"assets/templates/login.html":                assetsTemplatesLoginHtml,
	"assets/templates/noteview1.html":            assetsTemplatesNoteview1Html,
	"assets/templates/noteview2.html":            assetsTemplatesNoteview2Html,
	"assets/templates/noteview3.html":            assetsTemplatesNoteview3Html,
	"assets/templates/search_result.html":        assetsTemplatesSearch_resultHtml,
	"assets/templates/searchuser.html":           assetsTemplatesSearchuserHtml,
	"assets/templates/upload.html":               assetsTemplatesUploadHtml,
	"assets/templates/userform.html":             assetsTemplatesUserformHtml,
	"assets/templates/userpref_form.html":        assetsTemplatesUserpref_formHtml,
	"assets/templates/userpref_list_form.html":   assetsTemplatesUserpref_list_formHtml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"assets": &bintree{nil, map[string]*bintree{
		"media": &bintree{nil, map[string]*bintree{
			"ckeditor": &bintree{nil, map[string]*bintree{
				"CHANGES.md":      &bintree{assetsMediaCkeditorChangesMd, map[string]*bintree{}},
				"LICENSE.md":      &bintree{assetsMediaCkeditorLicenseMd, map[string]*bintree{}},
				"README.md":       &bintree{assetsMediaCkeditorReadmeMd, map[string]*bintree{}},
				"build-config.js": &bintree{assetsMediaCkeditorBuildConfigJs, map[string]*bintree{}},
				"ckeditor.js":     &bintree{assetsMediaCkeditorCkeditorJs, map[string]*bintree{}},
				"config.js":       &bintree{assetsMediaCkeditorConfigJs, map[string]*bintree{}},
				"contents.css":    &bintree{assetsMediaCkeditorContentsCss, map[string]*bintree{}},
				"styles.js":       &bintree{assetsMediaCkeditorStylesJs, map[string]*bintree{}},
			}},
			"js": &bintree{nil, map[string]*bintree{
				"ajax.js":                &bintree{assetsMediaJsAjaxJs, map[string]*bintree{}},
				"inlineeditor-source.js": &bintree{assetsMediaJsInlineeditorSourceJs, map[string]*bintree{}},
				"inlineeditor.js":        &bintree{assetsMediaJsInlineeditorJs, map[string]*bintree{}},
			}},
		}},
		"templates": &bintree{nil, map[string]*bintree{
			"cred_search_results.html":  &bintree{assetsTemplatesCred_search_resultsHtml, map[string]*bintree{}},
			"footer.html":               &bintree{assetsTemplatesFooterHtml, map[string]*bintree{}},
			"frontpage.html":            &bintree{assetsTemplatesFrontpageHtml, map[string]*bintree{}},
			"head_menu.html":            &bintree{assetsTemplatesHead_menuHtml, map[string]*bintree{}},
			"header.html":               &bintree{assetsTemplatesHeaderHtml, map[string]*bintree{}},
			"list_attachment.html":      &bintree{assetsTemplatesList_attachmentHtml, map[string]*bintree{}},
			"list_note_attachment.html": &bintree{assetsTemplatesList_note_attachmentHtml, map[string]*bintree{}},
			"login.html":                &bintree{assetsTemplatesLoginHtml, map[string]*bintree{}},
			"noteview1.html":            &bintree{assetsTemplatesNoteview1Html, map[string]*bintree{}},
			"noteview2.html":            &bintree{assetsTemplatesNoteview2Html, map[string]*bintree{}},
			"noteview3.html":            &bintree{assetsTemplatesNoteview3Html, map[string]*bintree{}},
			"search_result.html":        &bintree{assetsTemplatesSearch_resultHtml, map[string]*bintree{}},
			"searchuser.html":           &bintree{assetsTemplatesSearchuserHtml, map[string]*bintree{}},
			"upload.html":               &bintree{assetsTemplatesUploadHtml, map[string]*bintree{}},
			"userform.html":             &bintree{assetsTemplatesUserformHtml, map[string]*bintree{}},
			"userpref_form.html":        &bintree{assetsTemplatesUserpref_formHtml, map[string]*bintree{}},
			"userpref_list_form.html":   &bintree{assetsTemplatesUserpref_list_formHtml, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
