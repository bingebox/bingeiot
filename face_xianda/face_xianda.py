#!/usr/bin/env python
#coding=utf-8

import sys
import requests
import time


########################################################################################################################
face_url = 'http://119.29.147.210:16781'


########################################################################################################################
def add_face_repository(request, repository_name):
    url = '%s/repository' % face_url
    r = requests.post(url, json = {'name': repository_name, 'extra_meta': 'test_for_face'})
    print (r.status_code)
    print (r.text)

def del_face_repository(request, repository_id):
    url = '%s/repository?id=%s' % (face_url, repository_id)
    r = requests.delete(url)
    print (r.status_code)
    print (r.text)

def usage():
    print ('usage: ./face_xianda.py [cmd] [args...]')
    print ('    add_repo [repository_name]')
    print ('    del_repo [repository_id]')

def main():
    print ('---------------Face XianDa Test begin----------------------')
    if len(sys.argv) < 2:
        usage()
        return

    request = requests.Session()
    if sys.argv[1] == 'add_repo' and len(sys.argv) >= 3:
        add_face_repository(request, sys.argv[2])
    if sys.argv[1] == 'del_repo' and len(sys.argv) >= 3:
        del_face_repository(request, sys.argv[2])
    else:
        usage()
    print ('---------------Face XianDa Test end----------------------')

if __name__ == '__main__':
    main()
