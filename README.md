## drive2photos

A tool to transfer missing media (photo,video) from Google Drive to Google Photos.


### requirements

A `credentials.json` file which is the exported OAuth2 API Key from a Google Cloud project.
See https://developers.google.com/drive/api/quickstart/go.

### commands

|command|description|
|----|----|
|:q  |quit|
|:p  |photo listing enabled|
|:f  |folder listing enabled|
|ls  |list the contents of the current folder|
|cd [name] | change to the subfolder |
|cd .. | change to the parent folder |
|cp [name] | copy the photo to Google Photos (unless exists)
|rm [name] | remove the photo to Google Drive
|mv [name] | move the photo from Google Drive to Google Photos
|ff [name] | find file on Google Photos