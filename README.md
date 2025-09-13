# Jason
Jason is a small CLI-based JSON viewer/editor.

Put simply, it allows for quick editing of JSON files, as well as with undo/redo capabillites.
That's about it, just a basic lil Go program

# **Building**
To begin you need the jason.go file with:
```shell
git clone github.com/SHAPeS-Software/Jason.git
```
then go into the directory with:
```shell
cd Jason
```
Jason uses liner by peterh, so initalize a new go project with:
```shell
go mod init example.com
go mod tidy
```
AND finally run:
```shell
go build jason.go
```
After adding the executable to your path, you can test with:
```shell
jason help
```
Now, Jason has a dedicated shell which is opened when you open a JSON file.
Open up a JSON file with
jason open file.json
This initlizes a Jason shell
Jason>
Use help to view all the commands avaliable.
# Is that it?
Yeah, basically. You can build, modify and do about anything you want with the code, have fun!
