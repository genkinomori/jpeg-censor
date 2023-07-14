# jpeg-censor

A reversible image censoring tool.

## Build

Install the latest version of [go](https://go.dev/dl/) and run `bash build.sh` in shell.

## Usage

### decoder

This command is for recovering image censored by this toolset.

- First, gather all the image to recover into a single directory, without subdirectories (files in subdirectories would not be processed).
- Then, copy `decoder` executable file to this directory.
- Finally, double-click the `decoder` file. All recovered images will be prefixed `restored_`.

### encoder

This command is for censoring image.

- First, gather all the image to censor into a single directory, without subdirectories (files in subdirectories would not be processed). `.jpg` and `.png` formats are supported, and correct extension name must be added to each file. 
- For each image, make another image file, which file name is the original file name with `_m` suffix (e.g. `abc_m.jpg` for `abc.jpg`). This image has the same content as the original one, except all parts to be censored be painted black (or any other color you like which has a large enough contrast). 
- Then, copy `encoder` executable file to this directory.
- Finally, double-click the `encoder` file. All censored images will be suffixed `_o`.

## License

This software is released under WTFPL version 2.