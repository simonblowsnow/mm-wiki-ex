#coding:utf8
import os
import time
import logging
import colorlog
import sys

from src.libs.conf import GlobalConf
sys.path.append('..')
sys.path.append('../..')


color_config = {
    'DEBUG': 'white',  # cyan white
    'INFO': 'cyan',
    'WARNING': 'yellow',
    'ERROR': 'red',
    'CRITICAL': 'bold_red',
}



'''本地日志配置'''
def init():
    if GlobalConf.Log: return GlobalConf.Log
    '''
    logging.basicConfig(level=logging.DEBUG,
        format = '%(message)s - %(filename)s[line:%(lineno)d] - %(levelname)s: %(asctime)s',
        datefmt = '%Y-%m-%d %H:%M:%S',
        )
    '''

    log_path = os.path.dirname(os.path.dirname(__file__)) + '/logs/log'
    log_name = log_path + time.strftime("%Y-%m-%d_%p", time.localtime()) + '.txt'
    f_handle = logging.FileHandler(log_name)
    f_handle.setLevel(logging.DEBUG)
    
    c_handle = logging.StreamHandler()
    c_handle.setLevel(logging.DEBUG)
    
    datefmt = '%Y-%m-%d %H:%M:%S'
    f_formatter = logging.Formatter('%(asctime)s - %(filename)s[line:%(lineno)d] - %(levelname)s: %(message)s')
    c_formatter = colorlog.ColoredFormatter(
        fmt='%(log_color)s[%(levelname)s] %(message)s - %(filename)s[line:%(lineno)d], %(asctime)s', 
        datefmt=datefmt, log_colors=color_config)
    
    c_handle.setFormatter(c_formatter)
    f_handle.setFormatter(f_formatter)
    
    logger = colorlog.getLogger('example')
    logger.addHandler(c_handle)
    logger.addHandler(f_handle)
    # logging.config.fileConfig('logging.conf')
    logger.setLevel(logging.DEBUG)
    
    GlobalConf.Log = logger
    c_handle.close()
    f_handle.close()
    
    return GlobalConf.Log

L = init()
