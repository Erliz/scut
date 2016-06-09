Cut
===
Cut is a server to hold your screenshots and a client to post them on it

## Usage
* Start server to receive and store files
* Start client to watch directory in what you create screenshots
* Make screenshot (cmd+shift+3)
* Share your direct short image url with friends and colleague
* No more ADs! :tada:
## Server

#### Local
```
npm start
```
#### Docker
`URL` must be with trailing slash
```
docker run -e "URL=http://localhost:3000/" -v `pwd`/storage:/app/storage erliz/cut
```

## Client
`WorkDir` must be an absolute path with trailing slash
```
cut -w `pwd`/storage/
```

### Upload to dedicated server
`URL` must be with trailing slash
```
cut -w `pwd`/storage/ -u http://example.com/
```

### Remove file after upload
```
cut -w `pwd`/storage/ -r
```

### Other helpfull staff
```
cut --help
```

### Post upload
* Open response url in browser after upload
```
cut -w `pwd`/storage/ -c open
```
* Create your own post script
```
echo '#!/bin/bash
open $1
echo $1 | pbcopy' > cut-post.sh
chmod +x cut-post.sh
cut -w `pwd`/storage/ -c cut-post.sh
```

## Bugs
* Client `workdir` must be an absolute path
* Not work opening file in Docker, cause `open` file not forwarding to Docker
```
docker run --rm -it -v ~/Desktop:/storage:ro -v /usr/bin/open:/usr/bin/open erliz/cut-client app -u http://remote.host/ --verbose -m -c /usr/bin/open
```
