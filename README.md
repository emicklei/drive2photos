## drive2photos

A tool to transfer missing media (photo,video) from Google Drive to Google Photos.


### requirements

A `credentials.json` file which is the exported OAuth2 API Key from a Google Cloud project.
See https://developers.google.com/drive/api/quickstart/go.


### install

    go install github.com/emicklei/drive2photos@latest

### commands

|command|description|-gen@latest

### commands

|command|description|
|----|----|
|:q  |quit|
|:p  |photo listing enabled|
|:f  |folder listing enabled|
|ls  |list the contents of the current folder|
|cd [name] | change to the subfolder or a computer name |
|cd .. | change to the parent folder |
|cp [name] | copy the media to Google Photos (unless exists)
|rm [name] | remove the media from Google Drive
|mv [name] | move the media from Google Drive to Google Photos
|ff [name] | find the media file on Google Photos

For the commands `cp,rm`, the argument can be the wildcard character `*`

(c) 2023, https://ernestmicklei.com. MIT License.