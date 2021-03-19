import random
import requests
from PIL import Image
from typing import Optional
from pydantic import BaseModel
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles


def isMessageExist(height, width, pixels, binaryBeginCode):
    binaryMessage = ''
    index = 0
    
    for y in range(0, height):
        for x in range(0, width):
            for z in range(0, 3):
                if index < len(binaryBeginCode):
                    if pixels[x,y][z] % 2 == 1:
                        binaryMessage += '1'
                    else:
                        binaryMessage += '0'
                    index += 1
                else:
                    break
            if index == len(binaryBeginCode):
                break
        if index == len(binaryBeginCode):
            break
    
    if binaryBeginCode == binaryMessage:
        return True
    else:
        return False


def messageCapacity(height, width, binaryBeginCode, binaryEndCode):
    totalCapacity = height * width * 3 - len(binaryBeginCode) - len(binaryEndCode)
    return totalCapacity


def readMessage(height, width, pixels, binaryBeginCode, binaryEndCode):
    binaryMessage = ''
    index = 0
    stopTheLoop = False

    for y in range(0, height):
        for x in range(0, width):
            for z in range(0, 3):
                if pixels[x,y][z] % 2 == 1:
                    binaryMessage += '1'
                else:
                    binaryMessage += '0'
                index += 1
                if index > len(binaryBeginCode) + len(binaryEndCode):
                    if binaryMessage[len(binaryMessage)-len(binaryEndCode):] == binaryEndCode:
                        stopTheLoop = True
                        break
            if stopTheLoop:
                break
        if stopTheLoop:
            break
    
    return binaryMessage[len(binaryBeginCode):len(binaryMessage)-len(binaryEndCode)]


def hideMessage(height, width, pixels, binaryBeginCode, binaryMessage, binaryEndCode):
    index = 0
    stopTheLoop = False
    messageToBeHidden = binaryBeginCode + binaryMessage + binaryEndCode

    for y in range(0, height):
        for x in range(0, width):
            newPixel = []
            for z in range(0, 3):
                newPixelColorValue = pixels[x,y][z]
                if index < len(messageToBeHidden):
                    if pixels[x,y][z] % 2 == 1 and messageToBeHidden[index] == '0':
                        newPixelColorValue -= 1
                    if pixels[x,y][z] % 2 == 0 and messageToBeHidden[index] == '1':
                        newPixelColorValue += 1
                newPixel.append(newPixelColorValue)
                index += 1
            pixels[x,y] = (newPixel[0], newPixel[1], newPixel[2])
            if index > len(messageToBeHidden):
                stopTheLoop = True
                break
        if stopTheLoop:
            break
    return pixels


def eraseMessage(height, width, pixels, binaryBeginCode, binaryExistedMessage, binaryEndCode):
    index = 0
    stopTheLoop = False
    lengthOfExistedMessage = len(binaryBeginCode) + len(binaryExistedMessage) + len(binaryEndCode)

    for y in range(0, height):
        for x in range(0, width):
            newPixel = []
            for z in range(0, 3):
                newPixelColorValue = pixels[x,y][z]
                if index < lengthOfExistedMessage:
                    if newPixelColorValue == 255:
                        newPixelColorValue -= random.randint(0,1)
                    else:
                        newPixelColorValue += random.randint(0,1)
                newPixel.append(newPixelColorValue)
                index += 1
            pixels[x,y] = (newPixel[0], newPixel[1], newPixel[2])
            if index > lengthOfExistedMessage:
                stopTheLoop = True
                break
        if stopTheLoop:
            break
    return pixels


def appendMessage(height, width, pixels, binaryBeginCode, binaryExistedMessage, binaryMessage, binaryEndCode):
    index = 0
    stopTheLoop = False
    lengthOfExistedMessage = len(binaryBeginCode) + len(binaryExistedMessage)
    messageToBeHidden = binaryMessage + binaryEndCode
    totalLength = len(messageToBeHidden) + lengthOfExistedMessage

    for y in range(0, height):
        for x in range(0, width):
            newPixel = []
            for z in range(0, 3):
                newPixelColorValue = pixels[x,y][z]
                if index > lengthOfExistedMessage and index < totalLength:
                    if pixels[x,y][z] % 2 == 1 and messageToBeHidden[index - lengthOfExistedMessage] == '0':
                        newPixelColorValue -= 1
                    if pixels[x,y][z] % 2 == 0 and messageToBeHidden[index - lengthOfExistedMessage] == '1':
                        newPixelColorValue += 1
                newPixel.append(newPixelColorValue)
                index += 1
            pixels[x,y] = (newPixel[0], newPixel[1], newPixel[2])
            if index > totalLength:
                stopTheLoop = True
                break
        if stopTheLoop:
            break
    return pixels


def getFilename(image_url):
    filename = image_url[image_url.rfind('/') + 1: image_url.rfind('.') + 1] + 'png'
    return filename


app = FastAPI()

app.mount("/static", StaticFiles(directory="static"), name="static")


@app.get("/")
def showIntro():
    return {"Hello": "World"}


class ImageAndMessage(BaseModel):
    image_url: str
    binaryBeginCode: str
    binaryExistedMessage: Optional[str] = None
    binaryMessage: Optional[str] = None
    binaryEndCode: Optional[str] = None


@app.post("/checkmessage")
def beginCheckMessage(item: ImageAndMessage):
    img = Image.open(requests.get(item.image_url, stream=True).raw)
    pixels = img.load()
    isMessageExistBoolean = isMessageExist(img.height, img.width, pixels, item.binaryBeginCode)
    return {"IsMessageExist": isMessageExistBoolean}


@app.post("/readmessage")
def beginReadMessage(item: ImageAndMessage):
    img = Image.open(requests.get(item.image_url, stream=True).raw)
    pixels = img.load()
    binaryMessage = readMessage(img.height, img.width, pixels, item.binaryBeginCode, item.binaryEndCode)
    return {"BinaryMessage": binaryMessage}


@app.post("/messagecapacity")
def findMessageCapacity(item: ImageAndMessage):
    img = Image.open(requests.get(item.image_url, stream=True).raw)
    capacity = messageCapacity(img.height, img.width, item.binaryBeginCode, item.binaryEndCode)
    return {"MessageCapacity": capacity}


@app.post("/hidemessage")
def beginhideMessage(item: ImageAndMessage):
    img = Image.open(requests.get(item.image_url, stream=True).raw)
    pixels = img.load()
    pixels = hideMessage(img.height, img.width, pixels, item.binaryBeginCode, item.binaryMessage, item.binaryEndCode)
    filePath = "./static/" + getFilename(item.image_url)
    img.save(filePath)
    return {"NewImageURL": filePath[2:]}


@app.post("/erasemessage")
def beginEraseMessage(item: ImageAndMessage):
    img = Image.open(requests.get(item.image_url, stream=True).raw)
    pixels = img.load()
    pixels = eraseMessage(img.height, img.width, pixels, item.binaryBeginCode, item.binaryExistedMessage, item.binaryEndCode)
    filePath = "./static/" + getFilename(item.image_url)
    img.save(filePath)
    return {"NewImageURL": filePath[2:]}


@app.post("/appendmessage")
def beginAppendMessage(item: ImageAndMessage):
    img = Image.open(requests.get(item.image_url, stream=True).raw)
    pixels = img.load()
    pixels = appendMessage(img.height, img.width, pixels, item.binaryBeginCode, item.binaryExistedMessage, item.binaryMessage, item.binaryEndCode)
    filePath = "./static/" + getFilename(item.image_url)
    img.save(filePath)
    return {"NewImageURL": filePath[2:]}

