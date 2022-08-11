#coding: utf8
'''
==========================================================================================
Created on 2020-11-19 15:30:55
@author: Simon
==========================================================================================
'''


class GlobalConf():
    Log = None
    

class GlobalData():
    mg = None
    idx = None
    jobs = None
    
    @staticmethod
    def get_id():
        GlobalData.idx += 1 
        return GlobalData.idx
'''End Class'''
   

        
    
if __name__ == '__main__':
    pass