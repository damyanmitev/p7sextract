# p7sExtract 

####Extracts data from .p7s file

Users select .p7s file, whose data is to be extracted in one of three ways:
- opening a .p7s file with file chooser dialog box
- drag & drop of .p7s file onto the main window
- passing the path of .p7s file as command line parameter

## Building
p7sExtract currently requires Go 1.11.x or later.

It uses [Walk](https://github.com/lxn/walk) , [GoVersionInfo](https://github.com/josephspurrier/goversioninfo/) and [CMS](https://github.com/mastahyeti/cms)

### Get the dependencies
```
go get github.com/mastahyeti/cms
go get github.com/lxn/walk
go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
```

### Regenerate the resource file
Execute 
```
go generate
```
or start `build_resources.bat`

### Build the executable
Execute
```
go build -ldflags="-H windowsgui" -o p7sExtract.exe
```
or start `build_exe.bat`

### Get the dependencies and build with one command
Start `build_all.bat`

