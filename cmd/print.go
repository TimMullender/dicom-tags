package cmd

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/cobra"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type stringSort [][]string

func (s stringSort) Len() int {
	return len(s)
}
func (s stringSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s stringSort) Less(i, j int) bool {
	return s[i][1] < s[j][1]
}

type integerSort [][]string

func (s integerSort) Len() int {
	return len(s)
}
func (s integerSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s integerSort) Less(i, j int) bool {
	in, iErr := strconv.ParseInt(s[i][1], 10, 64)
	jn, jErr := strconv.ParseInt(s[j][1], 10, 64)
	if iErr == nil && jErr == nil {
		return in < jn
	}
	return false
}

var (
	Version = "dev"

	exclusions []string
	filters    map[string]string
	limit      uint
	numeric    bool
	offset     uint
	sorted     bool

	rootCmd = &cobra.Command{
		Use:     "dicom-tags [folder tag-list]",
		Version: Version,
		Args:    cobra.MinimumNArgs(2),
		Short:   "Prints selected DICOM tags",
		Long:    `Walks a directory and prints the selected tags for each DICOM that is found`,
		Run: func(cmd *cobra.Command, args []string) {
			printTags(args)
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(127)
	}
}

func init() {
	rootCmd.Flags().StringSliceVarP(&exclusions, "exclusion", "e", nil, "Exclude paths using glob")
	rootCmd.Flags().StringToStringVarP(&filters, "filter", "f", map[string]string{}, "Filter the printed records using tag=value")
	rootCmd.Flags().UintVarP(&limit, "limit", "l", 0, "Limit the number of records printed, 0 indicates no limit")
	rootCmd.Flags().BoolVarP(&numeric, "numeric", "n", false, "Sort by the first tag numerically")
	rootCmd.Flags().UintVarP(&offset, "offset", "o", 0, "Skip printing a number of records")
	rootCmd.Flags().BoolVarP(&sorted, "sort", "s", false, "Sort by the first tag")
}

func printTags(args []string) {
	tags := findAllTags(args[1:])
	if len(tags) < 1 {
		_, _ = fmt.Fprintln(os.Stderr, "No valid tags found")
		os.Exit(1)
	}
	filterValues := make(map[tag.Info]string, len(filters))
	for tagName, value := range filters {
		info, err := tag.FindByName(tagName)
		if err == nil {
			filterValues[info] = value
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "Invalid Filter tag: %s\n %v\n", tagName, err)
			os.Exit(4)
		}
	}
	headers := append([]string{"Filename"}, args[1:]...)
	values, err := walkDirectory(args[0], tags, filterValues)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error processing directory: %s\n %v\n", args[0], err)
		os.Exit(3)
	}
	if sorted {
		sort.Sort(stringSort(values))
	}
	if numeric {
		sort.Sort(integerSort(values))
	}

	recordCount := uint(len(values))
	if limit > 0 && offset+limit < recordCount {
		recordCount = offset + limit
	}
	if offset > recordCount {
		offset = recordCount
	}
	values = values[offset:recordCount]

	csvWriter := csv.NewWriter(os.Stdout)
	_ = csvWriter.Write(headers)
	_ = csvWriter.WriteAll(values)
	csvWriter.Flush()
}

func walkDirectory(directoryPath string, tags []tag.Info, filterValues map[tag.Info]string) ([][]string, error) {
	values := make([][]string, 0)
	err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		for _, exclusion := range exclusions {
			matched, _ := doublestar.PathMatch(exclusion, path)
			if matched {
				return nil
			}
		}
		zipReader, err := zip.OpenReader(path)
		if err == nil {
			defer zipReader.Close()
			for _, file := range zipReader.File {
				reader, err := file.Open()
				if err != nil || file.FileInfo().IsDir() {
					continue
				}
				dataset, err := dicom.Parse(reader, int64(file.UncompressedSize64), nil, dicom.SkipProcessingPixelDataValue())
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Skipping %s#%s due to: %v\n", path, file.Name, err)
					continue
				}
				if matches(dataset, filterValues) {
					values = append(values, getValues(path+"#"+file.Name, tags, dataset))
				}
			}
			return nil
		}
		dataset, err := dicom.ParseFile(path, nil, dicom.SkipProcessingPixelDataValue())
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Skipping %s due to: %v\n", path, err)
			return nil
		}
		if matches(dataset, filterValues) {
			values = append(values, getValues(path, tags, dataset))
		}
		return nil
	})
	return values, err
}

func matches(dataset dicom.Dataset, filterValues map[tag.Info]string) bool {
	if filterValues == nil {
		return true
	}
	for info, value := range filterValues {
		element, err := dataset.FindElementByTag(info.Tag)
		if err == nil {
			if getValue(info, element) != value {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func getValues(p string, tags []tag.Info, dataset dicom.Dataset) []string {
	values := make([]string, 0)
	values = append(values, p)
	for _, tagInfo := range tags {
		value := tagInfo.Name
		element, err := dataset.FindElementByTag(tagInfo.Tag)
		if err == nil {
			value = getValue(tagInfo, element)
		}
		values = append(values, value)
	}
	return values
}

func getValue(tagInfo tag.Info, element *dicom.Element) string {
	if tagInfo.VM == "1" {
		return getFirstValue(element.Value)
	} else {
		return element.Value.String()
	}
}

func getFirstValue(values dicom.Value) string {
	value := values.GetValue()
	switch value.(type) {
	case []int:
		return fmt.Sprintf("%d", value.([]int)[0])
	case []string:
		return value.([]string)[0]
	}
	return values.String()
}

func findAllTags(names []string) []tag.Info {
	tags := make([]tag.Info, 0)
	for _, tagName := range names {
		info, err := tag.FindByName(tagName)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Unable to locate tag: %s\n", tagName)
		} else {
			tags = append(tags, info)
		}
	}
	return tags
}
