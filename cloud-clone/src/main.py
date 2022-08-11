#coding: utf8
'''
==================================================================
Created on 2021年12月13日 By Simon

==================================================================
'''

import ctypes
from multiprocessing import Manager, freeze_support
from src.clone import clone_res, CloneJob
from src.libs.conf import GlobalData as G
from flask import Flask, request
from flask_cors import CORS
from src.config import Config as C
app = Flask(__name__)
CORS(app, supports_credentials=True)

'''
原理：
    1.接收下载请求，返回资源ID
    2.启动异步任务，下载资源，4次重试
    3.记录状态，写入result.csv
    4.接受状态查询

status: 0 - 未执行，1 - 执行中，2 - 已完成，3 - 已失败， -1 - 已失败
'''

@app.route('/')
def index():
    return 'Welcome，This is a cloud resource clone server'
@app.route('/cloneRes', methods=['GET','POST'])
def _clone_res():
    R = request.form if request.method=='POST' else request.args
    url = R.get('url', None)
    is_file = int(R.get('isFile', "1"))
    if not url: return {"error": 1, "message": "参数错误，缺少url"}
    
    _id = clone_res(G, url, is_file)
    return {"error": 0, "data": {"jobId": _id}}

@app.route('/getStatus', methods=['GET','POST'])
def request_status():
    _id = int(request.args.get('jobId', "-1"))
    if _id not in G.jobs: return {"error": 1, "message": "网络错误，请重试！"}
    j = G.jobs[_id]
    return {"error": 0, "data": {"status": j["status"], "folder": j["folder"], 
                                 "url": j["url"], "progress": j["progress"]}}
    
'''==============================End Request==================================='''


if __name__ == '__main__':
    freeze_support()
    mg = Manager()
    G.mg, G.idx, G.jobs = mg, mg.Value(ctypes.c_int, 0), mg.dict()
    CloneJob.load_history(G)
    app.run(host = "0.0.0.0", port = C.PORT)
    
    
    