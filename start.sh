port=8333
echo "StrictHostKeyChecking no" >~/.ssh/config

echo "先杀掉端口8333的程序"

pid=$(netstat -nlp | grep ":$port" | awk '{print $7}' | awk -F '[ / ]' '{print $1}')

if [ -n  "$pid" ]; then
    kill  -9  $pid;
    echo "pid="$pid
    echo "成功杀掉进程"$pid
else
    echo $port"没有启动，不存在该端口"
fi

chmod 755 quickweb
nohup ./quickweb &
echo "启动成功"
tail -f nohup.out
