const fs = require('fs');
const express = require('express');
const RandomString = require('randomstring');
const FileType = require('file-type');

var app = express();

var config = {
  workDir: __dirname + '/storage/',
  host: process.env.URL ? process.env.URL : 'http://localhost:3000/',
  port: process.env.PORT ? parseInt(process.env.PORT) : 3000,
  newFileNameLength: process.env.NAME_LENGTH ? parseInt(process.env.NAME_LENGTH) : 7,
};

console.log('config', config);

function fileReceiveMiddleware(req, res, next) {
  var content = [];
  var fileType;
  req.on('data', function(chunk){
    content.push(chunk);
    if (!fileType) {
      fileType = FileType(Buffer.concat(content));
    }
  });
  req.on('end', function () {
    req.file = {
      content: Buffer.concat(content),
      type: fileType,
    };
    next();
  });
}

var retryCount = 0;
const maxRetryCount = 10;
function writeFile(file) {
  return new Promise(function (resolve, reject) {
    if (!file.content.length) {
      return reject(new Error('Empty file'));
    }
    var newFileName = RandomString.generate(config.newFileNameLength) + "." + file.type.ext;
    var path = config.workDir + newFileName;
    new Promise(function(resolve, reject) {
      fs.stat(path, function(err, data) {
        if (err) {
          if (err.code == 'ENOENT') {
            return resolve();
          }
          return reject(err);
        }
        reject(new Error('File "' + newFileName + '" already exists'));
      });
    }).then(function(){
      fs.writeFile(path, file.content, 'binary', function(err) {
        if (err) {
          return reject(err);
        }
        console.log('write in to: ' + path);
        resolve(config.host + newFileName);
      });
    }).catch(function(err){
      console.error(err);
      if (retryCount < maxRetryCount) {
        retryCount++;
        writeFile(file)
          .catch(function(err){
            reject(err);
          });
      }
      reject(new Error('Too many attempts to create file'));
    });
  });
}

app.use(express.static(config.workDir));
app.get('/', function (req, res) {
  res.send('and CUT!');
});
app.put('/:file', fileReceiveMiddleware, function (req, res) {
  console.log('file name input: ' + req.params.file);
  retryCount = 0;
  writeFile(req.file)
    .then(function(url) {
      console.log('response: ' + url);
      res.send(url);
    })
    .catch(function(err) {
      console.error(err);
      res.status(500).send(err.message);
    });
});

app.listen(config.port, function () {
  console.log('Example app listening on port 3000!');
});

process.on('SIGINT', function() {
  console.log('\ncaught SIGINT, stopping gracefully');
  process.exit();
});
process.on('SIGTERM', function() {
  console.log('\ncaught SIGTERM, stopping gracefully');
  process.exit(1);
});
