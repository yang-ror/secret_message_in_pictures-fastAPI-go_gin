# Secret Message in Pictures

Upload a picture to the server, then hide a message in picrures or discover a message from pictures. A new picture will be available to download once a message is hidden in the picture uploaded. This project has 2 parts: a web server writen in go and using the web framework [Gin](https://github.com/gin-gonic/gin) and a image processing API writen by python.

Key features:
  - The message entered by user will be combined with 2 marker strings that indicates beginning and the end of the message.
  - Each character of the combined messages will be encoded to a 8-bit integer(0-255) based on the ASCII table.
  - Send the image url and the combined message in binary form to the image processing API
  - Change the RGB value of each pixels of the image increase or decrease by 1 correspond to the binary message (odd = 1 and even = 0).
  - Reverse this process when discovering a message from pictures.
  - New message can be appended to the existing message or replace it.
  - Existing message can be erased by randomly change the pixel values.
 

### Build based on:
  - Python 3.9 \ pipenv
  - FastAPI \ uvicorn
  - Pillow
  - Go Language
  - Gin-gonic
  - HTML \ CSS \ Javascript
  - JQuery \ Bootstrap

### Things to know:
  - Processed images will be saved as png format(pixel values will be changed if save in jpg format)
  - Only support Latin alphabet and characters in the ASCII table
  - Any changes in image size, format or edited by other software will most likely to lose the secret message.
  - to run the project:
    1. start pipenv in ./image_procesing_API-Python and run "uvicorn main:app"
    2. run "go run server.go" in web_server-Go_Gin
    3. Server will be running on port 8080 on default
