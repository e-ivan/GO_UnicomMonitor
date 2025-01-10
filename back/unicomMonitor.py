import asyncio
import math

import websockets, time
import urllib.parse
import base64

from urllib.parse import unquote

def encode(s):
    t = math.ceil(len(s)*1.0 / 2)
    shifted_string = s[t:] + s[:t]
    escaped_string = bytes(shifted_string, 'utf-8').decode('unicode_escape')
    encoded_string = base64.b64encode(escaped_string.encode()).decode()
    # random_number = str(random.randint(1e10, 1e11-1))
    # random_number_encoded = base64.b64encode(random_number.encode()).decode()
    # result = random_number_encoded[:10] + encoded_string
    result = 'MTc2NDAxND' + encoded_string
    return result


def decode(encoded_string):
    encoded_string = encoded_string[10:]
    decoded_string = base64.b64decode(encoded_string).decode()
    unescaped_string = bytes(decoded_string, 'utf-8').decode('unicode_escape')
    unescaped_string = unquote(unescaped_string)
    t = math.ceil(len(unescaped_string)*1.0 / 2)
    original_string = unescaped_string[t:] + unescaped_string[:t]
    return original_string


global_var = b''

async def savedata():
    global global_var
    if len(global_var) > 1024 * 1024 * 1:
        with open('video.flv', 'ab') as file:
            file.write(global_var)
        global_var = b''
async def savevoice():
    global global_voice
    if len(global_voice) > 1024 * 1024 * 1:
        with open(str(int(time.time())) + '.flv', 'ab') as file:
            file.write(global_voice)
        global_voice = b''
ditc = {}
async def send_message():
    global global_var,global_voice,ditc
    uri = "wss://vd-file-hnzz2-wcloud.wojiazongguan.cn:50443/h5player/live"

    async with websockets.connect(uri) as websocket:
        # 发送消息
        message = "_paramStr_="
        await websocket.send(message)
        # 接收服务器响应
        response = await websocket.recv()
        print(urllib.parse.unquote(response.decode()))
        response = await websocket.recv()
        print(urllib.parse.unquote(response.decode()))
        await websocket.send('{\"time\":1243,\"cmd\":3}')
        while True:
            response = await websocket.recv()
            if response[1] == 0x63:
                global_var += response[0x4e:]
                print(len(global_var))
                await savedata()



asyncio.get_event_loop().run_until_complete(send_message())


