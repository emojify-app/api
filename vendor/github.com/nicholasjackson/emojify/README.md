![Emojify](emojify.png)

Inspired by [Mat Ryer's](https://github.com/matryer) blog post [https://medium.com/@matryer/anonymising-images-with-go-and-machine-box-fd0866adb9f5](https://medium.com/@matryer/anonymising-images-with-go-and-machine-box-fd0866adb9f5)  

Emojify is an http server which replaces faces in an image with emoji.  

[https://www.emojione.com](https://www.emojione.com) Thanks to EmojiOne for providing free emoji icons

![Example](./output.png)

## Example
Work in progress...

```
$ docker run -p 8080:8080 -e "MB_KEY=$MB_KEY" machinebox/facebox                                                                        
$ FACEBOX=http://192.168.99.100:8080 go run main.go
$ curl localhost:9090 -d "http://i2.cdn.cnn.com/cnnnext/dam/assets/160504160857-03-donald-trump-0504-large-169.jpg" -o output.png
```
