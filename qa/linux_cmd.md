# cmd

## 查看进程

ps(process status) 命令是 Linux 下最常用的进程查看工具，使用该命令可以确定哪些进程正在运行和运行的状态、
进程是否结束、进程有没有僵尸、哪些进程占用了过多的资源等等。

```
[root@ip-172-27-122-2 ~]# ps
  PID TTY          TIME CMD
58541 pts/1    00:00:00 bash
58738 pts/1    00:00:00 ps
```
- PID：表示该进程的唯一 ID 号
- TTY 或 TT：启动进程的终端名
- TIME：该进程使用 CPU 的累计时间
- CMD：该进程所运行的命令

options:
- -e：显示系统内所有进程的信息。与 -A 选项功能相同
- -f：使用完整 (full) 的格式显示进程信息
- a：显示当前终端下的所有进程信息，包含其他用户的进程信息。和 x 选项结合使用可以显示系统中所有进程的信息
- x：显示当前用户在所有终端下的进程信息
- u：使用以用户为主的格式输出进程信息
- 排序`ps aux --sort=-%mem | head -n 10`

# 文本处理

**grep擅长查找功能，sed擅长取行和替换。awk擅长取列。**

- 排序 `sort test.txt`
- 去掉相邻的重复行 `sort test.txt | uniq`
- 文本行去重并按重复次数排序 `sort test.txt | uniq -c`
- 对文本行按重复次数进行排序 `sort test.txt | uniq -c | sort -rn`(倒序)
- 每行前面的删除重复次数。`sort test.txt | uniq -c | sort -rn | cut -c 9-`(可以看出前面的重复次数占8个字符，因此，可以用命令cut -c 9- 取出每行第9个及其以后的字符。)

## awk

```
awk [-F field-separator] 'commands' input-file(s)
```
[-F 分隔符]是可选的，因为awk使用空格，制表符作为缺省的字段分隔符，因此如果要浏览字段间有空格，制表符的文本，不必指定这个选项，但如果要浏览诸如/etc/passwd文件，此文件各字段以冒号作为分隔符，则必须指明-F选项

```
echo "this is a test" | awk '{ print $0 }'  
## 输出为  
this is a test
```

- 显示/etc/passwd的第1列和第7列，用逗号分隔显示，所有行开始前添加列名start1，start7，最后一行添加，end1，end7

    ```
    awk -F ':' 'BEGIN {print "start1,start7"} {print $1 "," $7} END {print "end1,end7"}' /etc/passwd
    ```
- 统计/etc/passwd文件中，每行的行号，每行的列数，对应的完整行内容
  ```
  awk -F : '{ print NR "  " NF "  " $0 }' /etc/passwd  
  ```
    ```
    变量名 解释
    FILENAMEawk浏览的文件名
    FS设置输入字段分隔符，等价于命令行-F选项
    NF 浏览记录的字段个数
    NR 已读的记录数
    ``` 
- 多行合并成一行并逗号分割：`awk ' { printf ("%s ", $0)} END {printf ("\n") } '`
[参考](https://www.linuxprobe.com/linux-awk-clever.html)

## sed
sed是一种流编辑器，它一次处理一行内容。处理时，把当前处理的行存储在临时缓冲区中，称为“模式空间”（pattern space）

- 删除第2行 `nl /etc/passwd | sed '2d' `
- 要删除第 3 到最后一行 `nl /etc/passwd | sed '3,$d' `
- 截取文件多行，并合并到一行，空格分割 `sort sands_tables.txt | sed -n '13,36p' | tr "\n" " "`


# netstat 
- 查看服务端口占用 `netstat -anp`