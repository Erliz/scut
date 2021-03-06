SCut / Screen Cut
===
SCut is a server to hold your screenshots and a client to post them on it

## Usage
* Start server to receive and store files
* Start client to watch directory in what you create screenshots and auto upload them
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
docker run -e "URL=http://localhost:3000/" -v `pwd`/storage:/app/storage erliz/scut
```

## Client
`WorkDir` must be an absolute path with trailing slash
```
scut -w `pwd`/storage/
```

### Upload to dedicated server
`URL` must be with trailing slash
```
scut -w `pwd`/storage/ -u http://example.com/
```

### Remove file after upload
```
scut -w `pwd`/storage/ -r
```

### Other helpfull staff
```
scut --help
```

### Post upload
* Open response url in browser after upload
```
scut -w `pwd`/storage/ -c open
```
* Create your own post script
```
echo '#!/bin/bash
open $1
echo $1 | pbcopy' > scut-post.sh
chmod +x scut-post.sh
scut -w `pwd`/storage/ -c scut-post.sh
```

## Bugs
* Client `workdir` must be an absolute path
* Not work opening file in Docker, cause `open` file not forwarding to Docker
```
docker run --rm -it -v ~/Desktop:/storage:ro -v /usr/bin/open:/usr/bin/open erliz/scut-client app -u http://remote.host/ --verbose -m -c /usr/bin/open
```
