# 操作系统 （Operating System） 
- 本质就是：管理计算机所有资源的软件， 包含：
 1. CPU 
 2. Memory 
 3. Disk 磁盘
 4. Network
 5. Process 
 6. Thread 线程
 7. File
 8. Device

## 用户态（User）、内核态（Kernel）
```sh
Application
System call
Kernel
Hardware 硬件

# System call :
read()
write()
socket()
accept()
send()
recv()
``` 

## 进程（Process）
- 运行中的程序， 包括： Code Data Heap Stack PCB(Process Control Block)

## 线程（Thread）
```sh
Process
  |- Thread1
  |- Thread2
  |- ...
# 特点
共享：
  Heap 堆
  Code
  Global Variable
独享：
  Stack 栈
  Register
  PC  
```

## 进程 vs 线程
- 进程： 资源独立、切换开销大、通信复杂、稳定性高
- 线程： 共享资源、切换开销小、通信方便、崩溃影响整个进程

