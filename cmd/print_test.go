package cmd

import (
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWalkDirectoryAllTags(t *testing.T) {
	err, tags := getAllTags(filepath.Join("..", "test-resources", "simple.dcm"))
	if err != nil {
		t.Error(err)
		return
	}
	result, err := walkDirectory(filepath.Join("..", "test-resources"), tags, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 1 {
		t.Fatalf(`walkDirectory should return 1 dicoms, got %d`, len(result))
	}
	if len(result[0])-1 != len(tags) {
		t.Fatalf(`walkDirectory should return %d tags, got %d`, len(tags), len(result[0])-1)
	}
}

func TestWalkDirectorySingleTag(t *testing.T) {
	singleTag, err := tag.FindByName("TransferSyntaxUID")
	actual, err := walkDirectory(filepath.Join("..", "test-resources"), []tag.Info{singleTag}, nil)
	if err != nil {
		t.Error(err)
		return
	}
	expected := [][]string{{filepath.Join("..", "test-resources", "simple.dcm"), "1.2.840.10008.1.2.1"}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf(`walkDirectory should return %v, got %v`, expected, actual)
	}
}

func TestWalkDirectoryExclusion(t *testing.T) {
	singleTag, err := tag.FindByName("TransferSyntaxUID")
	exclusions = []string{filepath.Join("**", "simple.*")}
	actual, err := walkDirectory(filepath.Join("..", "test-resources"), []tag.Info{singleTag}, nil)
	exclusions = []string{}
	if err != nil {
		t.Error(err)
		return
	}
	expected := make([][]string, 0)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf(`walkDirectory should return %v, got %v`, expected, actual)
	}
}

func TestWalkDirectoryFilterMatches(t *testing.T) {
	singleTag, err := tag.FindByName("TransferSyntaxUID")
	filterTag, _ := tag.FindByName("Modality")
	actual, err := walkDirectory(filepath.Join("..", "test-resources"), []tag.Info{singleTag}, map[tag.Info]string{filterTag: "CT", singleTag: "1.2.840.10008.1.2.1"})
	if err != nil {
		t.Error(err)
		return
	}
	expected := [][]string{{filepath.Join("..", "test-resources", "simple.dcm"), "1.2.840.10008.1.2.1"}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf(`walkDirectory should return %v, got %v`, expected, actual)
	}
}

func TestWalkDirectoryFilterMismatch(t *testing.T) {
	singleTag, err := tag.FindByName("TransferSyntaxUID")
	actual, err := walkDirectory(filepath.Join("..", "test-resources"), []tag.Info{singleTag}, map[tag.Info]string{singleTag: "1.2.840.10008.1.2.2"})
	if err != nil {
		t.Error(err)
		return
	}
	expected := make([][]string, 0)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf(`walkDirectory should return %v, got %v`, expected, actual)
	}
}

func getAllTags(path string) (error, []tag.Info) {
	file, err := dicom.ParseFile(path, nil)
	if err != nil {
		return err, nil
	}
	tags := make([]tag.Info, 0)
	for iter := file.FlatStatefulIterator(); iter.HasNext(); {
		info, _ := tag.Find(iter.Next().Tag)
		tags = append(tags, info)
	}
	return err, tags
}
