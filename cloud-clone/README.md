### **Web Resource Cloud Clone server**
一个线上资源转储服务，提供将网络资源转储到服务器的功能，支持网络文件和Git仓库克隆转储，支持资源转储的状态查询，含任务状态、进度、速度等。  
这是<font color="#FF0000">mm-wiki-ex</font> [<https://github.com/simonblowsnow/mm-wiki-ex.git>] 项目的一个组件服务, 该项目是一个知识及资源共享平台。

####**服务执行逻辑：**
+ 接收下载请求（请求参数为网络文件url或Git仓库地址）
+ 直接返回jobId
+ 启动线程下载资源，4次重试
+ 记录状态，写入result.csv
+ 接受jobId查询状态

####**任务状态标识：**
+ 0 - 未执行
+ 1 - 执行中
+ 2 - 已完成
+ 3 - 已失败
+ -1 - 已失败

####**资源存放目录：**
在 src/config.py 中配置
Notice：任务将资源下载到统一目录，并在任务状态中标识了目标资源的存放路径，若需将资源存放至其它路径，请后续迁移。


####**其它：**
pip install -r requirements 
或 
pip3.7 install flask flask_cors colorlog gitpython requests

基于Python3，Flask框架，使用multiprocessing管理线程，使用Manager共享数据，使用RLock同步，使用UUID命名资源。
未配置生产环境部署，未使用数据库和Redis，故不能多实例、分布式部署。

**启动：**./start.bat (On Windows) ， ./start.sh (On Linux)
**测试：**
+ 服务状态：<http://localhost:8885/>
+ 文件：<http://localhost:8885/cloneRes?isFile=1&url=https://www.baidu.com/img/flexible/logo/pc/result.png>
+ Git仓库: <http://localhost:8885/cloneRes?isFile=0&url=https://github.com/simonblowsnow/cloud-clone.git>
+ 查询：<http://localhost:8885/getStatus?jobId=1>

