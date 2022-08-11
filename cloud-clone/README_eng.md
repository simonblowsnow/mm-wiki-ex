### **Web Resource Cloud Clone server**
An online resource dump service provides the function of dumping network resources to the server, supports network files and git warehouse clone store to server, and supports the status query of resource clone dump, including task status, progress, speed, etc. 
This is a component server For [<font color="#FF0000">mm-wiki-ex</font> <https://github.com/simonblowsnow/mm-wiki-ex.git>], a Knowledge & Resource Share Platform.

####**Service execution logic:**
+ Receive the download request (the request parameter is the network file URL or the GIT warehouse address)
+ Directly return jobid
+ Start the thread to download resources, and try again for 4 times
+ Record the status and write result.csv
+ Accept jobid query status

####**Task status ID:**
+ 0 - not executed
+ 1 - in progress
+ 2 - completed
+ 3 - failed
+ -1 - failed

####**Resource storage directory:**
In src/config PY
**Notice**: the task downloads the resources to the unified directory and identifies the storage path of the target resources in the task status. If you need to store the resources to other paths, please migrate later.


####**Others:**
pip install -r requirements
Or
pip3.7 install flask flask_cors colorlog gitpython requests

Based on Python 3, the flask framework uses multiprocessing management threads, uses manager to share data, uses RLOCK synchronization, and uses UUID to name resources.
The production environment deployment is not configured, and the database and redis are not used, so multi instance and distributed deployment is not possible.

**Start: ** ./start.bat (On Windows) ， ./start.sh (On Linux)
**Test:**
+ Service status:< http://localhost:8885/>
+ File:< http://localhost:8885/cloneRes?isFile=1&url=https://www.baidu.com/img/flexible/logo/pc/result.png>
+ Git warehouse:< http://localhost:8885/cloneRes?isFile=0&url=https://github.com/simonblowsnow/cloud-clone.git>
+ Query:< http://localhost:8885/getStatus?jobId=1>

