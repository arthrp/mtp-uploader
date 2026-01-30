# mtp-uploader

Simple app to upload files to MTP devices. Useful in case when you know what and where exactly you want to upload. Compared to GUI MTP clients (like OpenMTP or your file manager if your OS supports MTP out of the box) you don't need to navigate to source and destination folders. 
All it takes is one command.

You need to install libusb to use it.

```
mtp-uploader -u /tmp/localfile.txt /Download
```
