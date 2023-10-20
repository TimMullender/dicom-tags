# DICOM Tags

I frequently want to extract a set of tags from multiple DICOM files so that I can analyse them

I came across Suyash Kumar's [DICOM module](https://github.com/suyashkumar/dicom), so I put this together...

```
$ dicom-tags --help
Walks a directory and prints the selected tags for each DICOM that is found

Usage:
  dicom-tags [folder tag-list] [flags]

Flags:
  -e, --exclusion strings   Exclude paths using glob
  -h, --help                help for dicom-tags
  -n, --numeric             Sort by the first tag numerically
  -s, --sort                Sort by the first tag

```

An example of pulling the image size and transfer syntax from a ZIP file in a folder called store and sorting them numerically
```
$ dicom-tags -n -e "**/*.json" store Rows Columns TransferSyntaxUID | head
Filename,Rows,Columns,TransferSyntaxUID
store/random.zip#1.2.3.4.10.1.2620998974.1277529672.185853571.295812190.dcm,64,64,1.2.840.10008.1.2
store/random.zip#1.2.3.4.10.1.3628331974.1213501031.4272008843.996711022.dcm,64,64,1.2.840.10008.1.2
store/random.zip#1.2.3.4.10.1.624798435.1131047349.1574243718.1082197643.dcm,64,64,1.2.840.10008.1.2
store/random.zip#1.2.3.4.10.1.509193296.1236787885.818108066.3499083941.dcm,304,364,1.2.840.10008.1.2
store/random.zip#1.2.3.4.10.1.3745541949.1251310250.1938151339.2270750292.dcm,364,364,1.2.840.10008.1.2
store/random.zip#1.2.3.4.10.1.3867731315.1282058466.975221437.1921460290.dcm,512,512,1.2.840.10008.1.2
store/random.zip#1.2.3.4.10.1.2162523366.1281643216.595071644.776254116.dcm,512,512,1.2.840.10008.1.2
store/random.zip#1.2.3.4.10.1.198187268.1119644752.1221044877.2327558677.dcm,512,512,1.2.840.10008.1.2
store/random.zip#1.2.3.4.10.1.4283601024.1120931215.2369348769.2304880576.dcm,512,512,1.2.840.10008.1.2
```