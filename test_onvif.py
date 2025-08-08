#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
ONVIF服务测试脚本
用于验证GO_UnicomMonitor的ONVIF功能是否正常工作
"""

import requests
import xml.etree.ElementTree as ET
import sys

def test_onvif_device_info(host="localhost", port=8080, username=None, password=None):
    """测试设备信息获取"""
    print("=== 测试ONVIF设备信息 ===")
    
    url = f"http://{host}:{port}/onvif/device_service"
    
    # ONVIF GetDeviceInformation请求
    soap_request = """<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <tds:GetDeviceInformation xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
  </soap:Body>
</soap:Envelope>"""
    
    headers = {
        'Content-Type': 'application/soap+xml; charset=utf-8',
        'SOAPAction': '"http://www.onvif.org/ver10/device/wsdl/GetDeviceInformation"'
    }
    
    # 如果提供了认证信息，添加到请求中
    auth = None
    if username and password:
        auth = (username, password)
    
    try:
        response = requests.post(url, data=soap_request, headers=headers, auth=auth, timeout=10)
        print(f"状态码: {response.status_code}")
        print(f"响应内容:\n{response.text}")
        
        if response.status_code == 200:
            print("✓ 设备信息获取成功")
            return True
        elif response.status_code == 401:
            print("✗ 认证失败，请检查用户名和密码")
            return False
        else:
            print("✗ 设备信息获取失败")
            return False
            
    except Exception as e:
        print(f"✗ 请求失败: {e}")
        return False

def test_onvif_capabilities(host="localhost", port=8080, username=None, password=None):
    """测试设备能力获取"""
    print("\n=== 测试ONVIF设备能力 ===")
    
    url = f"http://{host}:{port}/onvif/device_service"
    
    # ONVIF GetCapabilities请求
    soap_request = """<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <tds:GetCapabilities xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
      <tds:Category>All</tds:Category>
    </tds:GetCapabilities>
  </soap:Body>
</soap:Envelope>"""
    
    headers = {
        'Content-Type': 'application/soap+xml; charset=utf-8',
        'SOAPAction': '"http://www.onvif.org/ver10/device/wsdl/GetCapabilities"'
    }
    
    # 如果提供了认证信息，添加到请求中
    auth = None
    if username and password:
        auth = (username, password)
    
    try:
        response = requests.post(url, data=soap_request, headers=headers, auth=auth, timeout=10)
        print(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            print("✓ 设备能力获取成功")
            return True
        elif response.status_code == 401:
            print("✗ 认证失败，请检查用户名和密码")
            return False
        else:
            print("✗ 设备能力获取失败")
            return False
            
    except Exception as e:
        print(f"✗ 请求失败: {e}")
        return False

def test_onvif_profiles(host="localhost", port=8080, username=None, password=None):
    """测试配置文件获取"""
    print("\n=== 测试ONVIF配置文件 ===")
    
    url = f"http://{host}:{port}/onvif/device_service"
    
    # ONVIF GetProfiles请求
    soap_request = """<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <trt:GetProfiles xmlns:trt="http://www.onvif.org/ver10/media/wsdl"/>
  </soap:Body>
</soap:Envelope>"""
    
    headers = {
        'Content-Type': 'application/soap+xml; charset=utf-8',
        'SOAPAction': '"http://www.onvif.org/ver10/media/wsdl/GetProfiles"'
    }
    
    # 如果提供了认证信息，添加到请求中
    auth = None
    if username and password:
        auth = (username, password)
    
    try:
        response = requests.post(url, data=soap_request, headers=headers, auth=auth, timeout=10)
        print(f"状态码: {response.status_code}")
        print(f"响应内容:\n{response.text}")
        
        if response.status_code == 200:
            print("✓ 配置文件获取成功")
            return True
        elif response.status_code == 401:
            print("✗ 认证失败，请检查用户名和密码")
            return False
        else:
            print("✗ 配置文件获取失败")
            return False
            
    except Exception as e:
        print(f"✗ 请求失败: {e}")
        return False

def test_stream_access(host="localhost", port=8080, device_name="客厅", username=None, password=None):
    """测试流访问"""
    print(f"\n=== 测试流访问 ({device_name}) ===")
    
    url = f"http://{host}:{port}/onvif/stream/{device_name}"
    
    # 如果提供了认证信息，添加到请求中
    auth = None
    if username and password:
        auth = (username, password)
    
    try:
        response = requests.get(url, auth=auth, timeout=10)
        print(f"状态码: {response.status_code}")
        print(f"内容类型: {response.headers.get('Content-Type', 'Unknown')}")
        print(f"数据长度: {len(response.content)} bytes")
        
        if response.status_code == 200:
            print("✓ 流访问成功")
            return True
        elif response.status_code == 401:
            print("✗ 认证失败，请检查用户名和密码")
            return False
        else:
            print("✗ 流访问失败")
            return False
            
    except Exception as e:
        print(f"✗ 请求失败: {e}")
        return False

def main():
    """主函数"""
    print("GO_UnicomMonitor ONVIF功能测试")
    print("=" * 50)
    
    # 获取命令行参数
    host = "localhost"
    port = 8080
    username = None
    password = None
    
    if len(sys.argv) > 1:
        host = sys.argv[1]
    if len(sys.argv) > 2:
        port = int(sys.argv[2])
    if len(sys.argv) > 3:
        username = sys.argv[3]
    if len(sys.argv) > 4:
        password = sys.argv[4]
    
    print(f"测试目标: {host}:{port}")
    if username and password:
        print(f"认证信息: {username}:{password}")
    else:
        print("认证信息: 无（如果服务需要认证，测试可能失败）")
    print()
    
    # 执行测试
    tests = [
        test_onvif_device_info(host, port, username, password),
        test_onvif_capabilities(host, port, username, password),
        test_onvif_profiles(host, port, username, password),
        test_stream_access(host, port, "客厅", username, password)
    ]
    
    # 输出测试结果
    print("\n" + "=" * 50)
    print("测试结果汇总:")
    print(f"设备信息: {'✓' if tests[0] else '✗'}")
    print(f"设备能力: {'✓' if tests[1] else '✗'}")
    print(f"配置文件: {'✓' if tests[2] else '✗'}")
    print(f"流访问: {'✓' if tests[3] else '✗'}")
    
    success_count = sum(tests)
    total_count = len(tests)
    
    print(f"\n总体结果: {success_count}/{total_count} 测试通过")
    
    if success_count == total_count:
        print("🎉 所有测试通过！ONVIF服务运行正常。")
        return 0
    else:
        print("⚠️  部分测试失败，请检查ONVIF服务配置。")
        return 1

if __name__ == "__main__":
    sys.exit(main())
