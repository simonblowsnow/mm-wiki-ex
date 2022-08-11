#coding: utf8
'''
==================================================================
Created on 2021年12月13日 By Simon
    pip3.7 install gitpython
==================================================================
'''

import time
from datetime import datetime
import os
import uuid
import csv
import git
from git import RemoteProgress
import requests
from multiprocessing import Process, RLock
from src.config import Config as C
from src.libs.log import L

lock = RLock()

def TM(): return datetime.strftime(datetime.now(), '%Y-%m-%d %H:%M:%S')



class CloneProgress(RemoteProgress):
    def __init__(self, jobs, _id):
        git.RemoteProgress.__init__(self)
        self.jobs = jobs
        self.id = _id
        
    def update(self, op_code, cur_count, max_count=None, message=''):
        if message:
            self.jobs[self.id]["progress"] = "%s,%.2f%s" % (message, 100.0 * cur_count / max_count, "%")
        '''End If'''
   
def download_file(url, name, jobs, _id):
    headers = {'Proxy-Connection':'keep-alive'}
    r = requests.get(url, stream=True, headers=headers)
    length = float(r.headers['content-length'])
    with open(name, 'wb') as f:
        count, count_tmp = 0, 0
        time1 = time.time()
        for chunk in r.iter_content(chunk_size = 512):
            if not chunk: continue
            f.write(chunk)
            count += len(chunk)
            if time.time() - time1 > 2:
                p = count / length * 100
                speed = (count - count_tmp) / 1024 / 1024 / 2
                count_tmp = count
                time1 = time.time()
                jobs[_id]["progress"] = '%.2f%s%.2fM/S' % (p, ',',speed)
    '''End With'''



class CloneJob(Process):
    def __init__(self, G, url, _id, is_file=False):
        super().__init__()
        self.id = _id
        self.url = url
        self.status = 0
        self.is_file = is_file
        self.folder = str(uuid.uuid1())
        job = G.mg.dict()
        job['url'], job['status'], job['times'], job['progress'] = url, 0, 1, ""
        job['id'], job['folder'] = _id, self.folder
        job['startTime'], job['completeTime'] = time.time(), None
        G.jobs[self.id] =job
        self.jobs = G.jobs
        
    def run(self):
        self.jobs[self.id]['status'] = 1
        dst = C.FILE_ROOT + "/" + self.folder
        if not self.is_file and not os.path.exists(dst): os.mkdir(dst)
        for i in range(3):
            if i > 0: L.info("The %sth time Request: %s now" % (i + 1, self.url))
            '''核心下载逻辑'''
            try:
                self.download(dst) if self.is_file else self.clone(dst) 
                self.status = 2
                '''End If'''
            except Exception as _:
                self.jobs[self.id]['times'] += 1
                self.status = -1
                print(_)
            '''End Try'''
       
            if self.status == 2: break
        '''End For'''
        self.jobs[self.id]['status'] = self.status
        self.write_result()
        
    '''Git仓库'''
    def clone(self, dst):
        git.Repo.clone_from(self.url, dst, progress=CloneProgress(self.jobs, self.id))
        
    '''下载文件'''
    def download(self, dst):
        download_file(self.url, dst + ".tmp", self.jobs, self.id)
       
    def write_result(self):
        L.info("Job %s run End with status %s." % (self.id, self.status))
        completeTime = TM()
        self.jobs[self.id]['completeTime'] = completeTime
        
        lock.acquire()
        with open("./result.csv", "a") as f:
            line = '%s,%s,%s,%s,%s\n' % (self.id, self.status, self.folder, self.url, completeTime)
            f.write(line)
        '''End With'''
        lock.release()
        
    @staticmethod    
    def load_history(G):
        G.idx, idx = 0, 0
        if not os.path.exists("./result.csv"): return 
        with open("./result.csv", "r") as f:
            for d in csv.reader(f, skipinitialspace=True):
                _id, status, folder = int(d[0]), int(d[1]), d[2]
                url, tm = d[3], d[4]
                info = G.mg.dict()
                info['id'], info['status'], info['folder'] = _id, status, folder
                info['url'], info['progress'], info['completeTime'] = url, None, tm
                if _id > idx: idx = _id
                G.jobs[_id] = info
        '''End With'''
        G.idx = idx + 1
'''========================================End Class===================================================='''


def clone_res(G, url, is_file=False):
    _id = G.get_id()
    cj = CloneJob(G, url, _id, is_file)
    cj.start()
    return _id
    
    

if __name__ == '__main__':
    import ctypes
    from multiprocessing import Manager, freeze_support
    from src.libs.conf import GlobalData as G
    freeze_support()
    mg = Manager()
    url = "http://localhost:82/Temp/wm-data-net.zip"
    url2 = "https://github.com/rishit-singh/OpenShell.git"
    G.mg, G.idx, G.jobs = mg, mg.Value(ctypes.c_int, 0), mg.dict()
    CloneJob.load_history(G)
    clone_res(G, url, True)
    
    i = 0
    while i < 100:
        time.sleep(1)
        i += 1
    
    pass
