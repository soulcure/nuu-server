# -*- coding: utf-8 -*-
# email依赖
import smtplib
# 短信依赖
import sys
from email.header import Header
from email.mime.text import MIMEText
from email.utils import parseaddr, formataddr

sys.path.append("/usr/lib/python2.7/site-packages/")
from aliyunsdkcore.client import AcsClient
from aliyunsdkcore.profile import region_provider
import const

REGION = "cn-hangzhou"
PRODUCT_NAME = "Dysmsapi"
DOMAIN = "dysmsapi.aliyuncs.com"
acs_client = AcsClient(const.ACCESS_KEY_ID, const.ACCESS_KEY_SECRET, REGION)
region_provider.add_endpoint(PRODUCT_NAME, REGION, DOMAIN)


def _format_addr(s):
    name, addr = parseaddr(s)
    return formataddr(( \
        Header(name, 'utf-8').encode(), \
        addr.encode('utf-8') if isinstance(addr, unicode) else addr))


def send_email(pw, from_addr, password, stmp_server):
    to_addr = '287268221@qq.com'
    # cc_addr = 'nuujiegui@163.com'

    msg = MIMEText('Your password is : ' + pw, 'plain', 'utf-8')  # 邮件内容
    msg['From'] = _format_addr(from_addr)
    msg['To'] = ",".join(to_addr)
    # msg['Cc'] = _format_addr(cc_addr)
    msg['Subject'] = Header('Forget Password', 'utf-8').encode()  # 邮件标题

    server = smtplib.SMTP_SSL(stmp_server, 465)  # 默认是25
    server.set_debuglevel(1)
    server.login(from_addr, password)
    server.sendmail(from_addr, to_addr, msg.as_string())
    server.quit()
    print "send email from:%s to:%s" % (from_addr, to_addr)
